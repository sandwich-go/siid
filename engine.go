package siid

import (
	"context"
	"errors"
	"fmt"
	"github.com/sandwich-go/boost/retry"
	"github.com/sandwich-go/boost/xsync"
	"github.com/sandwich-go/boost/z"
	"github.com/sandwich-go/logbus"
	"sort"
	"sync"
	"time"
)

const (
	driverFlagInit   int32 = iota // driver还未初始化
	driverFlagInited              // driver已初始化
	driverFlagClosed              // driver已关闭
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)
	// nowFunc returns the current time
	nowFunc                     = time.Now
	segmentFactor time.Duration = 2
	errDomainLost               = errors.New("lost domain")
)

const tag = "siid"

func w(msg string) string {
	return fmt.Sprintf("[%s]: %s", tag, msg)
}

func panicIfErr(err error) {
	if err != nil {
		panic(w(err.Error()))
	}
}

// Register makes a database driver available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panicIfErr(errors.New("register driver is nil"))
	}
	if _, dup := drivers[name]; dup {
		panicIfErr(fmt.Errorf("register called twice for driver %s", name))
	}
	drivers[name] = driver
}

// Drivers returns a sorted list of the names of the registered drivers.
func Drivers() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()
	list := make([]string, 0, len(drivers))
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

type builder struct {
	driver        Driver
	visitor       OptionsVisitor
	engineGetters *sync.Map
	flag          xsync.AtomicInt32
}

func New(driverName string, opts *Options) Builder {
	driversMu.RLock()
	driver, ok := drivers[driverName]
	driversMu.RUnlock()
	if !ok {
		panicIfErr(fmt.Errorf("unknown driver %q (forgotten import?)", driverName))
	}
	return NewWithDriver(driver, opts)
}

func NewWithDriver(driver Driver, opts *Options) Builder {
	b := &builder{driver: driver, engineGetters: &sync.Map{}, visitor: opts}
	return b
}

type engineGetter func() Engine

func (b *builder) Range(f func(domain string, engine Engine) bool) {
	b.engineGetters.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(engineGetter)())
	})
}

func (b *builder) getEngineGetterByDomain(domain string, offsetOnCreate uint64) engineGetter {
	if f, ok := b.engineGetters.Load(domain); ok {
		return f.(engineGetter)
	}
	var e Engine
	var wg sync.WaitGroup
	wg.Add(1)
	waitGetter := func() Engine {
		wg.Wait()
		return e
	}
	f, loaded := b.engineGetters.LoadOrStore(domain, engineGetter(waitGetter))
	if loaded {
		return f.(engineGetter)
	}
	e = &engine{builder: b, domain: domain, offsetOnCreate: offsetOnCreate}
	wg.Done()
	getter := func() Engine {
		return e
	}
	b.engineGetters.Store(domain, engineGetter(getter))
	return getter
}

func (b *builder) checkAvailableFlag() error {
	switch b.flag.Get() {
	case driverFlagInit:
		return ErrorDriverHasNotInited
	case driverFlagClosed:
		return ErrorDriverHasClosed
	}
	return nil
}

func (b *builder) Build(domain string) (Engine, error) {
	return b.BuildWithOffset(domain, 0)
}

func (b *builder) BuildWithOffset(domain string, offsetOnCreate uint64) (Engine, error) {
	if err := b.checkAvailableFlag(); err != nil {
		return nil, err
	}
	if offsetOnCreate == 0 {
		offsetOnCreate = b.visitor.GetOffsetWhenAutoCreateDomain()
	}
	return b.getEngineGetterByDomain(domain, offsetOnCreate)(), nil
}

func (b *builder) Prepare(ctx context.Context) error {
	if b.flag.CompareAndSwap(driverFlagInit, driverFlagInited) {
		return b.driver.Prepare(ctx)
	}
	if b.flag.Get() == driverFlagInited {
		return nil
	}
	return ErrorDriverHasClosed
}

func (b *builder) Destroy(ctx context.Context) error {
	if b.flag.CompareAndSwap(driverFlagInited, driverFlagClosed) {
		return b.driver.Destroy(ctx)
	}
	return b.checkAvailableFlag()
}

type engine struct {
	builder        *builder
	domain         string
	offsetOnCreate uint64

	// current
	n        uint64 // 当前值
	max      uint64
	quantum  uint64
	ts       z.MonoTimeDuration
	critical uint64

	// next
	nextN       uint64
	nextMax     uint64
	nextQuantum uint64

	nextMutex  sync.RWMutex
	renewMutex sync.RWMutex

	renewCount    xsync.AtomicUint64
	renewErrCount xsync.AtomicUint64
}

func (e *engine) Next() (uint64, error) {
	return e.NextN(1)
}

func (e *engine) MustNext() uint64 {
	i, err := e.Next()
	panicIfErr(err)
	return i
}

func (e *engine) MustNextN(n int) uint64 {
	i, err := e.NextN(n)
	panicIfErr(err)
	return i
}

func (e *engine) NextN(n int) (uint64, error) {
	if n <= 0 {
		n = 1
	}
	now := z.MonoOffset()
	// lock-free swap current and next ID bucket if we really really really really really need that
	e.nextMutex.Lock()
	var err error
	var id uint64
	// 需要优化,todo
	for i := 0; i < n; i++ {
		id, err = e.nextOne()
		if err != nil {
			break
		}
	}
	e.nextMutex.Unlock()
	e.nextReport(n, now, err)
	return id, err
}

func (e *engine) Stats() Stats {
	e.nextMutex.Lock()
	defer e.nextMutex.Unlock()
	return Stats{Current: e.n, Max: e.max, RenewCount: e.renewCount.Get(), RenewErrCount: e.renewErrCount.Get()}
}

func nextQuantum(lastQuantum uint64, segmentTime z.MonoTimeDuration, segmentDuration time.Duration, minQuantum, maxQuantum uint64) uint64 {
	nq := lastQuantum
	// 第一次renew使用初始值，不进行流控
	if segmentTime > 0 {
		duration := z.MonoSince(segmentTime)
		if duration < segmentDuration {
			// 流量增长期
			nq *= 2
		} else if duration < segmentDuration*segmentFactor {
			// 流量相对平稳
		} else {
			// 流量下降,申请号段减半
			nq /= 2
		}
	}
	if nq < minQuantum {
		nq = minQuantum
	}
	// ID库保护，防止高峰期停机导致的号段损失
	if nq > maxQuantum {
		nq = maxQuantum
	}
	return nq
}

func (e *engine) preRenew() (quantum uint64, begin z.MonoTimeDuration) {
	begin = z.MonoOffset()
	quantum = nextQuantum(e.quantum, e.ts,
		e.builder.visitor.GetSegmentDuration(),
		e.builder.visitor.GetMinQuantum(),
		e.builder.visitor.GetMaxQuantum(),
	)
	return
}

func (e *engine) postRenew(quantum uint64, begin z.MonoTimeDuration, err error) {
	e.renewReport(quantum, begin, err)
}

func (e *engine) renewWithUnlock() {
	defer e.renewMutex.Unlock()
	quantum, begin := e.preRenew()
	e.postRenew(quantum, begin, retry.Do(func(attempt uint) (errRetry error) {
		defer func() {
			if r := recover(); r != nil {
				errRetry = fmt.Errorf("panic %v", r)
				logbus.Error(w("renew panic"), logbus.Uint("attempt", attempt), logbus.String("domain", e.domain),
					logbus.Any("recover", r))
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), e.builder.visitor.GetRenewTimeout())
		defer cancel()
		c, err := e.builder.driver.Renew(ctx, e.domain, quantum, e.offsetOnCreate)
		if err != nil {
			errRetry = err
			return errRetry
		}
		e.nextN = c
		e.nextMax = c + quantum
		e.nextQuantum = quantum
		return nil
	},
		retry.WithLimit(e.builder.visitor.GetRenewRetry()),
		retry.WithDelayType(func(n uint, _ error, _ *retry.Options) time.Duration {
			return time.Duration(n) * e.builder.visitor.GetRenewRetryDelay()
		})))
}

func (e *engine) nextOne() (uint64, error) {
	if e.n == e.critical {
		e.renewMutex.Lock()
		if e.n == 0 {
			e.renewWithUnlock()
		} else {
			go e.renewWithUnlock()
		}
	}
	if e.max < e.n+1 {
		// wait until renew finished, swap to the next id bucket
		e.renewMutex.Lock()
		defer e.renewMutex.Unlock()
		if e.nextMax == 0 {
			logbus.Error(w("next failed"), logbus.String("reason", "id run out"), logbus.String("domain", e.domain))
			return 0, ErrIdRunOut
		}
		e.n = e.nextN
		e.max = e.nextMax
		e.quantum = e.nextQuantum
		e.useNewQuantumReport()
		// 记录号段正式投入使用的时间点
		e.ts = z.MonoOffset()
		// 计算renew临界值critical,renewCount不能为0,否则无法触发renew机制
		renewCount := (e.max - e.n) * uint64(e.builder.visitor.GetRenewPercent()) / 100
		if renewCount == 0 {
			renewCount = 1
		}
		e.critical = e.n + renewCount
		e.nextMax = 0
		e.nextN = 0
	}
	e.n++
	e.leftReport()
	if e.n > e.builder.visitor.GetLimitation() {
		logbus.Error(w("next failed"), logbus.String("reason", "max id"), logbus.String("domain", e.domain))
		return 0, ErrReachIdLimitation
	}
	return e.n, nil
}

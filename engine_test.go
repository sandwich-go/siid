package siid

import (
	"context"
	"github.com/sandwich-go/boost/z"
	. "github.com/smartystreets/goconvey/convey"
	"sync"
	"testing"
	"time"
)

const defaultOffsetWhenAutoCreateDomain = 30000000

func TestNextQuantum(t *testing.T) {
	Convey("next quantum", t, func() {
		var initQuantum, minQuantum, maxQuantum uint64 = 20, 10, 40
		var segmentDuration = 100 * time.Millisecond
		var quantum = nextQuantum(initQuantum, 0, segmentDuration, minQuantum, maxQuantum)
		So(quantum, ShouldEqual, initQuantum)
		var segmentTime = z.MonoOffset()
		quantum = nextQuantum(quantum, segmentTime, segmentDuration, minQuantum, maxQuantum)
		So(quantum, ShouldEqual, maxQuantum)
		time.Sleep(segmentDuration * segmentFactor)
		quantum = nextQuantum(quantum, segmentTime, segmentDuration, minQuantum, maxQuantum)
		So(quantum, ShouldEqual, maxQuantum/2)
	})
}

func TestSIID(t *testing.T) {
	Convey("siid", t, func() {
		var driverName, domain = "dummy", "test"
		var quantum uint64 = 1000
		Register(driverName, getDummyDriver())

		So(len(Drivers()), ShouldEqual, 1)
		So(Drivers()[0], ShouldEqual, driverName)

		b := New(driverName,
			WithOffsetWhenAutoCreateDomain(defaultOffsetWhenAutoCreateDomain),
			WithInitialQuantum(quantum),
			WithDevelopment(false),
		)
		e, err := b.Build(domain)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrorDriverHasNotInited)
		So(e, ShouldBeNil)

		err = b.Prepare(context.Background())
		So(err, ShouldBeNil)
		err = b.Prepare(context.Background())
		So(err, ShouldBeNil)

		e, err = b.Build(domain)
		So(err, ShouldBeNil)
		So(e, ShouldNotBeNil)

		s := e.Stats()
		So(s.Max, ShouldBeZeroValue)
		So(s.Current, ShouldBeZeroValue)
		So(s.RenewErrCount, ShouldBeZeroValue)
		So(s.RenewCount, ShouldBeZeroValue)

		var id uint64
		for i := 0; i < int(quantum)/2; i++ {
			id, err = e.Next()
			So(err, ShouldBeNil)
			So(id, ShouldEqual, defaultOffsetWhenAutoCreateDomain+i+1)
		}
		s = e.Stats()
		So(s.Current, ShouldEqual, defaultOffsetWhenAutoCreateDomain+int(quantum)/2)
		So(s.RenewErrCount, ShouldBeZeroValue)
		So(s.RenewCount, ShouldNotBeZeroValue)

		err = b.Destroy(context.Background())
		So(err, ShouldBeNil)
		err = b.Destroy(context.Background())
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrorDriverHasClosed)

		e, err = b.Build(domain)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrorDriverHasClosed)
	})
}

var (
	once                             sync.Once
	dummyDriverName, mysqlDriverName = "dummy", "mysql"
)

func initBenchmark() {
	once.Do(func() {
		Register(dummyDriverName, getDummyDriver())
		Register(mysqlDriverName, getMysqlDriver(mysqlAddress))
	})
}

func getBenchmarkEngine(b *testing.B, driverName string) Engine {
	initBenchmark()
	var domain = "dummy"
	var quantum uint64 = 1000
	bd := New(driverName,
		WithOffsetWhenAutoCreateDomain(defaultOffsetWhenAutoCreateDomain),
		WithInitialQuantum(quantum),
		WithDevelopment(false),
		WithEnableSlow(false),
	)
	err := bd.Prepare(context.Background())
	if err != nil {
		b.Fatal(err)
	}
	e, err0 := bd.Build(domain)
	if err0 != nil {
		b.Fatal(err0)
	}
	return e
}

func benchmarkSIID(b *testing.B, driverName string) {
	e := getBenchmarkEngine(b, driverName)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.Next()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSIID_Dummy(b *testing.B) {
	benchmarkSIID(b, dummyDriverName)
}

func BenchmarkSIID_Mysql(b *testing.B) {
	benchmarkSIID(b, mysqlDriverName)
}

func BenchmarkParallelSIID_Mysql(b *testing.B) {
	e := getBenchmarkEngine(b, mysqlDriverName)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := e.Next()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

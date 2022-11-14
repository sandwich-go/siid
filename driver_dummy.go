package siid

import (
	"context"
	"sync"
)

type dummyDriver struct {
	mx sync.RWMutex
	mm map[string]uint64
}

func newDummyDriver() Driver {
	return &dummyDriver{mm: make(map[string]uint64)}
}
func (d *dummyDriver) Prepare(_ context.Context) error { return nil }
func (d *dummyDriver) Destroy(_ context.Context) error { return nil }
func (d *dummyDriver) Renew(_ context.Context, domain string, quantum, offset uint64) (uint64, error) {
	d.mx.Lock()
	defer d.mx.Unlock()
	val, ok := d.mm[domain]
	if !ok {
		val = offset
		d.mm[domain] = val
	}
	d.mm[domain] += quantum
	return val, nil
}

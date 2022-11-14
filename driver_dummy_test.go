package siid

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func getDummyDriver() *dummyDriver {
	driver := newDummyDriver()
	if err := driver.Prepare(context.Background()); err != nil {
		panic(err)
	}
	return driver.(*dummyDriver)
}

func Test_DummyDriverOffset(t *testing.T) {
	driver := getDummyDriver()
	t.Cleanup(func() {
		if err0 := driver.Destroy(context.Background()); err0 != nil {
			t.Error(err0)
		}
	})
	Convey("dummy driver offset", t, func() {
		current, err := driver.Renew(context.Background(), fmt.Sprintf("test_ts_%d", nowFunc().Unix()), 1000, defaultOffsetWhenAutoCreateDomain)
		So(err, ShouldBeNil)
		So(current, ShouldEqual, defaultOffsetWhenAutoCreateDomain)
	})
}

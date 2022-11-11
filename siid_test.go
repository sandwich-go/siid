package siid

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestNextQuantum(t *testing.T) {
	Convey("next quantum", t, func() {
		var initQuantum, minQuantum, maxQuantum uint64 = 20, 10, 40
		var segmentDuration = 100 * time.Millisecond
		var quantum = nextQuantum(initQuantum, time.Time{}, segmentDuration, minQuantum, maxQuantum)
		So(quantum, ShouldEqual, initQuantum)
		var segmentTime = nowFunc()
		quantum = nextQuantum(quantum, segmentTime, segmentDuration, minQuantum, maxQuantum)
		So(quantum, ShouldEqual, maxQuantum)
		time.Sleep(segmentDuration * segmentFactor)
		quantum = nextQuantum(quantum, segmentTime, segmentDuration, minQuantum, maxQuantum)
		So(quantum, ShouldEqual, maxQuantum/2)
	})
}

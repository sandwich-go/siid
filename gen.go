package siid

import (
	"math"
	"time"
)

type (
	// RenewRetryDelayFunc renew失败重试函数
	RenewRetryDelayFunc func(attempt int) time.Duration
)

var (
	defaultRenewRetryDelayFunc = func(attempt int) time.Duration {
		return time.Millisecond * time.Duration(10*attempt)
	}
)

//go:generate optiongen --option_with_struct_name=false --new_func=NewConfig --xconf=true --empty_composite_nil=true --usage_tag_name=usage
func OptionsOptionDeclareWithDefault() interface{} {
	return map[string]interface{}{
		"Limitation":                 uint64(math.MaxUint64),                          // @MethodComment(id最大限制，超过该值则会报ErrReachIdLimitation错误)
		"OffsetWhenAutoCreateDomain": uint64(30000000),                                // @MethodComment(当新建新的domain时，偏移多少开始自增，即预留值)
		"RenewPercent":               20,                                              // @MethodComment(renew百分比，当id达到百分比值时，会去server端或db拿新的id段)
		"RenewRetryDelay":            RenewRetryDelayFunc(defaultRenewRetryDelayFunc), // @MethodComment(renew失败重试函数，默认10ms重试一次)
		"RenewTimeout":               time.Duration(5 * time.Second),                  // @MethodComment(renew超时)
		"RenewRetry":                 99,                                              // @MethodComment(renew重试次数)
		"SegmentDuration":            time.Duration(900 * time.Second),                // @MethodComment(设定segment长度，renew号段尺寸调节的目的是使号段消耗稳定趋于SegmentDuration内。降低SegmentDuration，可以更迅速使缓存的号段达到设定的最大数值以提高吞吐能力)
		"MinQuantum":                 uint64(30),                                      // @MethodComment(根据renew请求频率自动伸缩的请求id缓存段，最小段长)
		"MaxQuantum":                 uint64(3000),                                    // @MethodComment(最大段长)
		"InitialQuantum":             uint64(30),                                      // @MethodComment(初始化段长)
		"EnableSlow":                 true,                                            // @MethodComment(是否开启慢日志)
		"SlowQuery":                  time.Duration(30 * time.Millisecond),            // @MethodComment(慢日志最小时长，大于该时长将输出日志)
		"EnableTimeSummary":          false,                                           // @MethodComment(是否开启Next/MustNext接口的time监控，否则为统计监控)
	}
}

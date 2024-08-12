package siid

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sandwich-go/boost/z"
	"github.com/sandwich-go/logbus"
	"github.com/sandwich-go/logbus/monitor"
	"sync"
	"time"
)

const (
	renewTimingName  = "siid_renew_time"
	renewCountName   = "siid_renew"
	nextTimingName   = "siid_next_time"
	nextCountName    = "siid_next"
	quantumGaugeName = "siid_quantum"
	maxGaugeName     = "siid_max"
	nleftGaugeName   = "siid_n_left"
)

var (
	registerMonitorOnce sync.Once

	renewTimingLatency = newSummaryVec(renewTimingName, "domain", "status")
	renewCounter       = newCounterVec(renewCountName, "domain", "status")
	nextTimingLatency  = newSummaryVec(nextTimingName, "domain")
	nextCounter        = newCounterVec(nextCountName, "domain")
	quantumGauge       = newGaugeVec(quantumGaugeName, "domain")
	maxGauge           = newGaugeVec(maxGaugeName, "domain")
	nleftGauge         = newGaugeVec(nleftGaugeName, "domain")
)

func newGaugeVec(name string, label ...string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
		}, label)
}

func newSummaryVec(name string, label ...string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: name,
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.95: 0.02,
				0.99: 0.001,
				1.0:  0,
			},
			MaxAge: time.Minute,
		}, label)
}

func newCounterVec(name string, label ...string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
		}, label)
}

func init() {
	registerMonitorOnce.Do(func() {
		monitor.RegisterCollector(renewTimingLatency)
		monitor.RegisterCollector(renewCounter)
		monitor.RegisterCollector(nextTimingLatency)
		monitor.RegisterCollector(nextCounter)
		monitor.RegisterCollector(quantumGauge)
		monitor.RegisterCollector(maxGauge)
		monitor.RegisterCollector(nleftGauge)
	})
}

func getRenewStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "ok"
}

func (e *engine) renewReport(currQuantum uint64, renewBegin z.MonoTimeDuration, err error) {
	if !e.builder.visitor.GetEnableMonitor() {
		return
	}

	if err != nil {
		_ = e.renewErrCount.Add(1)
		logbus.Error(w("renew error"), logbus.ErrorField(err), logbus.String("domain", e.domain))
	} else {
		_ = e.renewCount.Add(1)
		if e.builder.visitor.GetDevelopment() {
			logbus.Debug(w("renew ok"), logbus.Uint64("nextN", e.nextN),
				logbus.Uint64("quantum", currQuantum), logbus.Uint64("nextMax", e.nextMax), logbus.String("domain", e.domain))
		}
	}
	if e.builder.visitor.GetEnableTimeSummary() {
		renewTimingLatency.WithLabelValues(e.domain, getRenewStatus(err)).Observe(z.MonoSince(renewBegin).Seconds())
	} else {
		renewCounter.WithLabelValues(e.domain, getRenewStatus(err)).Inc()
	}
}

func (e *engine) nextReport(n int, nextBegin z.MonoTimeDuration, _ error) {
	if !e.builder.visitor.GetEnableMonitor() {
		return
	}

	cost := z.MonoSince(nextBegin)
	if e.builder.visitor.GetEnableTimeSummary() {
		nextTimingLatency.WithLabelValues(e.domain).Observe(cost.Seconds())
	} else {
		nextCounter.WithLabelValues(e.domain).Inc()
	}
	if e.builder.visitor.GetEnableSlow() && cost >= e.builder.visitor.GetSlowQuery() {
		logbus.Warn(w("next slow query"), logbus.Duration("cost", cost), logbus.String("domain", e.domain), logbus.Int("count", n))
	}
}

func (e *engine) useNewQuantumReport() {
	if !e.builder.visitor.GetEnableMonitor() {
		return
	}

	quantumGauge.WithLabelValues(e.domain).Set(float64(e.quantum))
	maxGauge.WithLabelValues(e.domain).Set(float64(e.max))
}

func (e *engine) leftReport() {
	if !e.builder.visitor.GetEnableMonitor() {
		return
	}

	nleftGauge.WithLabelValues(e.domain).Set(float64(e.builder.visitor.GetLimitation() - e.n))
}

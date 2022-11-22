package siid

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sandwich-go/boost/z"
	"github.com/sandwich-go/logbus"
	"github.com/sandwich-go/logbus/monitor"
)

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
		_ = monitor.Timing("siid_renew_time", z.MonoSince(renewBegin), prometheus.Labels{"domain": e.domain, "status": getRenewStatus(err)})
	} else {
		_ = monitor.Count("siid_renew", 1, prometheus.Labels{"domain": e.domain, "status": getRenewStatus(err)})
	}
}

func (e *engine) nextReport(n int, nextBegin z.MonoTimeDuration, _ error) {
	if !e.builder.visitor.GetEnableMonitor() {
		return
	}
	cost := z.MonoSince(nextBegin)
	if e.builder.visitor.GetEnableTimeSummary() {
		_ = monitor.Timing("siid_next_time", cost, prometheus.Labels{"domain": e.domain})
	} else {
		_ = monitor.Count("siid_next", int64(n), prometheus.Labels{"domain": e.domain})
	}
	if e.builder.visitor.GetEnableSlow() && cost >= e.builder.visitor.GetSlowQuery() {
		logbus.Warn(w("next slow query"), logbus.Duration("cost", cost), logbus.String("domain", e.domain), logbus.Int("count", n))
	}
}

func (e *engine) useNewQuantumReport() {
	if !e.builder.visitor.GetEnableMonitor() {
		return
	}
	pl := prometheus.Labels{"domain": e.domain}
	_ = monitor.Gauge("siid_quantum", float64(e.quantum), pl)
	_ = monitor.Gauge("siid_max", float64(e.max), pl)
}

func (e *engine) leftReport() {
	if !e.builder.visitor.GetEnableMonitor() {
		return
	}
	_ = monitor.Gauge("siid_n_left", float64(e.builder.visitor.GetLimitation()-e.n), prometheus.Labels{"domain": e.domain})
}

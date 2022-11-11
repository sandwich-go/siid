package siid

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sandwich-go/logbus/glog"
	"github.com/sandwich-go/logbus/monitor"
	"time"
)

func getRenewStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "ok"
}

func (e *engine) renewReport(currQuantum uint64, renewBegin time.Time, err error) {
	if err != nil {
		_ = e.renewErrCount.Add(1)
		glog.Error("renew error", glog.Err(err), glog.String("domain", e.domain))
	} else {
		_ = e.renewCount.Add(1)
		glog.Debug("renew ok", glog.Uint64("nextN", e.nextN),
			glog.Uint64("quantum", currQuantum), glog.Uint64("nextMax", e.nextMax), glog.String("domain", e.domain))
	}
	if e.builder.visitor.GetEnableTimeSummary() {
		_ = monitor.Timing("siid_renew_time", time.Since(renewBegin), prometheus.Labels{"domain": e.domain, "status": getRenewStatus(err)})
	} else {
		_ = monitor.Count("siid_renew", 1, prometheus.Labels{"domain": e.domain, "status": getRenewStatus(err)})
	}
}

func (e *engine) nextReport(n int, nextBegin time.Time, _ error) {
	cost := time.Since(nextBegin)
	if e.builder.visitor.GetEnableTimeSummary() {
		_ = monitor.Timing("siid_next_time", cost, prometheus.Labels{"domain": e.domain})
	} else {
		_ = monitor.Count("siid_next", int64(n), prometheus.Labels{"domain": e.domain})
	}
	if e.builder.visitor.GetEnableSlow() && cost >= e.builder.visitor.GetSlowQuery() {
		glog.Warn(w("next slow query"), glog.Duration("cost", cost), glog.String("domain", e.domain))
	}
}

func (e *engine) useNewQuantumReport() {
	pl := prometheus.Labels{"domain": e.domain}
	_ = monitor.Gauge("siid_quantum", float64(e.quantum), pl)
	_ = monitor.Gauge("siid_max", float64(e.max), pl)
}

func (e *engine) leftReport() {
	_ = monitor.Gauge("siid_n_left", float64(e.builder.visitor.GetLimitation()-e.n), prometheus.Labels{"domain": e.domain})
}

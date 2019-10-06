package metrictest

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// TestingT is an interface wrapper around *testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
}

// AssertCounter asserts the value of a prometheus.Counter
//
//    counter.Add(10)
//    metrictest.AssertCounter(t, 10, counter)
func AssertCounter(t TestingT, expected float64, counter prometheus.Counter) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	actual := testutil.ToFloat64(counter)
	if actual != expected {
		t.Errorf(fmt.Sprintf(`Not equal:
expected: %#v
actual  : %#v`, expected, actual))
	}
}

// AssertCounterVec asserts the value of a single counter in a prometheus.CounterVec
//
//   counterVec.WithLabelValues("a-label", "another-label").Inc(10)
//   metrictest.AssertCounterVec(t, 10, counterVec, "a-label", "another-label")
//
// Asserting that the value of a counter is 0 is equivalent to asserting that the counter does not exist.
func AssertCounterVec(t TestingT, expected float64, counterVec *prometheus.CounterVec, labels ...string) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	m, err := counterVec.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	assertVec(t, expected, m, counterVec)
}

// AssertGauge asserts the value of a prometheus.Gauge
//
//    gauge.Set(10)
//    metrictest.AssertGauge(t, 10, gauge)
func AssertGauge(t TestingT, expected float64, gauge prometheus.Gauge) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	actual := testutil.ToFloat64(gauge)
	if actual != expected {
		t.Errorf(fmt.Sprintf(`Not equal:
expected: %#v
actual  : %#v`, expected, actual))
	}
}

// AssertGaugeVec asserts the value of a single gauge in a prometheus.GaugeVec
//
//   gaugeVec.WithLabelValues("a-label", "another-label").Set(10)
//   metrictest.AssertGaugeVec(t, 10, gaugeVec, "a-label", "another-label")
//
// Asserting that the value of a gauge is 0 is equivalent to asserting that the gauge does not exist.
func AssertGaugeVec(t TestingT, expected float64, gaugeVec *prometheus.GaugeVec, labels ...string) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	m, err := gaugeVec.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	assertVec(t, expected, m, gaugeVec)
}

func assertVec(t TestingT, expected float64, metric prometheus.Collector, vec prometheus.Collector) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	actual := testutil.ToFloat64(metric)
	if actual == expected {
		return
	}

	expvar, err := ToExpVar(vec)
	if err != nil {
		expvar = err.Error()
	}
	t.Errorf(fmt.Sprintf(`Not equal:
expected: %#v
actual  : %#v

metric:
%v`, expected, actual, expvar))
}

// AssertHistogramSamples asserts the count and sum of all samples captured by the prometheus.Histogram
//
//    histogram.Observe(10)
//    histogram.Observe(20)
//    metrictest.AssertHistogramSamples(t, 2, 30, histogram)
func AssertHistogramSamples(t TestingT, count int, sum float64, histogram prometheus.Histogram) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	metric := collectOne(histogram)

	actualCount := metric.Histogram.GetSampleCount()
	actualSum := metric.Histogram.GetSampleSum()

	if actualCount != uint64(count) {
		t.Errorf(fmt.Sprintf(`Sample count not equal:
expected: %d
actual  : %d`, count, actualCount))
		return
	}
	if actualSum != sum {
		t.Errorf(fmt.Sprintf(`Sample sum not equal:
expected: %g
actual  : %g`, sum, actualSum))
	}
}

// AssertHistogramVecSamples asserts the count and sum of all samples captured by a single histogram
// contained by the prometheus.HistogramVec
//
//    histogramVec.WithLabelValues("a-label", "another-label").Observe(10)
//    histogramVec.WithLabelValues("a-label", "another-label").Observe(20)
//    metrictest.AssertHistogramSamples(t, 2, 30, histogram, "a-label", "another-label")
func AssertHistogramVecSamples(t TestingT, count int, sum float64, histogramVec *prometheus.HistogramVec, labels ...string) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	m, err := histogramVec.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	histogram := m.(prometheus.Histogram)
	metric := collectOne(histogram)

	actualCount := metric.Histogram.GetSampleCount()
	actualSum := metric.Histogram.GetSampleSum()

	assertHistogramVec(t, count, actualCount, sum, actualSum, histogramVec)
}

// AssertSummarySamples asserts the count and sum of all samples captured by the prometheus.Summary
//
//    summary.Observe(10)
//    summary.Observe(20)
//    metrictest.AssertSummarySamples(t, 2, 30, summary)
func AssertSummarySamples(t TestingT, count int, sum float64, summary prometheus.Histogram) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	// Histogram and summary have the same interface
	metric := collectOne(summary)

	actualCount := metric.Summary.GetSampleCount()
	actualSum := metric.Summary.GetSampleSum()

	if actualCount != uint64(count) {
		t.Errorf(fmt.Sprintf(`Sample count not equal:
expected: %d
actual  : %d`, count, actualCount))
		return
	}
	if actualSum != sum {
		t.Errorf(fmt.Sprintf(`Sample sum not equal:
expected: %g
actual  : %g`, sum, actualSum))
	}
}

// AssertSummaryVecSamples asserts the count and sum of all samples captured by a single summary
// contained by the prometheus.SummaryVec
//
//    summaryVec.WithLabelValues("a-label", "another-label").Observe(10)
//    summaryVec.WithLabelValues("a-label", "another-label").Observe(20)
//    metrictest.AssertSummarySamples(t, 2, 30, summary, "a-label", "another-label")
func AssertSummaryVecSamples(t TestingT, count int, sum float64, summaryVec *prometheus.SummaryVec, labels ...string) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	m, err := summaryVec.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	summary := m.(prometheus.Summary)
	metric := collectOne(summary)

	actualCount := metric.Summary.GetSampleCount()
	actualSum := metric.Summary.GetSampleSum()

	assertHistogramVec(t, count, actualCount, sum, actualSum, summaryVec)
}

func assertHistogramVec(t TestingT, count int, actualCount uint64, sum float64, actualSum float64, c prometheus.Collector) {
	if actualCount == uint64(count) && actualSum == sum {
		return
	}
	expvar, err := ToExpVar(c)
	if err != nil {
		expvar = err.Error()
	}
	expvar = filterSampleSumAndCount(expvar)
	if actualCount != uint64(count) {
		t.Errorf(fmt.Sprintf(`Sample count not equal:
expected: %d
actual  : %d

metric:
%v`, count, actualCount, expvar))
	}
	if actualSum != sum {
		t.Errorf(fmt.Sprintf(`Sample sum not equal:
expected: %g
actual  : %g

metric:
%v`, sum, actualSum, expvar))
	}
}

func collectOne(c prometheus.Collector) *dto.Metric {
	var (
		m     prometheus.Metric
		count int
		mChan = make(chan prometheus.Metric)
		done  = make(chan struct{})
	)

	go func() {
		for m = range mChan {
			count++
		}
		close(done)
	}()

	c.Collect(mChan)
	close(mChan)
	<-done

	if count != 1 {
		panic(fmt.Sprintf("expected to collect 1 metric, but got %d", count))
	}

	pb := &dto.Metric{}
	_ = m.Write(pb)

	return pb
}

// ToExpVar collects metrics from the collector and prints them in ExpVar format.
func ToExpVar(c prometheus.Collector) (string, error) {
	reg := prometheus.NewPedanticRegistry()
	if err := reg.Register(c); err != nil {
		return "", fmt.Errorf("registering collector failed: %s", err)
	}
	metricFamilies, err := reg.Gather()
	if err != nil {
		return "", err
	}
	buf := &strings.Builder{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)
	for _, mf := range metricFamilies {
		if err = enc.Encode(mf); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

var sumOrCount = regexp.MustCompile("_(sum|count)[{ ]")

func filterSampleSumAndCount(expvar string) string {
	parts := strings.Split(expvar, "\n")
	var newParts []string
	for _, p := range parts {
		if sumOrCount.MatchString(p) {
			newParts = append(newParts, p)
		}
	}
	return strings.Join(newParts, "\n")
}

type tHelper interface {
	Helper()
}

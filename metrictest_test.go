package metrictest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	counter    = prometheus.NewCounter(prometheus.CounterOpts{Name: "test_counter"})
	counterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test_counter_vec",
	}, []string{"hero", "villain"})
	gauge    = prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_gauge"})
	gaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "test_gauge_vec",
	}, []string{"hero", "villain"})
	histogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "test_histogram",
	})
	histogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "test_histogram",
	}, []string{"hero", "villain"})
	summary = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "test_summary",
	})
	summaryVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "test_summary",
	}, []string{"hero", "villain"})
)

func reset() {
	counterVec.Reset()
}

type recordingT []string

func (t *recordingT) Errorf(format string, args ...interface{}) {
	*t = append(*t, fmt.Sprintf(format, args...))
}

func TestAssertCounter(t *testing.T) {
	reset()

	recT := recordingT{}
	counter.Add(10)
	AssertCounter(&recT, 10, counter)

	require.Len(t, recT, 0)

	AssertCounter(&recT, 1, counter)
	require.Len(t, recT, 1)
	requireStringEqual(t, recT[0], `
Not equal:
expected: 1
actual  : 10
`)
}

func TestAssertCounterVec(t *testing.T) {
	reset()

	counterVec.WithLabelValues("batman", "joker").Add(2)
	counterVec.WithLabelValues("superman", "lex luthor").Add(100)

	recT := recordingT{}
	AssertCounterVec(&recT, 2, counterVec, "batman", "joker")
	AssertCounterVec(&recT, 100, counterVec, "superman", "lex luthor")

	require.Len(t, recT, 0)

	AssertCounterVec(&recT, 1, counterVec, "batman", "joker")

	require.Len(t, recT, 1)
	requireStringEqual(t, recT[0], `
Not equal:
expected: 1
actual  : 2

metric:
# HELP test_counter_vec 
# TYPE test_counter_vec counter
test_counter_vec{hero="batman",villain="joker"} 2
test_counter_vec{hero="superman",villain="lex luthor"} 100
`)
}

func TestAssertCounterVecZero(t *testing.T) {
	reset()

	counterVec.WithLabelValues("batman", "joker").Add(2)

	recT := recordingT{}
	AssertCounterVec(&recT, 0, counterVec, "foo", "bar")

	require.Len(t, recT, 0)
}

func TestAssertGauge(t *testing.T) {
	reset()

	recT := recordingT{}
	gauge.Set(10)
	AssertGauge(&recT, 10, gauge)

	require.Len(t, recT, 0)

	AssertGauge(&recT, 1, gauge)
	require.Len(t, recT, 1)
	requireStringEqual(t, recT[0], `
Not equal:
expected: 1
actual  : 10
`)
}

func TestAssertGaugeVec(t *testing.T) {
	reset()

	gaugeVec.WithLabelValues("batman", "joker").Add(2)
	gaugeVec.WithLabelValues("superman", "lex luthor").Add(100)

	recT := recordingT{}
	AssertGaugeVec(&recT, 2, gaugeVec, "batman", "joker")
	AssertGaugeVec(&recT, 100, gaugeVec, "superman", "lex luthor")

	require.Len(t, recT, 0)

	AssertGaugeVec(&recT, 1, gaugeVec, "batman", "joker")

	require.Len(t, recT, 1)
	requireStringEqual(t, recT[0], `
Not equal:
expected: 1
actual  : 2

metric:
# HELP test_gauge_vec 
# TYPE test_gauge_vec gauge
test_gauge_vec{hero="batman",villain="joker"} 2
test_gauge_vec{hero="superman",villain="lex luthor"} 100
`)
}

func TestAssertGaugeVecZero(t *testing.T) {
	reset()

	gaugeVec.WithLabelValues("batman", "joker").Add(2)

	recT := recordingT{}
	AssertGaugeVec(&recT, 0, gaugeVec, "foo", "bar")

	require.Len(t, recT, 0)
}

func TestAssertHistogramSamples(t *testing.T) {
	reset()

	histogram.Observe(10)
	histogram.Observe(20)

	recT := recordingT{}
	AssertHistogramSamples(&recT, 2, 30, histogram)

	require.Len(t, recT, 0)

	AssertHistogramSamples(&recT, 1, 30, histogram)

	require.Len(t, recT, 1)
	requireStringEqual(t, recT[0], `
Sample count not equal:
expected: 1
actual  : 2
`)

	recT = recordingT{}
	AssertHistogramSamples(&recT, 2, 42, histogram)
	requireStringEqual(t, recT[0], `
Sample sum not equal:
expected: 42
actual  : 30
`)
}

func TestAssertHistogramVecSamples(t *testing.T) {
	reset()

	histogramVec.WithLabelValues("batman", "joker").Observe(10)
	histogramVec.WithLabelValues("batman", "joker").Observe(20)
	histogramVec.WithLabelValues("superman", "lex luthor").Observe(10)

	recT := recordingT{}
	AssertHistogramVecSamples(&recT, 2, 30, histogramVec, "batman", "joker")

	require.Len(t, recT, 0)

	AssertHistogramVecSamples(&recT, 1, 30, histogramVec, "batman", "joker")

	require.Len(t, recT, 1)
	requireStringEqual(t, recT[0], `
Sample count not equal:
expected: 1
actual  : 2

metric:
test_histogram_sum{hero="batman",villain="joker"} 30
test_histogram_count{hero="batman",villain="joker"} 2
test_histogram_sum{hero="superman",villain="lex luthor"} 10
test_histogram_count{hero="superman",villain="lex luthor"} 1`)

	recT = recordingT{}
	AssertHistogramVecSamples(&recT, 2, 42, histogramVec, "batman", "joker")
	requireStringEqual(t, recT[0], `
Sample sum not equal:
expected: 42
actual  : 30

metric:
test_histogram_sum{hero="batman",villain="joker"} 30
test_histogram_count{hero="batman",villain="joker"} 2
test_histogram_sum{hero="superman",villain="lex luthor"} 10
test_histogram_count{hero="superman",villain="lex luthor"} 1
`)
}

func TestAssertHistogramVecSamplesZero(t *testing.T) {
	reset()

	histogramVec.WithLabelValues("batman", "joker").Observe(10)

	recT := recordingT{}
	AssertHistogramVecSamples(&recT, 0, 0, histogramVec, "foo", "bar")

	require.Len(t, recT, 0)
}

func TestAssertSummarySamples(t *testing.T) {
	reset()

	summary.Observe(10)
	summary.Observe(20)

	recT := recordingT{}
	AssertSummarySamples(&recT, 2, 30, summary)

	require.Len(t, recT, 0)

	AssertSummarySamples(&recT, 1, 30, summary)

	require.Len(t, recT, 1)
	requireStringEqual(t, recT[0], `
Sample count not equal:
expected: 1
actual  : 2
`)

	recT = recordingT{}
	AssertHistogramSamples(&recT, 2, 42, histogram)
	requireStringEqual(t, recT[0], `
Sample sum not equal:
expected: 42
actual  : 30
`)
}

func TestAssertSummaryVecSamples(t *testing.T) {
	reset()

	summaryVec.WithLabelValues("batman", "joker").Observe(10)
	summaryVec.WithLabelValues("batman", "joker").Observe(20)
	summaryVec.WithLabelValues("superman", "lex luthor").Observe(10)

	recT := recordingT{}
	AssertSummaryVecSamples(&recT, 2, 30, summaryVec, "batman", "joker")

	require.Len(t, recT, 0)

	AssertSummaryVecSamples(&recT, 1, 30, summaryVec, "batman", "joker")

	require.Len(t, recT, 1)
	requireStringEqual(t, recT[0], `
Sample count not equal:
expected: 1
actual  : 2

metric:
test_summary_sum{hero="batman",villain="joker"} 30
test_summary_count{hero="batman",villain="joker"} 2
test_summary_sum{hero="superman",villain="lex luthor"} 10
test_summary_count{hero="superman",villain="lex luthor"} 1
`)

	recT = recordingT{}
	AssertSummaryVecSamples(&recT, 2, 42, summaryVec, "batman", "joker")
	requireStringEqual(t, recT[0], `
Sample sum not equal:
expected: 42
actual  : 30

metric:
test_summary_sum{hero="batman",villain="joker"} 30
test_summary_count{hero="batman",villain="joker"} 2
test_summary_sum{hero="superman",villain="lex luthor"} 10
test_summary_count{hero="superman",villain="lex luthor"} 1
`)
}

func TestAssertSummaryVecSamplesZero(t *testing.T) {
	reset()

	summaryVec.WithLabelValues("batman", "joker").Observe(10)

	recT := recordingT{}
	AssertSummaryVecSamples(&recT, 0, 0, summaryVec, "foo", "bar")

	require.Len(t, recT, 0)
}

func TestToExpVar(t *testing.T) {
	reset()

	counterVec.WithLabelValues("batman", "joker").Add(2)
	counterVec.WithLabelValues("superman", "lex luthor").Add(100)
	out, err := ToExpVar(counterVec)
	require.NoError(t, err)
	assert.Equal(t, `# HELP test_counter_vec 
# TYPE test_counter_vec counter
test_counter_vec{hero="batman",villain="joker"} 2
test_counter_vec{hero="superman",villain="lex luthor"} 100
`, out)
}

func requireStringEqual(t *testing.T, subject, contains string) {
	require.Equal(t, strings.TrimSpace(subject), strings.TrimSpace(contains))
}

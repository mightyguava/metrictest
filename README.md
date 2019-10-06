# metrictest

[![go-doc](https://godoc.org/github.com/mightyguava/metrictest?status.svg)](https://godoc.org/github.com/mightyguava/metrictest)

A small metric testing library for [Prometheus Go client](https://github.com/prometheus/client_golang).

## Counter/Gauge

The current value of Counters and Gauges can be asserted.

```go
counter := prometheus.NewCounter(prometheus.CounterOpts{...})
counter.Add(10)

metrictest.AssertCounter(t, 10, counter)
```

## CounterVec/GaugeVec

The current value of a single Counter in a CounterVec can be asserted. Asserting a value of 0 is the same as asserting that the Counter is not contained in the CounterVec. The interface for GaugeVec is identical to that of CounterVec.

```go
counter := prometheus.NewCounterVec(prometheus.CounterOpts{...}, []string{"hero", "villain"})
counter.WithLabelValues("batman", "joker").Add(10)

metrictest.AssertCounterVec(t, 10, counter, "batman", "joker")
```

## Histogram/Summary

Asserts for Histogram and Summary work a little differently from Gauge and Counter. Instead of asserting the value of each bucket/percentile, we assert the count and sum of all samples observed by the Histogram or Summary. As before, asserting a value of 0 is the same as asserting that no samples have been observed.

```go
histogram := prometheus.NewHistogram(prometheus.HistogramOpts{...})
histogram.Observe(10)
histogram.Observe(20)

metrictest.AssertHistogramSamples(t, 2, 30, histogram)
```

## HistogramVec/SummaryVec

As with Histogram, we can assert the sample count and sum of a single Histogram in the HistogramVec. Asserting a value of 0 for count is the same as asserting that the Histogram is not contained in the HistogramVec. The interface for SummaryVec is identical to that of HistogramVec.

```go
histogramVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{...}, []string{"hero", "villain"})
histogramVec.WithLabelValues("a-label", "another-label").Observe(10)
histogramVec.WithLabelValues("a-label", "another-label").Observe(20)

metrictest.AssertHistogramVecSamples(t, 2, 30, histogram, "a-label", "another-label")
```

package n9epush_test

import (
	"bufio"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/n9epush"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"strconv"
)

var (
	defaultPushGateway string = "http://localhost:2080/v1/push"
	defaultHttpPort uint64 = 8888
)

func Example_DefaultMode() {
	metrics := n9epush.NewMetrics("rd", "metrics", "test", nil)
	metrics.StartPushLoop("26", 1, "http://n9e-v4.performance-test.bybit.com/api/transfer/push")
	bucket := []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10}
	defaultMode := n9epush.NewDefaultMode(bucket, true, []string{"400", "500"})
	defaultMode.MakeCode("200")
	defaultMode.MakeCode("400")
	defaultMode.MakeCode("500")
	defaultMode.MakeDuration(3, "code")
	defaultMode.MakeDuration(2, "")
	bufio.NewScanner(os.Stdin)
}

func Example_N9epush() {
	testSample := n9epush.NewN9EMetrics("testmetrics", "192.168.1.1", "26", "dwadwa",
		100, 1000, 1630898446, float64(2232), nil)

	reg := prometheus.NewRegistry()
	reg.MustRegister(testSample)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(":" + strconv.Itoa(int(defaultHttpPort)), nil)
}

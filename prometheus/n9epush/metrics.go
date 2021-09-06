package n9epush

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"net/http"
	"os"
	"strconv"
	"time"
)

func init() {
	// 1. 获取环境中通用标签 Done
	// 2. 读取指定文件或者 pod ENV 环境 Done
	// 3. 解析出通用标签 Done
	// 4. 初始化指定特殊 Tag。项目配置或者自定义 Done
	// 5. 初始化时指定时延区间
	// 6. 定义 push 上报端口

	// 1. 定义指标名
	// 2. 内部指标处理
	// 3. 定义逻辑执行前时间
	// 4. 定义自定义 Tag, 非必要
	// 5. 执行业务代码
	//    添加自定义 tag 或者 不带

	// 请求数 Gauge
	// metricName: user_request
	// 返回码 tag: retcode=0, retcode=1, recode=-1...
	// 错误 tag: success=0, success=1
	// 调用类型 tag: type=rpc, type=internal
	// 业务自定义 tag: tag=tags
	// val: 周期内次数累加

	// 时延分布区间 Histogram
	// metricName: user_request__time.histogram
	// 返回码 tag: retcode=0, retcode=1, recode=-1...
	// 错误 tag: success=0, success=1
	// 调用类型 tag: type=rpc, type=internal
	// 业务自定义 tag: tag=tags
	// 区间:
	// 20,val1
	// 50,val2
	// 500,val3
	// 1000,val4
	// 3000,val5

	// 时延分布中位值 Summary
	// metricName: user_request__time.histogram
	// 返回码 tag: retcode=0, retcode=1, recode=-1...
	// 错误 tag: success=0, success=1
	// 调用类型 tag: type=rpc, type=internal
	// 业务自定义 tag: tag=tags
	// 区间:
	// 50,val1
	// 80,val2
	// 90,val3
	// 95,val4
	// 99,val5
}

var (
	subsystem = "rd"
	isFailure bool
	defaultPushGateway = "http://localhost:2080/v1/push"

	code = make(map[string]string)
)

func NewRequestVec() *prometheus.CounterVec {
	userReqSuccessVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name: "_user_requests_total",
			Help: "Number of success requests",
		},
		[]string{},
	)
	return userReqSuccessVec
}

func NewFailsVec() *prometheus.CounterVec {
	userReqFails := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name: "_user_fails_total",
			Help: "Number of fails requests",
		},
		[]string{},
	)
	return userReqFails
}

func NewDurationVec() *prometheus.HistogramVec {
	userDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name: "_user_request_duration",
			Help: "Duration of request",
		},
		[]string{},
	)
	return userDuration
}

func NewRetCodeVec(label string) *prometheus.CounterVec {
	returnCode := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name: "_return_code",
			Help: "return code",
		},
		[]string{label},
	)
	return returnCode
}

func NewDefaultMode(buckets []float64, isFailure bool, codes []string) *DefaultMode {
	if len(buckets) == 0 {
		buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10}
	}

	// reg := prometheus.NewRegistry()
	userRequestVec := NewRequestVec()
	userFailsVec := NewFailsVec()
	retCodeVec := NewRetCodeVec("code")
	durationVec := NewDurationVec()

	code := make(map[string]string)
	if isFailure && len(codes) != 0 {
		for _, v := range codes {
			if _, ok := code[v]; !ok {
				code[v] = v
			}
		}
	}

	defaultMode := &DefaultMode{
		IsFailureCodes: isFailure,
		UserDurationVec: durationVec,
		UserRequestVec: userRequestVec,
		UserFailsVec: userFailsVec,
		ReturnCodeVec: retCodeVec,
		Codes: code,
	}

	prometheus.MustRegister(defaultMode.UserDurationVec)
	prometheus.MustRegister(defaultMode.UserRequestVec)
	prometheus.MustRegister(defaultMode.UserFailsVec)
	prometheus.MustRegister(defaultMode.ReturnCode)

	// buckets TODO
	return defaultMode
}

// 用户主动添加返回码
func (dm *DefaultMode) MakeCode(retCode ...string) {
	if len(dm.Codes) > 0 && dm.IsFailureCodes {
		for _, v := range retCode {
			if _, ok := dm.Codes[v]; ok {
				dm.UserFailsVec.WithLabelValues().Inc()
			}
		}
	}
	dm.ReqsAdd(float64(len(retCode)))
	dm.ReturnCodeVec.WithLabelValues(retCode...).Inc()
}

// 用户请求数加一
func (dm *DefaultMode) ReqsInc() {
	dm.UserRequestVec.WithLabelValues().Inc()
}

// 用户请求数加传入参数
func (dm *DefaultMode) ReqsAdd(reqs float64) {
	dm.UserRequestVec.WithLabelValues().Add(reqs)
}

// 用户请求数带标签加一
func (dm *DefaultMode) ReqsLabelInc(labels ...string) {
	dm.UserRequestVec.WithLabelValues(labels...).Inc()
}

func (dm *DefaultMode) MakeDuration(inDuration float64, labels ...string) {
	if len(labels) == 0 {
		dm.UserRequestVec.WithLabelValues().Inc()
		dm.UserDurationVec.WithLabelValues().Observe(inDuration)
	} else {
			dm.UserRequestVec.WithLabelValues(labels...).Inc()
			dm.UserDurationVec.WithLabelValues(labels...).Observe(inDuration)
	}
}

// 新建 Metrics 方法
func NewMetrics(bu, project, appName string, globalTags map[string]string) (metrics *Metrics) {
	registry := prometheus.NewRegistry()
	collector := collectors.NewGoCollector()
	metrics.Registry = registry
	metrics.Collector = collector

	if bu == "" || project == "" || appName == "" {
		glog.Fatal("please check if bu or project or appname if it's empty")
		return &Metrics{}
	}

	tags := make(map[string]string)
	// collector := prometheus.NewGoCollector()
	// prometheus.MustRegister(collector)
	// 将用户自定义 tag 写入
	if len(globalTags) != 0 {
		for k, v := range globalTags {
			tags[k] = v
		}
	}
	// 强制写固定的bu, project, appname到全局 tags
	tags["bu"] = bu
	tags["project"] = project
	tags["app"] = appName
	// 获取 pod 名称
	podName := os.Getenv("MY_POD_NAME")
	if podName == "" {
		podName = strconv.Itoa(os.Getpid())
	}
	tags["pid"] = podName
	// set global tags TODO
	return metrics
}

// PushGateway 循环推送
func (m *Metrics) StartPushLoop(nid string, routine time.Duration, pg string) {
	// 如果 nid 未指定，推送 n9e 默认租户 0
	if nid == "" {
		nid = "0"
	}
	// 推送周期态最小为 1s
	if routine < 1 {
		routine = time.Second * 1
	}
	// pg 默认本机 2080 端口的 n9e
	if pg == "" {
		pg = defaultPushGateway
	}

	m.Registry.MustRegister(m.Collector)
	pusher := push.New(defaultPushGateway, "rd")
	for {
		err := pusher.Push()
		if err != nil {
			glog.Fatal(err)
		}
		time.Sleep(routine)
	}
}

// PullHttpServer
func (m *Metrics) StartPullHttpServer(port int64) {

	http.Handle("/metrics", promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{}))
	http.ListenAndServe(":" + strconv.Itoa(int(port)), nil)
}

//type N9EMetrics struct {
//	Metric      string                `json:"metric"`
//  Nid         string                `json:"nid"`
//	Endpoint    string                `json:"endpoint,omitempty"`
//	Timestamp   int64                 `json:"timestamp,omitempty"`
//	Step        int64                 `json:"step"`
//	Value       float64               `json:"value"`
//  N9EDesc     *prometheus.Desc      `json:"n9edesc"`
//	Tags        interface{}           `json:"tags,omitempty"`
//	TagsMap     map[string]string     `json:"tagsMap,omitempty"`
//	Extra       string                `json:"extra,omitempty"`
//}

// NewN9EMetrics
func NewN9EMetrics(metrics, endpoints, nid, URL string, batchSize, step, ts int64, values float64, tags map[string]string) (n9eMetrics *N9EMetrics) {
	n9eMetrics = &N9EMetrics{
		Metric: metrics,
		Nid: nid,
		Endpoint: endpoints,
		Timestamp: ts,
		Step: step,
		Value: values,
		TagsMap: tags,
		N9EDesc: prometheus.NewDesc(
			metrics,
			"User define help",
			[]string{"endpoint"},
			prometheus.Labels{"nid": n9eMetrics.Nid, "endpoint": n9eMetrics.Endpoint,
				"timestamp": strconv.Itoa(int(n9eMetrics.Timestamp)),
				"step": strconv.Itoa(int(n9eMetrics.Step)),
			},
		),
	}
	return
}

// N9EMetrics Describe
func (n9e *N9EMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- n9e.N9EDesc
}

// N9EMetrics Collect
func (n9e *N9EMetrics) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		n9e.N9EDesc,
		prometheus.GaugeValue,
		n9e.Value,
		n9e.Endpoint,
	)
}

package n9epush

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// 数据模型
//{
//        "nid": "24",
//        "metric": "test.buss_n_service_n",
//        "endpoint": "ip_3.n",
//        "tags": "app_name=/buss_1/buss_2/test.service_n",
//        "value": 15.4,
//        "timestamp": 1628217974,
//        "step": 20,
//        "counterType": "GAUGE"
//    }

// Metrics 定义，Bu, project, App 需要上传
// Routine，Nid，PushGateway 不传按默认
type Metrics struct {
	Registry    *prometheus.Registry  `json:"registry"`
	Collector   prometheus.Collector  `json:"collector"`

	Bu          string                `json:"bu"`
	Project     string                `json:"project"`
	App         string                `json:"app"`

	Routine     time.Time             `json:"routine,omitempty"`
	Nid         string                `json:"nid,omitempty"`
	PushGateway string                `json:"pushgateway,omitempty"`
	GlobalTags  []string              `json:"globaltags,omitempty"`
}

type N9EMetrics struct {
	Metric      string                `json:"metric"`
	Nid         string                `json:"nid"`
	Endpoint    string                `json:"endpoint,omitempty"`
	Timestamp   int64                 `json:"timestamp,omitempty"`
	Step        int64                 `json:"step"`
	Value       float64               `json:"value"`
	N9EDesc     *prometheus.Desc      `json:"n9edesc"`
	Tags        interface{}           `json:"tags,omitempty"`
	TagsMap     map[string]string     `json:"tagsMap,omitempty"`
	Extra       string                `json:"extra,omitempty"`
}

type Countertype struct {
	Gauge         prometheus.Gauge
	Counter       prometheus.Counter
	Histogram     prometheus.Histogram
	Summary       prometheus.Summary
	GaugeOpts     prometheus.GaugeOpts
	CounterOpts   prometheus.CounterOpts
	SummaryOpts   prometheus.SummaryOpts
	HistogramOpts prometheus.HistogramOpts
}

type DefaultMode struct {
	IsFailureCodes         bool                     `json:"isfailurecodes"`
	UserDuration           prometheus.Histogram     `json:"userduration,omitempty"`
	UserDurationVec        *prometheus.HistogramVec `json:"userdurationvec,omitempty"`
	UserRequest            prometheus.Counter       `json:"userrequest,omitempty"`
	UserRequestVec         *prometheus.CounterVec   `json:"userrequestvec,omitempty"`
	UserFails              prometheus.Counter       `json:"userfails,omitempty"`
	UserFailsVec           *prometheus.CounterVec   `json:"userfailsvec,omitempty"`
	ReturnCode             prometheus.Counter       `json:"returncode,omitempty"`
	ReturnCodeVec          *prometheus.CounterVec   `json:"returncodevec,omitempty"`
	Codes                  map[string]string        `json:"codes,omitempty"`
}

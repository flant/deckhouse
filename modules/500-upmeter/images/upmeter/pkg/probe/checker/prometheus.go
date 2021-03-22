package checker

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/tidwall/gjson"

	"upmeter/pkg/check"
	"upmeter/pkg/kubernetes"
)

// PrometheusApiAvailable is a checker constructor and configurator
type PrometheusApiAvailable struct {
	Access   *kubernetes.Access
	Timeout  time.Duration
	Endpoint string
}

func (c PrometheusApiAvailable) Checker() check.Checker {
	verifier := prometheusAPIVerifier{
		endpoint:     c.Endpoint,
		kubeAccessor: c.Access,
	}
	checker := newHTTPChecker(insecureClient, verifier)
	return withTimeout(checker, c.Timeout)
}

type prometheusAPIVerifier struct {
	endpoint     string
	kubeAccessor *kubernetes.Access
}

func (v prometheusAPIVerifier) Request() *http.Request {
	req, err := newGetRequest(v.endpoint, v.kubeAccessor.ServiceAccountToken())
	if err != nil {
		panic(err)
	}
	return req
}

/*
	Expecting JSON like this

	{
	  "status": "success",
	  "data": {
	    "resultType": "vector",
	    "result": [
	      {
	        "metric": {},
	        "value": [
	          1613143228.991,
	          "1"               <- we check this
	        ]
	      }
	    ]
	  }
	}
*/
func (v prometheusAPIVerifier) Verify(body []byte) check.Error {
	path := "data.result.0.value.1"
	value := gjson.Get(string(body), path)

	if value.String() != "1" {
		return check.ErrFail(`cannot find "1" in path %q prometheus data %q`, path, body)
	}

	return nil
}

// MetricPresentInPrometheus is a checker constructor and configurator
type MetricPresentInPrometheus struct {
	Access   *kubernetes.Access
	Timeout  time.Duration
	Endpoint string
	Metric   string
}

func (c MetricPresentInPrometheus) Checker() check.Checker {
	verifier := &metricPresenceVerifier{
		access:   c.Access,
		endpoint: addMetricQuery(c.Endpoint, c.Metric),
	}

	checker := newHTTPChecker(insecureClient, verifier)

	return withTimeout(checker, c.Timeout)
}

type metricPresenceVerifier struct {
	access   *kubernetes.Access
	endpoint string
}

func (v *metricPresenceVerifier) Request() *http.Request {
	req, err := newGetRequest(v.endpoint, v.access.ServiceAccountToken())
	if err != nil {
		panic(err)
	}
	return req
}

/*
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [                 <- array must not be empty
      {
        "metric": {},
        "value": [
          1614179019.102,
          "24"                  <- string number must not be zero
        ]
      }
    ]
  }
}

*/
func (v *metricPresenceVerifier) Verify(body []byte) check.Error {
	resultPath := "data.result"
	result := gjson.Get(string(body), resultPath)

	if !result.IsArray() {
		return check.ErrFail("cannot parse path %q in prometheus response %q", resultPath, body)
	}

	if len(result.Array()) == 0 {
		return check.ErrFail("no metrics in prometheus response (did not count)")
	}

	countPath := "data.result.0.value.1"
	count := gjson.Get(string(body), countPath)
	if count.String() == "0" {
		return check.ErrFail("no metrics in prometheus response (zero count)")
	}

	return nil
}

func addMetricQuery(baseUrl, metricName string) string {
	endpoint, err := url.Parse(baseUrl)
	if err != nil {
		panic(fmt.Errorf("cannot parse baseUrl: %v", err))
	}

	query := make(url.Values)
	// e.g. ?query=count(kubelet_node_name)
	query.Set("query", fmt.Sprintf("count(%s)", metricName))
	endpoint.RawQuery = query.Encode()

	return endpoint.String()
}
/*
Copyright 2021 Flant CJSC
Licensed under the Deckhouse Platform Enterprise Edition (EE) license. See https://github.com/deckhouse/deckhouse/ee/LICENSE
*/

package madison

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/deckhouse/deckhouse/go_lib/dependency"
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	OnBeforeHelm: &go_hook.OrderedConfig{Order: 20},
}, dependency.WithExternalDependencies(registrationHandler))

const (
	madisonBaseURL          = "https://madison.flant.com"
	madisonUpdateURLPattern = madisonBaseURL + "/api/%s/self_update/%s"
	madisonRevokeURLPattern = madisonBaseURL + "/api/%s/self_status/%s"

	registrationURL = "https://connect.deckhouse.io/v1/madison_register"
)

func registrationHandler(input *go_hook.HookInput, dc dependency.Container) error {
	clusterName := input.Values.Get("global.clusterName").String()
	if clusterName == "" {
		input.LogEntry.Error("global clusterName required")
		return nil
	}

	projectName := input.Values.Get("global.project").String()
	if projectName == "" {
		input.LogEntry.Error("global project required")
		return nil
	}

	data, err := createMadisonPayload(input.Values, dc, clusterName)
	if err != nil {
		return err
	}

	const (
		madisonKeyPath = "flantIntegration.madisonAuthKey"
		licenseKeyPath = "flantIntegration.internal.licenseKey"
	)

	if authKey, ok := input.Values.GetOk(madisonKeyPath); ok {
		key := authKey.String()

		// form request
		endpoint := fmt.Sprintf(madisonUpdateURLPattern, projectName, key)
		req, err := newMadisonRequest(endpoint, data)
		if err != nil {
			input.LogEntry.Errorf("http request failed: %v", err)
			return nil
		}

		// call
		_, err = doMadisonRequest(req, dc, input.LogEntry)
		if err != nil {
			err = fmt.Errorf("cannot update in madison (%s %s): %v", req.Method, req.URL, err)
			input.LogEntry.Errorf(err.Error())
			return err
		}

	} else if licenseKey, ok := input.Values.GetOk(licenseKeyPath); ok {
		data.Type = "prometheus"
		key := licenseKey.String()

		// form request to d8-connect proxy
		req, err := newRegistrationRequest(registrationURL, data, projectName, key)
		if err != nil {
			input.LogEntry.Errorf("http request failed: %v", err)
			return nil
		}

		// call
		authKey, err := doMadisonRequest(req, dc, input.LogEntry)
		if err != nil {
			err := fmt.Errorf("cannot register in madison (%s %s): %v", req.Method, req.URL, err)
			input.LogEntry.Errorf(err.Error())
			return err
		}
		if authKey != "" {
			input.ConfigValues.Set(madisonKeyPath, authKey)
		}
	}

	return nil
}

type madisonRequestData struct {
	Type          string    `json:"type,omitempty"`
	Name          string    `json:"name"`
	PrometheusURL string    `json:"prometheus_url"`
	GrafanaURL    string    `json:"grafana_url"`
	ExtraData     extraData `json:"extra_data"`
}

type extraData struct {
	Labels map[string]string `json:"labels"`
}

func createMadisonPayload(values *go_hook.PatchableValues, dc dependency.Container, clusterName string) (madisonRequestData, error) {
	data := madisonRequestData{
		Name: "kubernetes-" + clusterName,
		ExtraData: extraData{
			Labels: map[string]string{
				"kubernetes": clusterName,
			},
		},
	}

	schema := "http"
	publicDomain := values.Get("global.modules.publicDomainTemplate").String()

	if publicDomain != "" {
		globalHTTPSMode := values.Get("global.modules.https.mode").String()

		kubeCl, err := dc.GetK8sClient()
		if err != nil {
			return madisonRequestData{}, fmt.Errorf("cannot init Kubernetes client: %v", err)
		}

		cm, err := kubeCl.CoreV1().
			ConfigMaps("d8-system").
			Get(context.TODO(), "deckhouse", metav1.GetOptions{})
		if err != nil {
			return madisonRequestData{}, fmt.Errorf("cannot get configmap deckhouse")
		}

		prometheusHTTPSMode := ""
		prometheusData, ok := cm.Data["prometheus"]
		if ok {
			var prometheus struct{ HTTPS struct{ Mode string } }
			err := yaml.Unmarshal([]byte(prometheusData), &prometheus)
			if err == nil {
				prometheusHTTPSMode = prometheus.HTTPS.Mode
			}
		}

		if prometheusHTTPSMode == "" {
			if globalHTTPSMode != "Disabled" {
				schema = "https"
			}
		} else if prometheusHTTPSMode != "Disabled" {
			schema = "https"
		}
	}

	data.GrafanaURL = schema + "://" + fmt.Sprintf(publicDomain, "grafana")
	data.PrometheusURL = data.GrafanaURL + "/prometheus"
	return data, nil
}

type madisonAuthKeyResp struct {
	Error   string `json:"error"`
	AuthKey string `json:"auth_key"`
}

// doMadisonRequest makes auth request and expect response in form of Madison API
func doMadisonRequest(req *http.Request, dc dependency.Container, logEntry *logrus.Entry) (string, error) {
	resp, err := dc.GetHTTPClient().Do(req)
	if err != nil {
		logEntry.Errorf("http call failed: %s", err)
		return "", nil
	}
	defer resp.Body.Close()

	var madisonResp madisonAuthKeyResp
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read response body: %v", err)
	}
	err = json.Unmarshal(body, &madisonResp)
	if err != nil {
		logEntry.Errorf("json unmarshaling failed, body=%q: %v", body, err)
		return "", err
	}

	if madisonResp.Error != "" {
		return "", fmt.Errorf(madisonResp.Error)
	}

	return madisonResp.AuthKey, nil
}

type registrationData struct {
	Project string `json:"project"`
	Payload string `json:"madisonData"`
}

func newRegistrationRequest(endpoint string, data madisonRequestData, projectName, key string) (*http.Request, error) {
	madisonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal madison request data")
	}
	proxyData := registrationData{
		Project: projectName,
		Payload: string(madisonData),
	}
	proxyPayload, err := json.Marshal(proxyData)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal d8-connect request data")
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(proxyPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func newMadisonRequest(endpoint string, data madisonRequestData) (*http.Request, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal madison request data")
	}
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

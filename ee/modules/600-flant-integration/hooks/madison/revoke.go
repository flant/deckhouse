/*
Copyright 2021 Flant CJSC
Licensed under the Deckhouse Platform Enterprise Edition (EE) license. See https://github.com/deckhouse/deckhouse/ee/LICENSE
*/

package madison

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"

	"github.com/deckhouse/deckhouse/go_lib/dependency"
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue: "/modules/flant-integration/madison_revoke",
	Schedule: []go_hook.ScheduleConfig{
		{
			Name:    "madison_revoke",
			Crontab: "*/5 * * * *",
		},
	},
}, dependency.WithExternalDependencies(revokeHandler))

type madisonResponse struct {
	Error string `json:"error"`
}

func revokeHandler(input *go_hook.HookInput, dc dependency.Container) error {
	project, ok := input.Values.GetOk("global.project")
	if !ok || project.String() == "" {
		input.LogEntry.Error("global project required")
		return nil // cronjob was with allowFailure: true, so we just log errors
	}

	key, ok := input.Values.GetOk("flantIntegration.madisonAuthKey")
	if !ok {
		return nil
	}

	uri := fmt.Sprintf(madisonRevokeURLPattern, project, key.String())

	httpCli := dc.GetHTTPClient()
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		input.LogEntry.Errorf("http request failed: %s", err)
		return nil
	}

	resp, err := httpCli.Do(req)
	if err != nil {
		input.LogEntry.Errorf("http call failed: %s", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		return nil // OK
	}

	var madisonResp madisonResponse
	err = json.NewDecoder(resp.Body).Decode(&madisonResp)
	if err != nil {
		input.LogEntry.Errorf("json unmarshaling failed: %s", err)
		return nil // dont this we need an error
	}

	if madisonResp.Error == "Archived setup" {
		input.ConfigValues.Remove("flantIntegration.licenseKey")
		input.ConfigValues.Remove("flantIntegration.madisonAuthKey")
	}

	return nil
}

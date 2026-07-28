package e2etests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	exutil "github.com/openshift/origin/test/extended/util"
	compat_otp "github.com/openshift/origin/test/extended/util/compat_otp"
	"k8s.io/apimachinery/pkg/util/wait"
)

func getConfiguredAlertRules(oc *exutil.CLI, ruleName string, namespace string) (string, error) {
	return oc.AsAdmin().WithoutNamespace().Run("get").Args("prometheusrules", ruleName, "-o=jsonpath='{.spec.groups[*].rules[*].alert}'", "-n", namespace).Output()
}

type prometheusAlertResult struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
		} `json:"result"`
	} `json:"data"`
}

// getAlertLabels queries Prometheus for an alert and returns its labels.
func getAlertLabels(oc *exutil.CLI, alertName string) (map[string]string, error) {
	bearerToken := getSAToken(oc, "prometheus-k8s", "openshift-monitoring")
	promRoute := "https://" + getRouteAddress(oc, "openshift-monitoring", "prometheus-k8s")
	query := fmt.Sprintf(`ALERTS{alertname="%s"}`, alertName)

	h := make(http.Header)
	h.Add("Content-Type", "application/json")
	h.Add("Authorization", "Bearer "+bearerToken)

	params := url.Values{}
	params.Add("query", query)

	resp, err := doHTTPRequest(h, promRoute, "/api/v1/query", params.Encode(), "GET", false, 5, nil, 200)
	if err != nil {
		return nil, err
	}

	var result prometheusAlertResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if len(result.Data.Result) == 0 {
		return nil, nil
	}
	return result.Data.Result[0].Metric, nil
}

func waitForAlertToBePending(oc *exutil.CLI, alertName string) {
	bearerToken := getSAToken(oc, "prometheus-k8s", "openshift-monitoring")
	promRoute := "https://" + getRouteAddress(oc, "openshift-monitoring", "prometheus-k8s")
	query := fmt.Sprintf(`ALERTS{alertname="%s"}`, alertName)

	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 300*time.Second, false, func(context.Context) (done bool, err error) {
		res, qErr := queryPrometheus(promRoute, query, bearerToken)
		if qErr != nil {
			return false, nil
		}
		if len(res.Data.Result) == 0 {
			return false, nil
		}
		return true, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("%s Alert did not become pending/active", alertName))
}

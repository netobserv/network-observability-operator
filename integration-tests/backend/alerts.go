package e2etests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
)

// prometheusAlertResult the response of querying prometheus ALERTS metric
type prometheusAlertResult struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
		} `json:"result"`
	} `json:"data"`
}

func getConfiguredAlertRules(ruleName string, namespace string) (string, error) {
	obj, err := getDynamicResource("prometheusRule", ruleName, namespace)
	if err != nil {
		return "", err
	}

	groups, found, _ := unstructured.NestedSlice(obj.Object, "spec", "groups")
	if !found {
		return "", fmt.Errorf("no spec.groups found in prometheusrule %s", ruleName)
	}

	var alertNames []string
	for _, g := range groups {
		group, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		rules, _, _ := unstructured.NestedSlice(group, "rules")
		for _, r := range rules {
			rule, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			if alert, ok := rule["alert"].(string); ok {
				alertNames = append(alertNames, alert)
			}
		}
	}
	return strings.Join(alertNames, " "), nil
}

func getAlertLabels(alertName string) (map[string]string, error) {
	bearerToken := getSAToken("prometheus-k8s", "openshift-monitoring")
	promRoute := "https://" + getRouteAddress("openshift-monitoring", "prometheus-k8s")
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
		return nil, fmt.Errorf("failed to unmarshal alert result: %w", err)
	}

	if len(result.Data.Result) == 0 {
		return nil, nil
	}

	return result.Data.Result[0].Metric, nil
}

func waitForAlertToBePending(alertName string) {
	bearerToken := getSAToken("prometheus-k8s", "openshift-monitoring")
	promRoute := "https://" + getRouteAddress("openshift-monitoring", "prometheus-k8s")
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
	assertWaitPollNoErr(err, fmt.Sprintf("%s Alert did not become pending", alertName))
}

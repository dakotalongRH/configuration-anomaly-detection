package aiassisted

import (
	"encoding/json"
	"testing"
)

func TestParsePrometheusLabels(t *testing.T) {
	tests := []struct {
		name     string
		firing   string
		expected map[string]string
	}{
		{
			name: "PVC alert with multiple labels",
			firing: `Labels:
 - alertname = KubePersistentVolumeFillingUp
 - endpoint = https-metrics
 - instance = 10.91.81.110:10250
 - job = kubelet
 - namespace = openshift-monitoring
 - node = ip-10-91-81-110.us-west-2.compute.internal
 - persistentvolumeclaim = prometheus-data-prometheus-k8s-0
 - severity = critical`,
			expected: map[string]string{
				"alertname":              "KubePersistentVolumeFillingUp",
				"endpoint":               "https-metrics",
				"instance":               "10.91.81.110:10250",
				"job":                    "kubelet",
				"namespace":              "openshift-monitoring",
				"node":                   "ip-10-91-81-110.us-west-2.compute.internal",
				"persistentvolumeclaim":  "prometheus-data-prometheus-k8s-0",
				"severity":               "critical",
			},
		},
		{
			name:     "empty firing string",
			firing:   "",
			expected: map[string]string{},
		},
		{
			name:     "no labels section",
			firing:   "Some random text without labels",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePrometheusLabels(tt.firing)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d labels, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, ok := result[key]; !ok {
					t.Errorf("Missing expected label: %s", key)
				} else if actualValue != expectedValue {
					t.Errorf("Label %s: expected %q, got %q", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestPayloadStructure(t *testing.T) {
	// Simulate what gets sent to Cora based on the Q3OFA2AM9FYDAB alert
	customDetails := map[string]interface{}{
		"alert_name": "KubePersistentVolumeFillingUp",
		"cluster_id": "d9045cbe-348d-4f7b-9f4b-ccaaacdc0977",
		"firing": `Labels:
 - alertname = KubePersistentVolumeFillingUp
 - namespace = openshift-monitoring
 - persistentvolumeclaim = prometheus-data-prometheus-k8s-0
 - node = ip-10-91-81-110.us-west-2.compute.internal
 - severity = critical`,
		"link":        "https://github.com/openshift/runbooks/blob/master/alerts/cluster-monitoring-operator/KubePersistentVolumeFillingUp.md",
		"num_firing":  "2",
		"ocm_link":    "https://console.redhat.com/openshift/details/d9045cbe-348d-4f7b-9f4b-ccaaacdc0977",
		"region":      "us-west-2",
	}

	// Simulate payload construction logic
	payloadData := make(map[string]interface{})

	// Parse Prometheus labels
	if firing, ok := customDetails["firing"].(string); ok && firing != "" {
		prometheusLabels := parsePrometheusLabels(firing)
		if len(prometheusLabels) > 0 {
			payloadData["labels"] = prometheusLabels
		}
	}

	// Add filtered custom details (excluding redundant fields)
	filteredDetails := make(map[string]interface{})
	for key, value := range customDetails {
		switch key {
		case "alert_name", "cluster_id", "firing", "resolved":
			continue // Skip redundant/verbose fields
		default:
			filteredDetails[key] = value
		}
	}
	if len(filteredDetails) > 0 {
		payloadData["alert_details"] = filteredDetails
	}

	// Verify the structure
	labels, hasLabels := payloadData["labels"].(map[string]string)
	if !hasLabels {
		t.Fatal("Expected labels in payload")
	}

	// Verify critical labels are present
	criticalLabels := []string{"namespace", "persistentvolumeclaim"}
	for _, label := range criticalLabels {
		if _, ok := labels[label]; !ok {
			t.Errorf("Missing critical label: %s", label)
		}
	}

	// Verify namespace is correct
	if labels["namespace"] != "openshift-monitoring" {
		t.Errorf("Expected namespace openshift-monitoring, got %s", labels["namespace"])
	}

	// Verify PVC name is correct
	if labels["persistentvolumeclaim"] != "prometheus-data-prometheus-k8s-0" {
		t.Errorf("Expected PVC prometheus-data-prometheus-k8s-0, got %s", labels["persistentvolumeclaim"])
	}

	// Verify alert_details contains useful info but not redundant data
	alertDetails, hasAlertDetails := payloadData["alert_details"].(map[string]interface{})
	if !hasAlertDetails {
		t.Fatal("Expected alert_details in payload")
	}

	// Should have link, num_firing, etc.
	if _, ok := alertDetails["link"]; !ok {
		t.Error("Expected link in alert_details")
	}

	// Should NOT have alert_name (redundant with top-level)
	if _, ok := alertDetails["alert_name"]; ok {
		t.Error("alert_name should not be in alert_details (redundant)")
	}

	// Should NOT have firing (verbose text blob, parsed into labels)
	if _, ok := alertDetails["firing"]; ok {
		t.Error("firing should not be in alert_details (verbose, parsed into labels)")
	}

	// Print JSON representation for manual inspection
	jsonBytes, _ := json.MarshalIndent(payloadData, "", "  ")
	t.Logf("Payload structure:\n%s", string(jsonBytes))
}

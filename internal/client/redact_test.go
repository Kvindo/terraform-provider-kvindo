package client

import (
	"strings"
	"testing"
)

// Regression coverage for review finding #8: TF_LOG=DEBUG logged every request/response body
// verbatim, including secrets (root_password, private_key_pem, tokens, kubeconfigs, ...).

func TestRedactSecretsForLog_MasksKnownKeysAtAnyDepth(t *testing.T) {
	body := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "abc123", "name": "test"},
		"spec": map[string]interface{}{
			"rootPassword": "hunter2",
			"tier":         "standard", // not a secret key - must survive unchanged
		},
		"status": map[string]interface{}{
			"token":      "eyJ...",
			"kubeconfig": "apiVersion: v1\n...",
			"state":      "stable", // not a secret key
		},
	}

	redacted := redactSecretsForLog(body).(map[string]interface{})
	spec := redacted["spec"].(map[string]interface{})
	status := redacted["status"].(map[string]interface{})

	if spec["rootPassword"] != redactedPlaceholder {
		t.Errorf("expected spec.rootPassword to be redacted, got %v", spec["rootPassword"])
	}
	if spec["tier"] != "standard" {
		t.Errorf("expected spec.tier to survive unredacted, got %v", spec["tier"])
	}
	if status["token"] != redactedPlaceholder {
		t.Errorf("expected status.token to be redacted, got %v", status["token"])
	}
	if status["kubeconfig"] != redactedPlaceholder {
		t.Errorf("expected status.kubeconfig to be redacted, got %v", status["kubeconfig"])
	}
	if status["state"] != "stable" {
		t.Errorf("expected status.state to survive unredacted, got %v", status["state"])
	}

	metadata := redacted["metadata"].(map[string]interface{})
	if metadata["id"] != "abc123" || metadata["name"] != "test" {
		t.Errorf("expected metadata to survive unredacted, got %v", metadata)
	}
}

func TestRedactSecretsForLog_RedactsInsideLists(t *testing.T) {
	body := map[string]interface{}{
		"spec": map[string]interface{}{
			"gitlabInstances": []interface{}{
				map[string]interface{}{"runnerToken": "tok1", "url": "https://gitlab.example"},
			},
		},
	}
	// runnerToken isn't in secretJSONKeys by name (the gitlab_runner nested-object sensitivity is
	// a Terraform-schema concern, not this log-redaction one) - this test instead proves the
	// recursion reaches into list elements at all, using a key that IS in the map.
	body["spec"].(map[string]interface{})["gitlabInstances"].([]interface{})[0].(map[string]interface{})["password"] = "leaked"

	redacted := redactSecretsForLog(body).(map[string]interface{})
	instances := redacted["spec"].(map[string]interface{})["gitlabInstances"].([]interface{})
	item := instances[0].(map[string]interface{})
	if item["password"] != redactedPlaceholder {
		t.Errorf("expected password nested inside a list element to be redacted, got %v", item["password"])
	}
	if item["url"] != "https://gitlab.example" {
		t.Errorf("expected non-secret list element field to survive, got %v", item["url"])
	}
}

func TestRedactedJSONForLog_NonJSONBody_ReturnsPlaceholder(t *testing.T) {
	got := redactedJSONForLog([]byte("<html>Bad Gateway</html>"))
	if got != "<non-JSON body>" {
		t.Errorf("expected the non-JSON placeholder, got %q", got)
	}
}

func TestRedactedJSONForLog_RealBody_DoesNotLeakSecret(t *testing.T) {
	got := redactedJSONForLog([]byte(`{"status":{"token":"super-secret-value"}}`))
	if strings.Contains(got, "super-secret-value") {
		t.Errorf("expected the secret value to be redacted from the logged JSON, got %q", got)
	}
	// marshalForLog disables HTML escaping (this is a log line, not HTML), so the placeholder
	// must appear literally, not HTML-escaped as <redacted>.
	if !strings.Contains(got, redactedPlaceholder) {
		t.Errorf("expected the literal (non-HTML-escaped) redaction placeholder in the logged JSON, got %q", got)
	}
}

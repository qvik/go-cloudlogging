package cloudlogging

import (
	"fmt"
	"testing"

	gcloudlog "cloud.google.com/go/logging"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

func TestCreateAppEngineLogger(t *testing.T) {
	// Simply test compilation
	_ = MustNewAppEngineLogger()
}

func TestGoogleCloudLoggingLogger(t *testing.T) {
	projectID := "test"
	serviceID := "test"
	versionID := "test"

	// Create a monitored resource descriptor that will target GAE
	monitoredRes := &monitoredres.MonitoredResource{
		Type: "gae_app",
		Labels: map[string]string{
			"project_id": projectID,
			"module_id":  serviceID,
			"version_id": versionID,
		},
	}

	logEntries := make(map[string]gcloudlog.Entry)

	logHook := func(entry gcloudlog.Entry) {
		logEntries[fmt.Sprint(entry.Payload)] = entry
	}

	rootLog := MustNewLogger(
		WithGoogleCloudLogging(projectID, "", "gae_app", monitoredRes),
		withGoogleCloudLoggingUnitTestHook(logHook),
	)

	labels1 := []interface{}{"key1", "value1", "key2", false}
	baseLog := rootLog.WithAdditionalKeysAndValues(labels1...)

	labels2 := []interface{}{"key3", 123}
	log := baseLog.WithAdditionalKeysAndValues(labels2...)

	if log == baseLog {
		t.Error("indistinctive logger instances")
	}

	if log.logLevel != baseLog.logLevel {
		t.Error("distinctive log levels")
	}

	log.Debug("test1", "key3", "changed value", "key_number_4", true)
	entry1 := logEntries["test1"]
	if entry1.Labels["key1"] != "value1" {
		t.Error("value mismatch")
	}
	if entry1.Labels["key2"] != "false" {
		t.Error("value mismatch")
	}
	if entry1.Labels["key3"] != "changed value" {
		t.Error("value mismatch")
	}
	if entry1.Labels["key_number_4"] != "true" {
		t.Error("value mismatch")
	}

	newLog := log.WithAdditionalKeysAndValues("key1", "new value", "new_key", "all new value")

	newLog.Debug("test2", "key3", "changed once again")
	entry2 := logEntries["test2"]
	if entry2.Labels["key1"] != "new value" {
		t.Error("value mismatch")
	}
	if entry2.Labels["key2"] != "false" {
		t.Error("value mismatch")
	}
	if entry2.Labels["key3"] != "changed once again" {
		t.Error("value mismatch")
	}
	if entry2.Labels["new_key"] != "all new value" {
		t.Error("value mismatch")
	}
}

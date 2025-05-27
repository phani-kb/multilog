package multilog

import (
	"log/slog"
	"testing"
)

func TestReplaceAttr(t *testing.T) {
	replaceMap := map[string]string{
		"old_key1": "new_key1",
		"old_key2": "new_key2",
	}

	replaceFunc := ReplaceAttr(replaceMap)

	attr1 := slog.Attr{Key: "old_key1", Value: slog.StringValue("value1")}
	result1 := replaceFunc(nil, attr1)
	if result1.Key != "new_key1" {
		t.Errorf("Expected key to be replaced with 'new_key1', got '%s'", result1.Key)
	}

	attr2 := slog.Attr{Key: "unchanged_key", Value: slog.StringValue("value")}
	result2 := replaceFunc(nil, attr2)
	if result2.Key != "unchanged_key" {
		t.Errorf("Expected key to remain 'unchanged_key', got '%s'", result2.Key)
	}
}

func TestRemoveKeys(t *testing.T) {
	removeFunc := RemoveKeys()

	timeAttr := slog.Attr{Key: slog.TimeKey, Value: slog.StringValue("time_value")}
	result1 := removeFunc(nil, timeAttr)
	if result1.Key != "" {
		t.Errorf("Expected time key to be removed, got key '%s'", result1.Key)
	}

	levelAttr := slog.Attr{Key: slog.LevelKey, Value: slog.StringValue("level_value")}
	result2 := removeFunc(nil, levelAttr)
	if result2.Key != "" {
		t.Errorf("Expected level key to be removed, got key '%s'", result2.Key)
	}

	sourceAttr := slog.Attr{Key: slog.SourceKey, Value: slog.StringValue("source_value")}
	result3 := removeFunc(nil, sourceAttr)
	if result3.Key != "" {
		t.Errorf("Expected source key to be removed, got key '%s'", result3.Key)
	}

	msgAttr := slog.Attr{Key: slog.MessageKey, Value: slog.StringValue("message_value")}
	result4 := removeFunc(nil, msgAttr)
	if result4.Key != "" {
		t.Errorf("Expected message key to be removed, got key '%s'", result4.Key)
	}

	otherAttr := slog.Attr{Key: "other_key", Value: slog.StringValue("other_value")}
	result5 := removeFunc(nil, otherAttr)
	if result5.Key != "other_key" {
		t.Errorf("Expected other key to remain, got key '%s'", result5.Key)
	}
}

func TestRemoveGivenKeys(t *testing.T) {
	keysToRemove := []string{"key1", "key2", "key3"}
	removeFunc := RemoveGivenKeys(keysToRemove...)

	for _, key := range keysToRemove {
		attr := slog.Attr{Key: key, Value: slog.StringValue("value")}
		result := removeFunc(nil, attr)
		if result.Key != "" {
			t.Errorf("Expected key '%s' to be removed, got key '%s'", key, result.Key)
		}
	}

	otherAttr := slog.Attr{Key: "other_key", Value: slog.StringValue("other_value")}
	result := removeFunc(nil, otherAttr)
	if result.Key != "other_key" {
		t.Errorf("Expected 'other_key' to remain, got key '%s'", result.Key)
	}
}

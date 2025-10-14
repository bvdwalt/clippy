package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTickMsg(t *testing.T) {
	// Test that TickMsg can be created with a time.Time
	now := time.Now()
	tickMsg := TickMsg(now)

	// Verify the underlying time value
	if time.Time(tickMsg) != now {
		t.Errorf("Expected TickMsg to wrap time %v, got %v", now, time.Time(tickMsg))
	}
}

func TestTick(t *testing.T) {
	// Test that Tick() returns a command
	cmd := Tick()
	if cmd == nil {
		t.Error("Expected Tick() to return a non-nil command")
	}

	// We can't easily test the internal behavior of tea.Tick without
	// running the actual command, but we can verify it doesn't panic
	// and returns a proper tea.Cmd
}

func TestTickFunctionality(t *testing.T) {
	// Test the conceptual functionality of the tick system
	// In practice, this would be an integration test

	// Create a tick command
	cmd := Tick()

	// Verify it's not nil (basic smoke test)
	if cmd == nil {
		t.Fatal("Tick command should not be nil")
	}

	// In a real bubbletea program, this command would be executed
	// and would eventually produce a TickMsg after 2 seconds
	// We can't easily test the timing in a unit test without
	// making the test slow, but we can test the structure
}

func TestTickDuration(t *testing.T) {
	// Test that the tick is configured for 2 seconds
	// This is more of a documentation test since we can't
	// easily inspect the internal duration without running the command

	// The Tick function should create a command that fires every 2 seconds
	// This is tested implicitly through integration testing of the full UI

	// For now, just verify that calling Tick multiple times
	// returns different command instances (they should be independent)
	cmd1 := Tick()
	cmd2 := Tick()

	// Commands should be independent instances
	// Note: We can't directly compare tea.Cmd instances for equality
	// but we can verify they're both non-nil
	if cmd1 == nil || cmd2 == nil {
		t.Error("Both tick commands should be non-nil")
	}
}

func TestTickMsgConversion(t *testing.T) {
	// Test converting between time.Time and TickMsg
	originalTime := time.Date(2023, 10, 13, 12, 0, 0, 0, time.UTC)

	// Convert to TickMsg
	tickMsg := TickMsg(originalTime)

	// Convert back to time.Time
	convertedTime := time.Time(tickMsg)

	// Should be identical
	if !convertedTime.Equal(originalTime) {
		t.Errorf("Time conversion failed: expected %v, got %v", originalTime, convertedTime)
	}
}

func TestTickMsgZeroValue(t *testing.T) {
	// Test with zero time value
	var zeroTime time.Time
	tickMsg := TickMsg(zeroTime)

	convertedTime := time.Time(tickMsg)
	if !convertedTime.IsZero() {
		t.Error("Zero time should remain zero after conversion")
	}
}

func TestTickMsgWithNanoseconds(t *testing.T) {
	// Test that nanosecond precision is preserved
	timeWithNanos := time.Date(2023, 10, 13, 12, 0, 0, 123456789, time.UTC)

	tickMsg := TickMsg(timeWithNanos)
	convertedTime := time.Time(tickMsg)

	if convertedTime.Nanosecond() != 123456789 {
		t.Errorf("Nanoseconds not preserved: expected %d, got %d",
			123456789, convertedTime.Nanosecond())
	}
}

func TestTickMsgTimezone(t *testing.T) {
	// Test that timezone information is preserved
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("Could not load timezone for testing")
	}

	timeWithTZ := time.Date(2023, 10, 13, 12, 0, 0, 0, location)

	tickMsg := TickMsg(timeWithTZ)
	convertedTime := time.Time(tickMsg)

	if convertedTime.Location() != location {
		t.Errorf("Timezone not preserved: expected %v, got %v",
			location, convertedTime.Location())
	}
}

func TestTickIntegrationWithModel(t *testing.T) {
	// Test how TickMsg would be used in the context of the model
	// This is more of an integration test

	tickTime := time.Now()
	tickMsg := TickMsg(tickTime)

	// Verify that TickMsg can be used as a tea.Msg
	var msg tea.Msg = tickMsg

	// Should be able to type assert back
	switch m := msg.(type) {
	case TickMsg:
		if time.Time(m) != tickTime {
			t.Error("TickMsg type assertion failed to preserve time")
		}
	default:
		t.Error("TickMsg should be assignable to tea.Msg")
	}
}

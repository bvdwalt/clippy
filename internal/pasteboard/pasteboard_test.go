package pasteboard

import (
	"slices"
	"testing"
)

func TestHasConcealedMarker(t *testing.T) {
	tests := []struct {
		name  string
		types []string
		want  bool
	}{
		{"nil", nil, false},
		{"plain text", []string{"public.utf8-plain-text", "NSStringPboardType"}, false},
		{"macOS concealed", []string{"public.utf8-plain-text", "org.nspasteboard.ConcealedType"}, true},
		{"macOS transient", []string{"org.nspasteboard.TransientType", "public.utf8-plain-text"}, true},
		{"KDE password manager hint", []string{"text/plain", "x-kde-passwordManagerHint"}, true},
		{"marker as substring only", []string{"org.nspasteboard.ConcealedTypeExtra"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasConcealedMarker(tt.types); got != tt.want {
				t.Errorf("hasConcealedMarker(%v) = %v, want %v", tt.types, got, tt.want)
			}
		})
	}
}

func TestParseTypeList(t *testing.T) {
	got := parseTypeList([]byte("TIMESTAMP\nTARGETS\n\n  x-kde-passwordManagerHint  \ntext/plain\n"))
	want := []string{"TIMESTAMP", "TARGETS", "x-kde-passwordManagerHint", "text/plain"}
	if !slices.Equal(got, want) {
		t.Errorf("parseTypeList() = %q, want %q", got, want)
	}
}

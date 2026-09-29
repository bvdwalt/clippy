// Package pasteboard reads clipboard metadata that atotto/clipboard does not expose.
package pasteboard

import "strings"

// concealedMarkers are clipboard types that password managers attach to secrets.
var concealedMarkers = map[string]struct{}{
	"org.nspasteboard.ConcealedType": {},
	"org.nspasteboard.TransientType": {},
	"x-kde-passwordManagerHint":      {},
}

// IsConcealed reports whether a password manager marked the current clipboard as secret.
func IsConcealed() (bool, error) {
	types, err := clipboardTypes()
	if err != nil {
		return false, err
	}
	return hasConcealedMarker(types), nil
}

func hasConcealedMarker(types []string) bool {
	for _, t := range types {
		if _, ok := concealedMarkers[t]; ok {
			return true
		}
	}
	return false
}

// parseTypeList splits newline-separated type names, as printed by wl-paste and xclip.
func parseTypeList(out []byte) []string {
	var types []string
	for _, line := range strings.Split(string(out), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			types = append(types, t)
		}
	}
	return types
}

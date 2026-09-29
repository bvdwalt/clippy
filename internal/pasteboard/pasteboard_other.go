//go:build !darwin

package pasteboard

import (
	"os"
	"os/exec"
)

// ChangeCount is only available on macOS; elsewhere ok is false.
func ChangeCount() (int, bool) {
	return 0, false
}

// clipboardTypes follows atotto's tool order; xsel and termux can't list types.
func clipboardTypes() ([]string, error) {
	var cmd *exec.Cmd
	switch {
	case os.Getenv("WAYLAND_DISPLAY") != "" && hasCommand("wl-paste"):
		cmd = exec.Command("wl-paste", "--list-types")
	case hasCommand("xclip"):
		cmd = exec.Command("xclip", "-selection", "clipboard", "-t", "TARGETS", "-o")
	default:
		return nil, nil
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseTypeList(out), nil
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

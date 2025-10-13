# Clippy

A lightweight terminal-based clipboard history manager, built with Go and the Bubble Tea TUI framework.

## Features

- 📋 **Automatic Clipboard Monitoring** - Continuously tracks clipboard changes in real-time
- 🕒 **Persistent History** - Saves clipboard history to disk across sessions
- 🎯 **Duplicate Detection** - Automatically filters out duplicate entries using SHA-256 hashing
- ⌨️ **Keyboard Navigation** - Navigate through history with vim-style keybindings
- 📱 **Clean Terminal UI** - Beautiful, responsive interface that fits your workflow
- 🔄 **Instant Copy** - Copy any historical item back to clipboard with a single keypress

## Installation

### Prerequisites

- Go 1.20 or later
- MacOS
- Linux with X11 or Wayland (uses `xclip` or `wl-clipboard`)

### Install from source

```bash
git clone https://github.com/bvdwalt/clippy.git
cd clippy
go build -o clippy ./cmd/clippy
sudo mv clippy /usr/local/bin/
```

### Install dependencies (if needed)

For X11:
```bash
sudo apt install xclip  # Ubuntu/Debian
sudo dnf install xclip  # Fedora
```

For Wayland:
```bash
sudo apt install wl-clipboard  # Ubuntu/Debian
sudo dnf install wl-clipboard  # Fedora
```

## Usage

Start the clipboard history manager:

```bash
clippy
```

### Keybindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Navigate up through history |
| `↓` / `j` | Navigate down through history |
| `Enter` | Copy selected item to clipboard |
| `q` / `Ctrl+C` | Quit application |

## How It Works

Clippy monitors your system clipboard every 2 seconds and automatically captures any new content. Each clipboard entry is:

1. **Hashed** using SHA-256 to detect duplicates
2. **Timestamped** for chronological organization  
3. **Persisted** to `history.json` in the current directory
4. **Displayed** in a scrollable terminal interface

The application shows a preview of each clipboard entry (truncated to 60 characters) and replaces newlines with spaces for clean display.

## Project Structure

```
clippy/
├── cmd/clippy/           # Main application entry point
│   └── main.go
├── internal/
│   ├── history/          # Clipboard history management
│   │   ├── types.go      # Data structures
│   │   └── history.go    # History manager
│   └── ui/               # Terminal user interface
│       ├── model.go      # Bubble Tea model
│       └── commands.go   # UI commands
├── history.json          # Persistent clipboard history (created at runtime)
├── go.mod
└── README.md
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [clipboard](https://github.com/atotto/clipboard) - Cross-platform clipboard access

## Privacy & Security

- Clipboard history is stored locally in `history.json`
- No data is transmitted over the network
- SHA-256 hashes are used only for duplicate detection, not security
- All clipboard content is stored in plain text locally

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the [MIT License](LICENSE).

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm
- As well as [Clipboard](https://github.com/atotto/clipboard)
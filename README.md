# Clippy
[![Build Format Test](https://github.com/bvdwalt/clippy/actions/workflows/build_format_test.yml/badge.svg)](https://github.com/bvdwalt/clippy/actions/workflows/build_format_test.yml)
[![codecov](https://codecov.io/gh/bvdwalt/clippy/branch/main/graph/badge.svg)](https://codecov.io/gh/bvdwalt/clippy)

A lightweight terminal-based clipboard history manager, built with Go and the Bubble Tea TUI framework.

## Features

- 📋 **Automatic Clipboard Monitoring** - Continuously tracks clipboard changes in real-time
- 🕒 **Persistent History** - Saves clipboard history to disk across sessions
- 🎯 **Duplicate Detection** - Automatically filters out duplicate entries using SHA-256 hashing
- ⌨️ **Keyboard Navigation** - Navigate through history with vim-style keybindings
- 📱 **Clean Terminal UI** - Beautiful, responsive interface that fits your workflow
- 🔄 **Instant Copy** - Copy any historical item back to clipboard with a single keypress

## Demo
![Demo app showing some clipboard items](<demo/demo.png>)

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
just run
```

### Keybindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Navigate up through history |
| `↓` / `j` | Navigate down through history |
| `Enter` / `c` | Copy selected item to clipboard |
| `d` | Delete selected item from history |
| `/` | Enter search mode |
| `r` | Refresh/clear search results |
| `Esc` | Exit search mode (when in search) |
| `q` / `Ctrl+C` | Quit application |

#### Search Mode
When you press `/`, you'll enter search mode where you can:
- Type to filter clipboard history using fuzzy search (similar to fzf)
- Press `Enter` to apply the search filter
- Press `Esc` to cancel and return to normal view

## How It Works

Clippy monitors your system clipboard every 2 seconds and automatically captures any new content. Each clipboard entry is:

1. **Hashed** using SHA-256 to detect duplicates
2. **Timestamped** for chronological organization
3. **Persisted** to `clipp.db` in the config dir `~/.clippy/clippy.db`
4. **Displayed** in a scrollable terminal interface

The application shows a preview of each clipboard entry (truncated to 60 characters) and replaces newlines with spaces for clean display.

## Project Structure

```
clippy/
├── cmd/clippy/           # Main application entry point
│   ├── main.go           # Application entry point
│   └── main_test.go      # Main package tests
├── demo/                 # Demo application
│   └── main.go           # Demo runner
├── internal/
│   ├── history/          # Clipboard history management
│   │   ├── history.go    # History manager implementation
│   │   ├── types.go      # Data structures and types
│   │   └── *_test.go     # History package tests
│   ├── search/           # Fuzzy search functionality
│   │   ├── fuzzy.go      # Fuzzy search implementation
│   │   └── *_test.go     # Search package tests
│   └── ui/               # Terminal user interface
│       ├── model.go      # Bubble Tea model
│       ├── commands.go   # UI commands and messaging
│       ├── styles/       # UI styling and themes
│       │   └── theme.go  # Color themes and styling
│       ├── table/        # Table display management
│       │   └── manager.go # Table rendering and state
│       └── *_test.go     # UI package tests
├── go.mod                # Go module definition
├── go.sum                # Go module dependencies
└── README.md             # Project documentation
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components for Bubble Tea
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - TUI styling
- [clipboard](https://github.com/atotto/clipboard) - Cross-platform clipboard access

## Privacy & Security

- Clipboard history is stored locally in `~/.clippy/clippy.db`
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

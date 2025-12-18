# Fana-Envy

Fana-Envy is a powerful, persistent terminal environment manager built with Go. It allows you to manage environment variables with a TUI, featuring interactive terminals, persistent history, and seamless environment switching.

![fana-envy UI](image.png)

## Features

- **Environments**: Switch between sets of environment variables instantly (renamed from Profiles).
- **Multiple Terminals**: Manage multiple terminal instances in tabs/panes with dynamic naming (e.g., changes to `python` when running python).
- **Interactive Experience**:
  - Full support for interactive commands (e.g., Python `input()`, REPLs).
  - **Unbuffered Output**: Automatically injects `PYTHONUNBUFFERED=1` so Python scripts output immediately.
  - **Colored Output**: Forces color output (`FORCE_COLOR=1`, `CLICOLOR_FORCE=1`) for better visibility in the TUI.
- **Persistent History**: Command history is saved relative to the application binary and deduplicated to avoid clutter.
- **Environment Editor**: Built-in editor to modify environment variables on the fly (`.env` format).
- **Cross-Platform**: Works on Windows, macOS, and Linux.
- **Portable**: Configuration and history are stored next to the executable.

## Installation

```bash
git clone https://github.com/MasFana/fana-envy.git
cd fana-envy
go build -o envy.exe ./cmd/fana-envy
```

## Usage

Run the executable:

```bash
./envy.exe
```

### Shortcuts

| Shortcut            | Description                |
| ------------------- | -------------------------- |
| `Ctrl+N`            | New Terminal               |
| `Ctrl+W`            | Close Terminal             |
| `Ctrl+H` / `Ctrl+L` | Switch Terminal Left/Right |
| `Ctrl+E`            | Toggle Environment Editor  |
| `Ctrl+D`            | Exit Application           |

### Commands

| Command             | Description                                           |
| ------------------- | ----------------------------------------------------- |
| `open`              | Open the `envs` folder in your system's file explorer |
| `new <name>`        | Create a new environment                              |
| `switch <name>`     | Switch to an environment                              |
| `set <KEY> <VALUE>` | Set an environment variable                           |
| `unset <KEY>`       | Remove a variable                                     |
| `cd <path>`         | Change directory                                      |
| `env`               | List current environment variables                    |
| `clear`             | Clear terminal output                                 |
| `exit`              | Quit the application                                  |

## Folder Structure

The project follows the standard Go project layout:

```
fana-envy/
├── cmd/
│   └── fana-envy/    # Entry point
├── internal/
│   ├── config/       # Configuration & History
│   ├── styles/       # UI styling (Lipgloss)
│   ├── terminal/     # Terminal pane logic
│   ├── tui/          # Main Bubble Tea model & view
│   └── utils/        # Helper functions
└── README.md
```

## Configuration

Configuration files (`envs/*.env`) and history (`.fana_history`) are stored in the directory where the binary is located. This allows you to carry the tool on a USB drive or move it between folders without losing your settings.

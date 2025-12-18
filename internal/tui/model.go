package tui

import (
	"github.com/MasFana/fana-envy/internal/terminal"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

type Mode int

const (
	ModeTerminal Mode = iota
	ModeProfiles
	ModeInput
	ModeEditor
)

// Model is the main application state
type Model struct {
	// Terminals
	Terminals []*terminal.TerminalPane
	ActiveIdx int
	NextID    int

	// Profile state
	CurrentProfile  string
	RootPath        string
	ConfigPath      string // Path where config/envs are stored
	EnvVars         map[string]string
	Profiles        []string
	SelectedIdx     int
	Editor          textarea.Model // Full text editor
	OriginalContent string         // Logic to track changes

	// Editor Header
	FilenameInput textinput.Model
	HeaderFocus   bool

	// Mode
	Mode Mode

	// History
	History    []string
	HistoryIdx int
	GitBranch  string

	// UI
	Width  int
	Height int

	// Input Overlay
	InputModel   textinput.Model
	InputPurpose string // "new", "delete"

	// Autocomplete State
	Completions    []string
	CompletionIdx  int
	CompletionBase string // The string prefix we are completing against (e.g. "no")

	// State
	Quitting bool
}

// Messages
type OutputMsg struct {
	TermID int
	Line   string
}

type CmdDoneMsg struct {
	TermID int
	Err    error
}

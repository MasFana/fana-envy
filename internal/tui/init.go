package tui

import (
	"os"
	"path/filepath"

	"github.com/MasFana/fana-envy/internal/config"
	"github.com/MasFana/fana-envy/internal/styles"
	"github.com/MasFana/fana-envy/internal/terminal"
	"github.com/MasFana/fana-envy/internal/utils"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func InitialModel() Model {
	cwd, _ := os.Getwd()
	// Use executable directory for storage
	exeDir := utils.GetExecutableDir()
	envDir := filepath.Join(exeDir, config.EnvFolderName)
	os.MkdirAll(envDir, 0755)

	// Load config
	cfg := config.LoadConfig(envDir)
	profileName := cfg.LastProfile
	if profileName == "" {
		profileName = "default"
	}

	ti := textinput.New()
	ti.Cursor.Style = styles.Running
	ti.Prompt = ""

	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.Prompt = " "

	fi := textinput.New()
	fi.Cursor.Style = styles.Selected
	fi.Prompt = ""
	fi.Placeholder = "Profile Name"

	m := Model{
		Terminals:      []*terminal.TerminalPane{terminal.NewTerminalPane(1)},
		ActiveIdx:      0,
		NextID:         2,
		CurrentProfile: profileName,
		RootPath:       cwd,    // Keep RootPath as CWD for file operations
		ConfigPath:     exeDir, // New field for config storage location
		EnvVars:        make(map[string]string),
		Mode:           ModeTerminal,
		History:        utils.LoadHistory(exeDir), // Load history from exe dir
		HistoryIdx:     -1,
		Width:          100,
		Height:         30,
		InputModel:     ti,
		Editor:         ta,
		FilenameInput:  fi,
	}

	// Load profile
	m.LoadProfile(filepath.Join(envDir, profileName+".env"))
	m.UpdateGitBranch()
	m.LoadProfiles()

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

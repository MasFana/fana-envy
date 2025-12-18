package tui

import (
	"fmt"

	"github.com/MasFana/fana-envy/internal/styles"
	"github.com/MasFana/fana-envy/internal/terminal"
	"github.com/MasFana/fana-envy/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.UpdateViewportSizes()
		return m, nil

	case tea.MouseMsg:
		if m.Mode == ModeTerminal && len(m.Terminals) > 0 {
			t := m.Terminals[m.ActiveIdx]
			var cmd tea.Cmd
			t.Viewport, cmd = t.Viewport.Update(msg)
			return m, cmd
		}
		return m, nil

	case OutputMsg:
		for _, t := range m.Terminals {
			if t.ID == msg.TermID {
				t.AddOutput(msg.Line)
				break
			}
		}
		return m, nil

	case CmdDoneMsg:
		for _, t := range m.Terminals {
			if t.ID == msg.TermID {
				t.Mu.Lock()
				t.Running = false
				t.Cmd = nil
				t.Mu.Unlock()
				if msg.Err != nil {
					t.AddOutput(styles.Error.Render(fmt.Sprintf("Exit: %v", msg.Err)))
				}
				if t.OriginalName != "" {
					t.Name = t.OriginalName
					t.OriginalName = ""
				}
				break
			}
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Mode-specific handling
	switch m.Mode {
	case ModeTerminal:
		// Update active terminal's input
		if len(m.Terminals) > 0 {
			var cmd tea.Cmd
			t := m.Terminals[m.ActiveIdx]
			t.Input, cmd = t.Input.Update(msg)
			return m, cmd
		}
	case ModeInput:
		var cmd tea.Cmd
		m.InputModel, cmd = m.InputModel.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global shortcuts
	switch msg.String() {
	case "ctrl+n":
		// New terminal
		t := terminal.NewTerminalPane(m.NextID)
		m.NextID++
		t.AddOutput(styles.Muted.Render(fmt.Sprintf("── Terminal %d ──", t.ID)))
		m.Terminals = append(m.Terminals, t)
		m.ActiveIdx = len(m.Terminals) - 1
		m.UpdateViewportSizes()
		m.Mode = ModeTerminal
		return m, nil

	case "ctrl+w":
		// Close terminal
		if len(m.Terminals) > 1 {
			// Kill any running process
			t := m.Terminals[m.ActiveIdx]
			if t.Running && t.Cmd != nil {
				utils.KillProcess(t.Cmd)
			}
			m.Terminals = append(m.Terminals[:m.ActiveIdx], m.Terminals[m.ActiveIdx+1:]...)
			if m.ActiveIdx >= len(m.Terminals) {
				m.ActiveIdx = len(m.Terminals) - 1
			}
		}
		return m, nil

	case "ctrl+h", "ctrl+left":
		// Previous terminal
		if m.ActiveIdx > 0 {
			m.ActiveIdx--
		}
		m.Mode = ModeTerminal
		return m, nil

	case "ctrl+l", "ctrl+right":
		// Next terminal
		if m.ActiveIdx < len(m.Terminals)-1 {
			m.ActiveIdx++
		}
		m.Mode = ModeTerminal
		return m, nil

	case "ctrl+e":
		// Toggle profile editor
		if m.Mode == ModeTerminal {
			m.Mode = ModeProfiles
			m.LoadProfiles()
		} else {
			m.Mode = ModeTerminal
		}
		return m, nil

	case "ctrl+d":
		// Exit
		if m.Mode == ModeTerminal {
			t := m.Terminals[m.ActiveIdx]
			if !t.Running {
				m.SaveState()
				m.Quitting = true
				return m, tea.Quit
			}
		}
		return m, nil

	case "ctrl+c":
		// Kill running process
		if m.Mode == ModeTerminal {
			t := m.Terminals[m.ActiveIdx]
			if t.Running && t.Cmd != nil {
				utils.KillProcess(t.Cmd)
				t.AddOutput("^C")
			} else {
				t.Input.SetValue("")
			}
		}
		return m, nil
	}

	// Mode-specific handling
	switch m.Mode {
	case ModeTerminal:
		return m.handleTerminalKey(msg)
	case ModeProfiles:
		return m.handleProfileKey(msg)
	case ModeInput:
		return m.handleInputKey(msg)
	case ModeEditor:
		return m.handleEditorKey(msg)
	}
	return m, nil
}

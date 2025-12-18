package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MasFana/fana-envy/internal/config"
	"github.com/MasFana/fana-envy/internal/styles"
	"github.com/MasFana/fana-envy/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleTerminalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	t := m.Terminals[m.ActiveIdx]

	switch msg.Type {
	case tea.KeyEnter:
		// Check if running
		if t.Running {
			if t.Stdin != nil {
				// Pass to stdin
				inputLine := t.Input.Value() + "\n"
				t.Stdin.Write([]byte(inputLine))
				prompt := m.buildPromptText()
				t.AddOutput(prompt + t.Input.Value()) // Echo
				t.Input.SetValue("")
			}
			return m, nil
		}

		input := strings.TrimSpace(t.Input.Value())
		if input == "" {
			return m, nil
		}

		// Show command in output
		prompt := m.buildPromptText()

		// Add separator before command if not first
		if len(t.Output) > 0 {
			width := t.Viewport.Width
			sep := strings.Repeat("┈", width)
			t.AddOutput(styles.Muted.Render(sep))
		}

		t.AddOutput(prompt + input)

		// Add to history (if not same as last)
		if len(m.History) == 0 || m.History[len(m.History)-1] != input {
			m.History = append(m.History, input)
		}
		m.HistoryIdx = len(m.History)
		utils.SaveHistory(m.ConfigPath, m.History)

		t.Input.SetValue("")

		// Execute command
		return m.ExecuteCommand(input)

	case tea.KeyUp:
		if len(m.History) > 0 && m.HistoryIdx > 0 {
			m.HistoryIdx--
			t.Input.SetValue(m.History[m.HistoryIdx])
			t.Input.CursorEnd()
		}
		return m, nil

	case tea.KeyDown:
		if m.HistoryIdx < len(m.History)-1 {
			m.HistoryIdx++
			t.Input.SetValue(m.History[m.HistoryIdx])
			t.Input.CursorEnd()
		} else {
			m.HistoryIdx = len(m.History)
			t.Input.SetValue("")
		}
		return m, nil

	case tea.KeyPgUp:
		t.Viewport.LineUp(5)
		return m, nil

	case tea.KeyPgDown:
		t.Viewport.LineDown(5)
		return m, nil

	case tea.KeyShiftUp:
		t.Viewport.LineUp(1)
		return m, nil

	case tea.KeyTab:
		if len(m.Completions) > 0 {
			m.CompletionIdx = (m.CompletionIdx + 1) % len(m.Completions)
			t.Input.SetValue(m.Completions[m.CompletionIdx])
			t.Input.CursorEnd()
			return m, nil
		}

		input := t.Input.Value()
		candidates, base := m.GenerateCompletions(input)

		if len(candidates) > 0 {
			m.Completions = candidates
			m.CompletionIdx = 0
			m.CompletionBase = base
			t.Input.SetValue(candidates[0])
			t.Input.CursorEnd()
		}
		return m, nil

	default:
		m.Completions = nil
		m.CompletionIdx = 0
	}

	var cmd tea.Cmd
	t.Input, cmd = t.Input.Update(msg)
	return m, cmd
}

func (m Model) handleProfileKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedIdx > 0 {
			m.SelectedIdx--
			m.LoadEditorContent()
		}
		return m, nil

	case "down", "j":
		if m.SelectedIdx < len(m.Profiles)-1 {
			m.SelectedIdx++
			m.LoadEditorContent()
		}
		return m, nil

	case "enter":
		if len(m.Profiles) > 0 {
			m.CurrentProfile = m.Profiles[m.SelectedIdx]
			envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
			m.LoadProfile(filepath.Join(envDir, m.CurrentProfile+".env"))
			config.SaveConfig(envDir, m.CurrentProfile)
			t := m.Terminals[m.ActiveIdx]
			t.AddOutput(styles.Success.Render("✓ Switched to " + m.CurrentProfile))
			m.Mode = ModeTerminal
		}
		return m, nil

	case "n":
		m.Mode = ModeInput
		m.InputPurpose = "new"
		m.InputModel.Placeholder = "New profile name..."
		m.InputModel.SetValue("")
		m.InputModel.Focus()
		return m, nil

	case "d":
		if len(m.Profiles) > 0 {
			name := m.Profiles[m.SelectedIdx]
			t := m.Terminals[m.ActiveIdx]

			switch name {
			case m.CurrentProfile:
				t.AddOutput(styles.Error.Render("Cannot delete active profile"))
			case "default":
				t.AddOutput(styles.Error.Render("Cannot delete default"))
			default:
				m.Mode = ModeInput
				m.InputPurpose = "delete"
				m.InputModel.Placeholder = "Delete " + name + "? (y/n)"
				m.InputModel.SetValue("")
				m.InputModel.Focus()
			}
		}
		return m, nil

	case "r":
		if len(m.Profiles) > 0 {
			name := m.Profiles[m.SelectedIdx]
			if name == "default" {
				t := m.Terminals[m.ActiveIdx]
				t.AddOutput(styles.Error.Render("Cannot rename default"))
			} else {
				m.Mode = ModeInput
				m.InputPurpose = "rename"
				m.InputModel.Placeholder = "Rename " + name + " to..."
				m.InputModel.SetValue(name)
				m.InputModel.Focus()
			}
		}
		return m, nil

	case "tab":
		m.Mode = ModeEditor
		m.Editor.Focus()
		return m, nil

	case "esc":
		m.Mode = ModeTerminal
		return m, nil
	}

	return m, nil
}

func (m Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		value := strings.TrimSpace(m.InputModel.Value())

		switch m.InputPurpose {
		case "new":
			if value == "" {
				return m, nil
			}
			if !utils.IsValidProfileName(value) {
				m.InputPurpose = "error"
				m.InputModel.SetValue("Invalid Name")
				return m, nil
			}

			envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
			path := filepath.Join(envDir, value+".env")
			if _, err := os.Stat(path); err == nil {
				// Exists
			} else {
				content := fmt.Sprintf("# %s\n# Created: %s\n", value, time.Now().Format("2006-01-02"))
				os.WriteFile(path, []byte(content), 0644)
			}
			m.LoadProfiles()

			for i, p := range m.Profiles {
				if p == value {
					m.SelectedIdx = i
					break
				}
			}

			m.Mode = ModeEditor
			m.LoadEditorContent()
			m.Editor.Focus()
			m.InputModel.Blur()
			return m, nil

		case "delete":
			if strings.ToLower(value) == "y" || strings.ToLower(value) == "yes" {
				name := m.Profiles[m.SelectedIdx]
				if name != m.CurrentProfile && name != "default" {
					envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
					os.Remove(filepath.Join(envDir, name+".env"))
					m.LoadProfiles()
					if m.SelectedIdx >= len(m.Profiles) {
						m.SelectedIdx = len(m.Profiles) - 1
					}
				}
			}
		case "rename":
			if value == "" || !utils.IsValidProfileName(value) {
				// Invalid
			} else {
				m.TryRenameProfile(value)
			}

		case "confirm_save":
			if strings.ToLower(value) == "y" || strings.ToLower(value) == "yes" {
				m.SaveEditorContent()
				m.Mode = ModeProfiles
			} else if strings.ToLower(value) == "n" || strings.ToLower(value) == "no" {
				m.Editor.SetValue(m.OriginalContent)
				m.Mode = ModeProfiles
			}
		}

		if m.InputPurpose != "new" && m.InputPurpose != "confirm_save" {
			m.Mode = ModeProfiles
			m.InputModel.Blur()
		} else if m.InputPurpose == "confirm_save" {
			// stay or blur done above
			m.InputModel.Blur()
		}
		return m, nil

	case tea.KeyEsc:
		if m.InputPurpose == "confirm_save" {
			m.Mode = ModeEditor
			m.InputModel.Blur()
			m.Editor.Focus()
			return m, nil
		}

		m.Mode = ModeProfiles
		m.InputModel.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.InputModel, cmd = m.InputModel.Update(msg)
	return m, cmd
}

func (m Model) handleEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+s":
		m.SaveEditorContent()
		return m, nil
	case "esc", "tab":
		if m.HeaderFocus {
			m.HeaderFocus = false
			m.Editor.Focus()
			return m, nil
		}

		if m.Editor.Value() != m.OriginalContent {
			m.Mode = ModeInput
			m.InputPurpose = "confirm_save"
			m.InputModel.Placeholder = ""
			m.InputModel.SetValue("")
			m.InputModel.Focus()
			return m, nil
		}

		m.Mode = ModeProfiles
		return m, nil
	}

	if m.HeaderFocus {
		switch msg.Type {
		case tea.KeyDown, tea.KeyEnter:
			newName := strings.TrimSpace(m.FilenameInput.Value())
			if newName != "" && newName != m.Profiles[m.SelectedIdx] {
				m.TryRenameProfile(newName)
			}
			m.HeaderFocus = false
			m.Editor.Focus()
			return m, nil
		}

		var cmd tea.Cmd
		m.FilenameInput, cmd = m.FilenameInput.Update(msg)
		return m, cmd
	} else {
		switch msg.String() {
		case "up":
			if m.Editor.Line() == 0 {
				m.HeaderFocus = true
				m.Editor.Blur()
				m.FilenameInput.Focus()
				return m, nil
			}
		}

		var cmd tea.Cmd
		m.Editor, cmd = m.Editor.Update(msg)
		return m, cmd
	}
}

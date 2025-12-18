package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MasFana/fana-envy/internal/config"
	"github.com/MasFana/fana-envy/internal/styles"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Quitting {
		return styles.Success.Render("\n👋 Bye!\n")
	}

	mainHeight := m.Height - 1
	if mainHeight < 10 {
		mainHeight = 10
	}

	contentHeight := mainHeight - 2
	paneWidth := m.Width - styles.SidebarWidth - 6

	if paneWidth < 40 {
		paneWidth = 40
	}

	sidebar := m.buildSidebar(contentHeight)

	var pane string
	if m.Mode == ModeProfiles || m.Mode == ModeEditor {
		pane = m.buildProfilePane(paneWidth, contentHeight)
	} else {
		pane = m.buildTerminalPane(paneWidth, contentHeight)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", pane)
	status := m.buildStatusBar()
	mainView := lipgloss.JoinVertical(lipgloss.Left, content, status)

	if m.Mode == ModeInput {
		return m.overlayInput(mainView)
	}
	return mainView
}

func (m Model) overlayInput(underlying string) string {
	var title string
	var prompt string

	switch m.InputPurpose {
	case "new":
		title = " Create Environment "
		prompt = "Enter environment name:"
	case "delete":
		title = " Confirm Delete "
		prompt = "Are you sure? (y/n)"
	case "rename":
		title = " Rename Environment "
		prompt = "Enter new name:"
	case "confirm_save":
		title = " Unsaved Changes "
		prompt = "Save changes before exiting? (y/n)"
	case "error":
		title = " Error "
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Magenta).
		Padding(1, 2).
		Width(40).
		Align(lipgloss.Center).
		Render(
			styles.Title.Render(title) + "\n\n" +
				styles.Normal.Render(prompt) + "\n" +
				m.InputModel.View() + "\n\n" +
				styles.Muted.Render("Enter: Confirm • Esc: Cancel"),
		)
	return PlaceOverlay(m.Width, m.Height, box, underlying)
}

func PlaceOverlay(width, height int, overlay, background string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, overlay,
		lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceForeground(lipgloss.Color("")))
}

func (m Model) buildSidebar(height int) string {
	var b strings.Builder

	termTitle := "  Terminals"
	if m.Mode == ModeTerminal {
		termTitle = styles.Title.Render(termTitle)
	} else {
		termTitle = styles.Muted.Render(termTitle)
	}
	b.WriteString(termTitle + "\n")
	b.WriteString(strings.Repeat("─", styles.SidebarWidth-4) + "\n")

	for i, t := range m.Terminals {
		marker := "  "
		style := styles.Normal
		if i == m.ActiveIdx && m.Mode == ModeTerminal {
			marker = "➤ "
			style = styles.Selected
		}

		status := ""
		if t.Running {
			status = styles.Running.Render(" ●")
		}

		name := t.Name
		if len(name) > styles.SidebarWidth-8 {
			name = name[:styles.SidebarWidth-8]
		}
		b.WriteString(marker + style.Render(name) + status + "\n")
	}

	profTitle := "  Environment"
	if m.Mode == ModeProfiles {
		profTitle = styles.Title.Render(profTitle)
	} else {
		profTitle = styles.Muted.Render(profTitle)
	}
	b.WriteString("\n" + profTitle + "\n")
	b.WriteString(strings.Repeat("─", styles.SidebarWidth-4) + "\n")

	for i, p := range m.Profiles {
		isActive := (p == m.CurrentProfile)
		isSelected := (i == m.SelectedIdx)

		var line strings.Builder
		cursor := "  "
		if isSelected {
			if m.Mode == ModeProfiles {
				cursor = "➤ "
			} else if m.Mode == ModeEditor {
				cursor = "✎ "
			}
		}

		line.WriteString(cursor)

		activeMark := "  "
		if isActive {
			activeMark = styles.Success.Render("● ")
		}
		line.WriteString(activeMark)

		nameStyle := styles.Muted
		if isSelected {
			if m.Mode == ModeProfiles {
				nameStyle = styles.Selected
			} else if m.Mode == ModeEditor {
				nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
			}
		} else if isActive {
			nameStyle = styles.Normal
		}

		dispName := p
		maxLen := styles.SidebarWidth - 8
		if len(dispName) > maxLen {
			dispName = dispName[:maxLen] + "…"
		}

		line.WriteString(nameStyle.Render(dispName) + "\n")

		b.WriteString(line.String())
	}

	if m.Mode == ModeProfiles {
		b.WriteString("\n" + styles.Muted.Render("↑↓:nav Enter:sel"))
	}

	return styles.Sidebar.Height(height).Render(b.String())
}

func (m Model) buildTerminalPane(width, height int) string {
	if len(m.Terminals) == 0 {
		return styles.Pane.Width(width).Height(height).Render("No terminals")
	}

	t := m.Terminals[m.ActiveIdx]
	t.Viewport.Width = width
	t.Viewport.Height = height - 4

	var b strings.Builder

	title := styles.Title.Render(" " + t.Name + " ")
	if t.Running {
		title += styles.Running.Render(" ● ")
	}
	b.WriteString(title + "\n")
	b.WriteString(strings.Repeat("─", width) + "\n")

	t.Viewport.SetContent(t.GetOutput())
	b.WriteString(t.Viewport.View() + "\n")

	b.WriteString(strings.Repeat("─", width) + "\n")

	prompt := m.buildPrompt()
	b.WriteString(prompt + t.Input.View())

	return styles.Pane.Width(width).Height(height).Render(b.String())
}

func (m Model) buildProfilePane(width, height int) string {
	var b strings.Builder

	profileName := ""
	if len(m.Profiles) > 0 && m.SelectedIdx < len(m.Profiles) {
		profileName = m.Profiles[m.SelectedIdx]
	}

	style := styles.Pane.Width(width).Height(height)
	if m.Mode == ModeEditor {
		style = style.BorderForeground(styles.Green)
	}

	headerStr := "Editor: "
	var nameStr string

	if m.HeaderFocus {
		nameStr = m.FilenameInput.View() + ".env"
	} else {
		nameStr = styles.Profile.Render(profileName + ".env")
	}

	b.WriteString(styles.Title.Render(headerStr) + nameStr + "\n")
	b.WriteString(strings.Repeat("─", width-4) + "\n")

	b.WriteString(m.Editor.View())

	hint := "n: new │ d: delete │ r: rename │ Tab: edit"
	if m.Mode == ModeEditor {
		hint = "Ctrl+S: save │ Esc/Tab: back"
	}

	return style.Render(b.String() + "\n" + styles.Muted.Render(hint))
}

func (m Model) buildStatusBar() string {
	left := fmt.Sprintf(" %s v%s │ [%s]", config.AppName, config.Version, m.CurrentProfile)
	shortcuts := "'help' | Ctrl+N:new │ Ctrl+H/L:switch | Ctrl+W:close │ Ctrl+E:env │ Ctrl+D:exit"

	gap := m.Width - len(left) - len(shortcuts)
	if gap < 0 {
		gap = 0
	}

	return styles.StatusBar.Width(m.Width).Render(left + strings.Repeat(" ", gap) + shortcuts)
}

func (m Model) buildPrompt() string {
	profile := styles.Profile.Render("[" + m.CurrentProfile + "]")

	cwd, _ := os.Getwd()
	dir := filepath.Base(cwd)
	if cwd == m.RootPath {
		dir = "~"
	}
	path := styles.Path.Render(dir)

	git := ""
	if m.GitBranch != "" {
		git = " " + styles.Git.Render("("+m.GitBranch+")")
	}

	return profile + " " + path + git + styles.Prompt.Render(" ➤ ")
}

func (m Model) buildPromptText() string {
	cwd, _ := os.Getwd()
	dir := filepath.Base(cwd)
	if cwd == m.RootPath {
		dir = "~"
	}
	git := ""
	if m.GitBranch != "" {
		git = " (" + m.GitBranch + ")"
	}
	return "[" + m.CurrentProfile + "] " + dir + git + " ➤ "
}

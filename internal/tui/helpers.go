package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MasFana/fana-envy/internal/config"
	"github.com/MasFana/fana-envy/internal/styles"
	"github.com/MasFana/fana-envy/internal/utils"
)

func (m *Model) UpdateViewportSizes() {
	paneWidth := m.Width - styles.SidebarWidth - 6
	if paneWidth < 40 {
		paneWidth = 40
	}

	mainHeight := m.Height - 1
	if mainHeight < 10 {
		mainHeight = 10
	}
	contentHeight := mainHeight - 2

	for _, t := range m.Terminals {
		t.Viewport.Width = paneWidth
		t.Viewport.Height = contentHeight - 4
		t.Input.Width = paneWidth - 20
	}

	editorH := contentHeight - 4
	if editorH < 1 {
		editorH = 1
	}
	m.Editor.SetWidth(paneWidth - 4)
	m.Editor.SetHeight(editorH)
}

func (m *Model) LoadProfiles() {
	m.Profiles = []string{}
	envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
	files, _ := os.ReadDir(envDir)
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".env") {
			m.Profiles = append(m.Profiles, strings.TrimSuffix(f.Name(), ".env"))
		}
	}
	sort.Strings(m.Profiles)

	for i, p := range m.Profiles {
		if p == m.CurrentProfile {
			m.SelectedIdx = i
			break
		}
	}

	m.LoadEditorContent()
}

func (m *Model) LoadEditorContent() {
	if len(m.Profiles) == 0 || m.SelectedIdx >= len(m.Profiles) {
		m.Editor.SetValue("No profiles found")
		return
	}

	name := m.Profiles[m.SelectedIdx]
	envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
	content, err := os.ReadFile(filepath.Join(envDir, name+".env"))
	if err != nil {
		m.Editor.SetValue("Error loading file: " + err.Error())
		return
	}

	m.Editor.SetValue(string(content))
	m.OriginalContent = string(content)
}

func (m *Model) TryRenameProfile(newName string) {
	if !utils.IsValidProfileName(newName) {
		return
	}
	if len(m.Profiles) == 0 {
		return
	}

	oldName := m.Profiles[m.SelectedIdx]
	if oldName == "default" {
		return
	}

	envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
	oldPath := filepath.Join(envDir, oldName+".env")
	newPath := filepath.Join(envDir, newName+".env")

	if _, err := os.Stat(newPath); err == nil {
		return
	}

	err := os.Rename(oldPath, newPath)
	if err == nil {
		m.LoadProfiles()
		if m.CurrentProfile == oldName {
			m.CurrentProfile = newName
			config.SaveConfig(envDir, m.CurrentProfile) // Fix: Pass string, not Model
		}
	}
}

func (m *Model) SaveEditorContent() {
	if len(m.Profiles) > 0 {
		name := m.Profiles[m.SelectedIdx]
		envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
		path := filepath.Join(envDir, name+".env")
		content := m.Editor.Value()
		os.WriteFile(path, []byte(content), 0644)

		if name == m.CurrentProfile {
			m.LoadProfile(path)
		}

		m.OriginalContent = m.Editor.Value()
	}
}

func (m *Model) BuildEnv() []string {
	env := os.Environ()
	for k, v := range m.EnvVars {
		env = append(env, k+"="+v)
	}
	return env
}

func (m *Model) UpdateGitBranch() {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err == nil {
		m.GitBranch = strings.TrimSpace(string(output))
	} else {
		m.GitBranch = ""
	}
}

func (m *Model) SaveState() {
	envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
	config.SaveConfig(envDir, m.CurrentProfile)
	utils.SaveHistory(m.ConfigPath, m.History)
}

func (m *Model) SaveProfile() {
	envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
	var lines []string

	path := filepath.Join(envDir, m.CurrentProfile+".env")
	if content, err := os.ReadFile(path); err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
				lines = append(lines, line)
			}
		}
	}

	keys := make([]string, 0, len(m.EnvVars))
	for k := range m.EnvVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		lines = append(lines, k+"="+m.EnvVars[k])
	}

	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

func (m *Model) LoadProfile(path string) {
	m.EnvVars = make(map[string]string)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			os.WriteFile(path, []byte("# Environment\n"), 0644)
		}
		return
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			m.EnvVars[key] = value
			os.Setenv(key, value)
		}
	}
}

func (m *Model) GetHelp() string {
	return `
` + styles.Title.Render("Commands") + `
  env           Show variables
  set K V       Set variable
  unset K       Remove variable
  switch NAME   Change profile
  new NAME      Create profile
  cd [DIR]      Change directory
  clear         Clear terminal
  exit          Quit

` + styles.Title.Render("Shortcuts") + `
  Ctrl+N        New terminal
  Ctrl+W        Close terminal
  Ctrl+H/L      Switch terminals
  Ctrl+E        Profile editor
  Ctrl+C        Kill process
  Ctrl+D        Exit`
}

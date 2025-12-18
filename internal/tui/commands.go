package tui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/MasFana/fana-envy/internal/config"
	"github.com/MasFana/fana-envy/internal/styles"
	"github.com/MasFana/fana-envy/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) ExecuteCommand(input string) (tea.Model, tea.Cmd) {
	parts := utils.SmartSplit(input)
	if len(parts) == 0 {
		return m, nil
	}

	cmd := parts[0]
	args := parts[1:]
	t := m.Terminals[m.ActiveIdx]

	switch cmd {
	case "exit", "quit":
		m.SaveState()
		m.Quitting = true
		return m, tea.Quit

	case "clear", "cls":
		t.Output = []string{}
		t.Viewport.SetContent("")
		return m, nil

	case "cd":
		dir := ""
		if len(args) > 0 {
			dir = args[0]
		} else {
			dir, _ = os.UserHomeDir()
		}
		if strings.HasPrefix(dir, "~") {
			home, _ := os.UserHomeDir()
			dir = filepath.Join(home, dir[1:])
		}
		if err := os.Chdir(dir); err != nil {
			t.AddOutput(styles.Error.Render(fmt.Sprintf("cd: %v", err)))
		} else {
			m.UpdateGitBranch()
		}
		return m, nil

	case "open":
		envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("explorer", envDir)
		case "darwin":
			cmd = exec.Command("open", envDir)
		default:
			cmd = exec.Command("xdg-open", envDir)
		}
		if err := cmd.Start(); err != nil {
			t.AddOutput(styles.Error.Render("Error opening folder: " + err.Error()))
		} else {
			t.AddOutput(styles.Success.Render("✓ Opened envs folder"))
		}
		return m, nil

	case "pwd":
		cwd, _ := os.Getwd()
		t.AddOutput(styles.Path.Render(cwd))
		return m, nil

	case "env":
		if len(m.EnvVars) == 0 {
			t.AddOutput(styles.Muted.Render("No variables"))
		} else {
			keys := make([]string, 0, len(m.EnvVars))
			for k := range m.EnvVars {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				t.AddOutput(styles.Profile.Render(k) + "=" + m.EnvVars[k])
			}
		}
		return m, nil

	case "set":
		if len(args) < 2 {
			t.AddOutput(styles.Error.Render("Usage: set KEY VALUE"))
			return m, nil
		}
		key := args[0]
		value := strings.Join(args[1:], " ")
		if !utils.IsValidEnvVar(key) {
			t.AddOutput(styles.Error.Render("Invalid variable name"))
			return m, nil
		}
		m.EnvVars[key] = value
		os.Setenv(key, value)
		m.SaveProfile()
		t.AddOutput(styles.Success.Render("✓ Set " + key))
		return m, nil

	case "unset":
		if len(args) < 1 {
			t.AddOutput(styles.Error.Render("Usage: unset KEY"))
			return m, nil
		}
		delete(m.EnvVars, args[0])
		os.Unsetenv(args[0])
		m.SaveProfile()
		t.AddOutput(styles.Success.Render("✓ Unset " + args[0]))
		return m, nil

	case "switch":
		if len(args) < 1 {
			t.AddOutput(styles.Muted.Render("Usage: switch <profile>"))
			return m, nil
		}
		name := args[0]
		envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
		path := filepath.Join(envDir, name+".env")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.AddOutput(styles.Error.Render("Not found: " + name))
			return m, nil
		}
		m.LoadProfile(path)
		m.CurrentProfile = name
		config.SaveConfig(envDir, m.CurrentProfile)
		t.AddOutput(styles.Success.Render("✓ Switched to " + name))
		return m, nil

	case "new":
		if len(args) < 1 {
			t.AddOutput(styles.Error.Render("Usage: new <name>"))
			return m, nil
		}
		name := args[0]
		if !utils.IsValidProfileName(name) {
			t.AddOutput(styles.Error.Render("Invalid name"))
			return m, nil
		}
		envDir := filepath.Join(m.ConfigPath, config.EnvFolderName)
		path := filepath.Join(envDir, name+".env")
		if _, err := os.Stat(path); err == nil {
			t.AddOutput(styles.Error.Render("Already exists"))
			return m, nil
		}
		content := fmt.Sprintf("# %s\n# Created: %s\n", name, time.Now().Format("2006-01-02"))
		os.WriteFile(path, []byte(content), 0644)
		m.LoadProfiles()
		t.AddOutput(styles.Success.Render("✓ Created " + name))
		return m, nil

	case "help":
		t.AddOutput(m.GetHelp())
		return m, nil
	}

	t.OriginalName = t.Name
	t.Name = cmd
	return m, m.RunExternalCmd(t.ID, cmd, args)
}

func (m *Model) RunExternalCmd(termID int, name string, args []string) tea.Cmd {
	return func() tea.Msg {
		for _, t := range m.Terminals {
			if t.ID == termID {
				c := exec.Command(name, args...)
				env := m.BuildEnv()
				env = append(env, "PYTHONUNBUFFERED=1")
				env = append(env, "FORCE_COLOR=1")
				env = append(env, "CLICOLOR_FORCE=1")
				c.Env = env

				stdout, _ := c.StdoutPipe()
				stderr, _ := c.StderrPipe()
				stdin, _ := c.StdinPipe()

				t.Mu.Lock()
				t.Cmd = c
				t.Stdin = stdin
				t.Running = true
				t.Mu.Unlock()

				if err := c.Start(); err != nil {
					return CmdDoneMsg{termID, err}
				}

				go func() {
					reader := bufio.NewReader(stdout)
					for {
						line, err := reader.ReadString('\n')
						if err != nil {
							break
						}
						t.AddOutput(strings.TrimRight(line, "\r\n"))
					}
				}()

				go func() {
					reader := bufio.NewReader(stderr)
					for {
						line, err := reader.ReadString('\n')
						if err != nil {
							break
						}
						t.AddOutput(styles.Error.Render(strings.TrimRight(line, "\r\n")))
					}
				}()

				err := c.Wait()
				return CmdDoneMsg{termID, err}
			}
		}
		return nil
	}
}

func (m Model) GenerateCompletions(input string) ([]string, string) {
	if input == "" {
		return nil, ""
	}

	var candidates []string
	seen := make(map[string]bool)
	add := func(s string) {
		if !seen[s] && s != "" {
			candidates = append(candidates, s)
			seen[s] = true
		}
	}

	if !strings.Contains(input, " ") {
		start := input
		cmds := []string{"help", "env", "set", "unset", "switch", "new", "open", "cd", "exit", "quit", "clear", "cls"}
		for _, cmd := range cmds {
			if strings.HasPrefix(cmd, start) {
				add(cmd)
			}
		}

		if len(start) >= 2 {
			pathEnv := os.Getenv("PATH")
			for _, dir := range filepath.SplitList(pathEnv) {
				entries, _ := os.ReadDir(dir)
				for _, e := range entries {
					name := e.Name()
					lowerName := strings.ToLower(name)
					lowerStart := strings.ToLower(start)
					if strings.HasPrefix(lowerName, lowerStart) {
						add(name)
					}
				}
				if len(candidates) > 50 {
					break
				}
			}
		}
	} else {
		parts := strings.Fields(input)
		cmd := parts[0]
		lastArg := ""
		prefix := input

		if !strings.HasSuffix(input, " ") {
			lastArg = parts[len(parts)-1]
			prefix = input[:len(input)-len(lastArg)]
		}

		switch cmd {
		case "switch":
			for _, p := range m.Profiles {
				if strings.HasPrefix(p, lastArg) {
					add(prefix + p)
				}
			}
		case "unset":
			for k := range m.EnvVars {
				if strings.HasPrefix(k, lastArg) {
					add(prefix + k)
				}
			}
		case "cd":
			cwd, _ := os.Getwd()
			entries, _ := os.ReadDir(cwd)
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), lastArg) {
					add(prefix + e.Name())
				}
			}
		}

		cwd, _ := os.Getwd()
		entries, _ := os.ReadDir(cwd)
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), lastArg) {
				add(prefix + e.Name())
			}
		}

		return candidates, input
	}

	return candidates, input
}

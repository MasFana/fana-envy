package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/MasFana/fana-envy/internal/config"
)

func LoadHistory(root string) []string {
	content, _ := os.ReadFile(filepath.Join(root, config.HistoryFile))
	var history []string
	for _, line := range strings.Split(string(content), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			history = append(history, line)
		}
	}
	if len(history) > 1000 {
		history = history[len(history)-1000:]
	}
	return history
}

func SaveHistory(root string, history []string) {
	if len(history) > 1000 {
		history = history[len(history)-1000:]
	}
	os.WriteFile(filepath.Join(root, config.HistoryFile), []byte(strings.Join(history, "\n")), 0644)
}

func SmartSplit(input string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range input {
		if (ch == '"' || ch == '\'') && !inQuote {
			inQuote = true
			quoteChar = ch
		} else if ch == quoteChar && inQuote {
			inQuote = false
		} else if ch == ' ' && !inQuote {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func IsValidEnvVar(name string) bool {
	if len(name) == 0 {
		return false
	}
	for i, ch := range name {
		if i == 0 && unicode.IsDigit(ch) {
			return false
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return false
		}
	}
	return true
}

func IsValidProfileName(name string) bool {
	if len(name) == 0 || len(name) > 50 {
		return false
	}
	for _, ch := range name {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '-' && ch != '_' {
			return false
		}
	}
	return true
}

func KillProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", cmd.Process.Pid)).Run()
}

func GetExecutableDir() string {
	ex, err := os.Executable()
	if err != nil {
		// Fallback to wd if executable path fails
		wd, _ := os.Getwd()
		return wd
	}
	return filepath.Dir(ex)
}

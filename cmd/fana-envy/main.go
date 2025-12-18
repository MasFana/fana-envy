package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/MasFana/fana-envy/internal/config"
	"github.com/MasFana/fana-envy/internal/tui"
	"github.com/MasFana/fana-envy/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "open" {
		exeDir := utils.GetExecutableDir()
		envDir := filepath.Join(exeDir, config.EnvFolderName)
		if err := os.MkdirAll(envDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Opening %s...\n", envDir)

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
			fmt.Fprintf(os.Stderr, "Error opening folder: %v\n", err)
			os.Exit(1)
		}
		return
	}

	setupConsole()

	p := tea.NewProgram(
		tui.InitialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

package terminal

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/MasFana/fana-envy/internal/styles"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// TerminalPane represents a single terminal session
type TerminalPane struct {
	ID       int
	Name     string
	Output   []string
	Input    textinput.Model
	Viewport viewport.Model
	Cmd      *exec.Cmd
	Running  bool
	Mu       sync.Mutex
}

func NewTerminalPane(id int) *TerminalPane {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60
	ti.Prompt = ""

	vp := viewport.New(60, 15)
	vp.SetContent("")

	return &TerminalPane{
		ID:       id,
		Name:     fmt.Sprintf("Term %d", id),
		Output:   []string{},
		Input:    ti,
		Viewport: vp,
	}
}

func (t *TerminalPane) AddOutput(line string) {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	t.Output = append(t.Output, line)
	if len(t.Output) > styles.MaxOutput {
		t.Output = t.Output[len(t.Output)-styles.MaxOutput:]
	}
	t.Viewport.SetContent(strings.Join(t.Output, "\n"))
	t.Viewport.GotoBottom()
}

func (t *TerminalPane) GetOutput() string {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	return strings.Join(t.Output, "\n")
}

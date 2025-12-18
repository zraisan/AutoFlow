package extractors

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Scripts struct {
	Lint    string
	Install string
	Test    string
	Build   string
	Deploy  string
}

type Directory struct {
	value textinput.Model
	err   error
}

type (
	errMsg error
)

func initialModel() Directory {
	ti := textinput.New()
	ti.Placeholder = "./"
	ti.SetValue("./")
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return Directory{
		value: ti,
		err:   nil,
	}
}

func (m Directory) Init() tea.Cmd {
	return textinput.Blink
}

func (m Directory) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.value, cmd = m.value.Update(msg)
	return m, cmd
}

func (m Directory) View() string {
	msg := fmt.Sprintf("Project Directory: %s", m.value.View())
	entries, _ := os.ReadDir(m.value.Value())
	for _, entry := range entries {
		if entry.IsDir() {
			msg += fmt.Sprintf("\n%s", entry.Name())
		}
	}
	return msg
}

func Extract(framework string) (Scripts, string, error) {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return Scripts{}, "", fmt.Errorf("TUI error: %w", err)
	}
	if directory, ok := m.(Directory); ok {
		switch framework {
		case "Node":
			return ReadNodeFiles(directory.value.Value())
		}
	}
	return Scripts{}, "",fmt.Errorf("unknown framework: %s", framework)
}

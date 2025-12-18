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
	choices []string
	cursor   int
	selected int
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
				case tea.KeyUp, tea.KeyCtrlK:
					if m.cursor > 0 {
						m.cursor--
					}
				case tea.KeyDown, tea.KeyCtrlJ:
					if m.cursor < len(m.choices)-1 {
						m.cursor++
					}
				case tea.KeySpace:
					m.selected = m.cursor
					m.value.SetValue(m.choices[m.selected])
			}
		case errMsg:
			m.err = msg
			return m, nil
	}

	m.value, cmd = m.value.Update(msg)
	entries, _ := os.ReadDir(m.value.Value())
	if len(m.choices) > 0 {
		m.choices = m.choices[:0]
	}
	for _, entry := range entries {
		if entry.IsDir() {
			m.choices = append(m.choices, entry.Name())
		}
	}
	return m, cmd
}

func (m Directory) View() string {
	msg := fmt.Sprintf("Project Directory: %s", m.value.View())
	for i, choice := range m.choices {
		cursor := ""
		if m.cursor == i {
			cursor = ">"
		}
		checked := " "
		if m.selected == i {
			checked = "x"
		}
		msg += fmt.Sprintf("\n %s [%s] %s", cursor, checked, choice)
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

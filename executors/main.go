package executors

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mzkux/AutoFlow/extractors"
)

type Extractor struct {
	choices  []string
	cursor   int
	selected int
}

func initialModel() Extractor {
	return Extractor{
		choices:  []string{"Node", "Python", "Golang"},
		selected: -1,
	}
}

func (m Extractor) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m Extractor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.cursor
			fmt.Println(m.choices[m.selected])
			return m, tea.Quit

		}
	}
	return m, nil
}

func (m Extractor) View() string {
	s := "What Extractor Would You Like To Use?\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		checked := " "
		if m.selected == i {
			checked = "x"
		}
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}
	s += "\nPress q to quit.\n"
	return s
}

func Execute(service string) {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	var scripts extractors.Scripts
	var path string
	if extractor, ok := m.(Extractor); ok {
		scripts, path, err = extractors.Extract(extractor.choices[extractor.selected])
		if err != nil {
			fmt.Printf("Extraction error: %v", err)
			os.Exit(1)
		}
	}

	switch service {
		case "Github":
			WriteGithubYaml(scripts, path)
	}
}

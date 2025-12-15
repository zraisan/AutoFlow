package executors

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
	"github.com/mzkux/AutoFlow/extractors"

)

type Extractor struct {
	choices  []string
	cursor   int
	selected int
}

type Workflow struct {
	Name string         `yaml:"name"`
	On   On             `yaml:"on"`
	Jobs map[string]Job `yaml:"jobs"`
}

type On struct {
	Push        *PushEvent        `yaml:"push,omitempty"`
	PullRequest *PullRequestEvent `yaml:"pull_request,omitempty"`
}

type PushEvent struct {
	Branches []string `yaml:"branches,omitempty"`
}

type PullRequestEvent struct {
	Branches []string `yaml:"branches,omitempty"`
}

type Job struct {
	RunsOn string `yaml:"runs-on"`
	Steps  []Step `yaml:"steps"`
}

type Step struct {
	Name string            `yaml:"name,omitempty"`
	Uses string            `yaml:"uses,omitempty"`
	Run  string            `yaml:"run,omitempty"`
	With map[string]string `yaml:"with,omitempty"`
}

func initialModel() Extractor {
	return Extractor{
		choices:  []string{"Node", "Python", "Golang"},
		selected: -1,
	}
}

func (m Extractor) Init() tea.Cmd {
	return nil
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

func WriteYaml(data string) string {
	dataz, err := yaml.Marshal(&Workflow{
		Name: "test",
		On: On{
			Push: &PushEvent{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]Job{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps: []Step{
					{Name: "Checkout", Uses: "actions/checkout@v4"},
					{Name: "Install", Run: "npm install"},
					{Name: "Test", Run: "npm test"},
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(dataz))
	return string(dataz)
}

func GithubExecute() {
	p := tea.NewProgram(initialModel())
	m, err := p.Run(); 
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	if extractor, ok := m.(Extractor); ok {
		switch extractor.selected {
			case 0: 
				extractors.NodeExtract()
		}
	}
	return
}

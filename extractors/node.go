package extractors

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	tea "github.com/charmbracelet/bubbletea"
)

type NodeScripts struct {
	Scripts map[string]string `json:"scripts"`
}

type Directory struct {
	value string
}

func ReadFiles(path string) string {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		fmt.Println(entry.Name())
		var structure NodeScripts
		if entry.Name() == "package.json" {
			data, err := os.ReadFile(filepath.Join(path, entry.Name()))
			if err != nil {
				log.Fatal(err)
			}
			err = json.Unmarshal(data, &structure)
			if err != nil {
				log.Fatal(err)
			}
			output, err := json.Marshal(structure)
			fmt.Println(string(output))

			return string(data)
		}
	}

	return ""
}

func initialModel() Directory {
	return Directory{
		value: "",
	}
}

func (m Directory) Init() tea.Cmd {
	return nil
}


func (m Directory) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
			case "enter":
				return m, tea.Quit
			case "ctrl+c":
				return m, tea.Quit
			case "backspace":
				if len(m.value) > 0 {
					m.value = m.value[:len(m.value)-1]
				}
			case "space":
				m.value += " "
			default:
				m.value += msg.String()
		}
	}
	return m, nil
}

func (m Directory) View() string {
	s := "Provide the target path (./)? "

	s += fmt.Sprintf("%s\n", m.value)

	s += "\nPress ctrl+c to quit.\n"

	return s
}

func NodeExtract() {
	p := tea.NewProgram(initialModel())
	m, err := p.Run(); 
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	if directory, ok := m.(Directory); ok {
		result := ReadFiles(directory.value)
		fmt.Println("Result:", result)
	}
	return 
}

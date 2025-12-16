package extractors

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type NodeScripts struct {
	Scripts map[string]string `json:"scripts"`
}

type Directory struct {
	value textinput.Model
	err   error
}

type (
	errMsg error
)

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
	ti := textinput.New()
	ti.Placeholder = "./"
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

	// We handle errors just like any other message
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

func NodeExtract() {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	if directory, ok := m.(Directory); ok {
		result := ReadFiles(directory.value.Value())
		fmt.Println("Result:", result)
	}
	return
}

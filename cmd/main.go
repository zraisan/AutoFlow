package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/mzkux/AutoFlow/executors"
	"github.com/mzkux/AutoFlow/extractors"
	"github.com/mzkux/AutoFlow/types"
	"github.com/spf13/cobra"
)

type Model struct {
	screen    types.Screen
	executor  types.Executor
	extractor types.Extractor
	directory types.Directory
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FC8FF6")).Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FC8FF6")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#D1CAD1"))
	choiceStyles  = map[string]lipgloss.Style{
		"Github":       lipgloss.NewStyle().Foreground(lipgloss.Color("#A855F7")),
		"Gitlab":       lipgloss.NewStyle().Foreground(lipgloss.Color("#F97316")),
		"Azure Devops": lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")),
	}
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "autoflow [path]",
	Short: "A workflow generator",
	Long:  `autoflow is a CI/CD automation tool...`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(initialModel(), tea.WithAltScreen())
		_, err := p.Run()
		if err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}

	},
}

var listCmd = &cobra.Command{
	Use:   "list [directory]",
	Short: "list files",
	Long:  "list all files and subfolders of a given path",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		entries, err := os.ReadDir(args[0])
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range entries {
			fmt.Println(entry.Name())
		}
	},
}

func initialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "./"
	ti.SetValue("./")
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	return Model{
		screen: types.ScreenMain,
		executor: types.Executor{
			Choices: []string{"Github", "Gitlab", "Azure Devops"},
		},
		extractor: types.Extractor{
			Choices:  []string{"Node", "Python", "Golang"},
			Selected: -1,
		},
		directory: types.Directory{
			Value: ti,
			Err:   nil,
		},
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.screen {
	case types.ScreenMain:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.executor.Cursor > 0 {
					m.executor.Cursor--
				}
			case "down", "j":
				if m.executor.Cursor < len(m.executor.Choices)-1 {
					m.executor.Cursor++
				}
			case "enter", " ":
				m.executor.Selected = m.executor.Cursor
				m.screen = types.ScreenExecutor
			}

		}

	case types.ScreenExecutor:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.extractor.Cursor > 0 {
					m.extractor.Cursor--
				}
			case "down", "j":
				if m.extractor.Cursor < len(m.extractor.Choices)-1 {
					m.extractor.Cursor++
				}
			case "enter", " ":
				m.extractor.Selected = m.extractor.Cursor
				m.screen = types.ScreenExtractor
			case "shift+tab":
				m.screen = types.ScreenMain
			}
		}

	case types.ScreenExtractor:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "enter":
				GenerateWorkflow(m, m.directory.Value.Value())
				return m, tea.Quit
			case "up", "k":
				if m.directory.Cursor > 0 {
					m.directory.Cursor--
				}
			case "down", "j":
				if m.directory.Cursor < len(m.directory.Choices)-1 {
					m.directory.Cursor++
				}
			case " ":
				m.directory.Selected = m.directory.Cursor
				m.directory.Value.SetValue(m.directory.Choices[m.directory.Selected])
			case "shift+tab":
				m.screen = types.ScreenExecutor
			}

		case types.ErrMsg:
			m.directory.Err = msg
			return m, nil
		}

		m.directory.Value, cmd = m.directory.Value.Update(msg)
		entries, _ := os.ReadDir(m.directory.Value.Value())
		if len(m.directory.Choices) > 0 {
			m.directory.Choices = m.directory.Choices[:0]
		}
		for _, entry := range entries {
			if entry.IsDir() {
				m.directory.Choices = append(m.directory.Choices, entry.Name())
			}
		}
	}
	return m, cmd
}

func GenerateWorkflow(m Model, directory string) {
	switch m.extractor.Selected {
	case 0:
		scripts, err := extractors.ReadNodeFiles(directory)
		if err != nil {
			fmt.Printf("Extraction error: %v", err)
			os.Exit(1)
		}
		switch m.executor.Selected {
		case 0:
			executors.WriteGithubYaml(scripts, directory)
		}
	}
}

func (m Model) View() string {
	var msg string
	switch m.screen {
	case types.ScreenMain:
		msg += titleStyle.Render("What Executor Would You Like To Use?") + "\n"
		for i, choice := range m.executor.Choices {
			cursor := " "
			if m.executor.Cursor == i {
				cursor = ">"
			}
			checked := " "
			if m.executor.Selected == i {
				checked = "x"
			}
			styledChoice := choice
			if style, ok := choiceStyles[choice]; ok {
				styledChoice = style.Render(choice)
			}
			s := fmt.Sprintf("%s [%s] %s", cursor, checked, styledChoice)
			if m.executor.Cursor == i {
				msg += selectedStyle.Render(s) + "\n"
			} else {
				msg += normalStyle.Render(s) + "\n"
			}
		}
		msg += "\nPress q to quit.\n"

	case types.ScreenExecutor:
		msg += "What Extractor Would You Like To Use?\n\n"
		for i, choice := range m.extractor.Choices {
			cursor := " "
			if m.extractor.Cursor == i {
				cursor = ">"
			}
			checked := " "
			if m.extractor.Selected == i {
				checked = "x"
			}
			msg += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
		}
		msg += "\nPress q to quit.\n"
		return msg

	case types.ScreenExtractor:
		msg += fmt.Sprintf("Project Directory: %s", m.directory.Value.View())
		for i, choice := range m.directory.Choices {
			cursor := ""
			if m.directory.Cursor == i {
				cursor = ">"
			}
			checked := " "
			if m.directory.Selected == i {
				checked = "x"
			}
			msg += fmt.Sprintf("\n %s [%s] %s", cursor, checked, choice)
		}
	}

	return msg
}

func main() {
	rootCmd.AddCommand(listCmd)
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		fmt.Scanln("An Error Ocurred")
		os.Exit(1)
	}
}

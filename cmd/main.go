package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/mzkux/AutoFlow/executors"
	"github.com/spf13/cobra"
)

type Executor struct {
	choices  []string
	cursor   int
	selected int
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
		p := tea.NewProgram(initialModel())
		m, err := p.Run()
		if err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
		if executorModel, ok := m.(Executor); ok {
			executors.Execute(executorModel.choices[executorModel.selected])
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

func initialModel() Executor {
	return Executor{
		choices:  []string{"Github", "Gitlab", "Azure Devops"},
		selected: -1,
	}
}

func (m Executor) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m Executor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Executor) View() string {
	var content string
	content = titleStyle.Render("What Executor Would You Like To Use?") + "\n"
	for i, choice := range m.choices {
		
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		checked := " "
		if m.selected == i {
			checked = "x"
		}
		styledChoice := choice
		if style, ok := choiceStyles[choice]; ok {
			styledChoice = style.Render(choice)
		}
		s := fmt.Sprintf("%s [%s] %s", cursor, checked, styledChoice)
		if m.cursor == i {
			content += selectedStyle.Render(s) + "\n"
		} else {
			content += normalStyle.Render(s) + "\n"
		}
	}
	content += "\nPress q to quit.\n"
	return content
}

func main() {
	rootCmd.AddCommand(listCmd)
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		fmt.Scanln("An Error Ocurred")
		os.Exit(1)
	}
}

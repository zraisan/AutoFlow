package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/mzkux/AutoFlow/executors"
)

type Executor struct {
	choices  []string
	cursor   int
	selected int
}

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
		m, err := p.Run(); 
		if err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
		if executorModel, ok := m.(Executor); ok {
			switch executorModel.selected {
			case 0:
				executors.GithubExecute()

			case 1:
				fmt.Println("Gitlab executor selected")
			case 2:
				fmt.Println("Azure Devops executor selected")
			default:
				fmt.Println("No executor selected")
				return
			}
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
	return nil
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
	s := "What Executor Would You Like To Use?\n\n"
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

func main() {
	rootCmd.AddCommand(listCmd)
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		fmt.Scanln("An Error Ocurred")
		os.Exit(1)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
	landing   types.Landing
	executor  types.Executor
	extractor types.Extractor
	directory types.Directory
	output    string
	width     int
	height    int
}

var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#B0E2FF"))
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#1C6EA4")).Bold(true)
	normalStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#DFDFDF"))
	outputStyle    = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#B0E2FF")).Padding(1, 2)
	containerStyle = lipgloss.NewStyle().Width(70).Align(lipgloss.Left)
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
	dirti := textinput.New()
	dirti.Placeholder = "./"
	dirti.SetValue(".")
	dirti.Focus()
	dirti.CharLimit = 156
	dirti.Width = 20
	lanti := textinput.New()
	lanti.Placeholder = "autoflow"
	lanti.Focus()
	lanti.CharLimit = 156
	lanti.Width = 20
	return Model{
		screen: types.ScreenLanding,
		landing: types.Landing{
			Value: lanti,
			Err:   nil,
		},
		executor: types.Executor{
			Choices:  []string{"Github", "Gitlab", "Azure Devops"},
			Selected: -1,
		},
		extractor: types.Extractor{
			Choices:  []string{"Node", "Python", "Golang"},
			Selected: -1,
		},
		directory: types.Directory{
			Value:      dirti,
			Err:        nil,
			FocusInput: true,
		},
		output: "",
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	switch m.screen {
	case types.ScreenLanding:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.screen = types.ScreenExecutor
			}

		}

		m.landing.Value, cmd = m.landing.Value.Update(msg)

	case types.ScreenExecutor:
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
				m.screen = types.ScreenExtractor
			case "shift+tab":
				m.screen = types.ScreenLanding
			}

		}

	case types.ScreenExtractor:
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
				m.screen = types.ScreenDirectory
				entries, _ := os.ReadDir(m.directory.Value.Value())
				m.directory.Choices = m.directory.Choices[:0]
				for _, entry := range entries {
					if entry.IsDir() {
						m.directory.Choices = append(m.directory.Choices, entry.Name())
					}
				}
			case "shift+tab":
				m.screen = types.ScreenExecutor
			}
		}

	case types.ScreenDirectory:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "tab":
				m.directory.FocusInput = !m.directory.FocusInput
				if m.directory.FocusInput {
					m.directory.Value.Focus()
				} else {
					m.directory.Value.Blur()
				}
			case "shift+tab":
				m.screen = types.ScreenExtractor
				m.directory.FocusInput = true
				m.directory.Value.Focus()
			case "enter":
				if m.directory.FocusInput {
					m.directory.FocusInput = false
				}
				if !m.directory.FocusInput && len(m.directory.Choices) > 0 {
					m.directory.Selected = m.directory.Cursor
					currentPath := m.directory.Value.Value()
					selectedFolder := m.directory.Choices[m.directory.Selected]
					m.directory.Value.SetValue(currentPath + "/" + selectedFolder)
					m.output = GenerateWorkflow(m, m.directory.Value.Value())
					m.screen = types.ScreenResult
				}
			case "up", "k":
				if !m.directory.FocusInput && m.directory.Cursor > 0 {
					m.directory.Cursor--
				}
			case "down", "j":
				if !m.directory.FocusInput && m.directory.Cursor < len(m.directory.Choices)-1 {
					m.directory.Cursor++
				}
			case " ":
				if !m.directory.FocusInput && len(m.directory.Choices) > 0 {
					m.directory.Selected = m.directory.Cursor
					currentPath := m.directory.Value.Value()
					selectedFolder := m.directory.Choices[m.directory.Selected]
					m.directory.Value.SetValue(currentPath + "/" + selectedFolder)
					m.directory.Cursor = 0
				}
			}

		case types.ErrMsg:
			m.directory.Err = msg
			return m, nil
		}

		if m.directory.FocusInput {
			m.directory.Value, cmd = m.directory.Value.Update(msg)
			m.directory.Cursor = 0
		}

		entries, _ := os.ReadDir(m.directory.Value.Value())
		if len(m.directory.Choices) > 0 {
			m.directory.Choices = m.directory.Choices[:0]
		}
		for _, entry := range entries {
			if entry.IsDir() {
				m.directory.Choices = append(m.directory.Choices, entry.Name())
			}
		}

	case types.ScreenResult:
		if len(m.output) < 1 {
			return m, tea.Quit
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", "ctrl+c", "q", "esc":
				return m, tea.Quit

			}
		}
	}
	return m, cmd
}

func GenerateWorkflow(m Model, directory string) string {
	switch m.extractor.Selected {
	case 0: //Node
		scripts, err := extractors.ReadNodeFiles(directory)
		if err != nil {
			fmt.Printf("Extraction error: %v", err)
			os.Exit(1)
		}
		switch m.executor.Selected {
		case 0: //Github
			return executors.WriteGithubYaml(scripts, directory, m.landing.Value.Value())
		}
	}
	return ""
}

func (m Model) View() string {
	var sb strings.Builder

	asciiTitle := []string{
		` █████╗ ██╗   ██╗████████╗ ██████╗ ███████╗██╗      ██████╗ ██╗    ██╗`,
		`██╔══██╗██║   ██║╚══██╔══╝██╔═══██╗██╔════╝██║     ██╔═══██╗██║    ██║`,
		`███████║██║   ██║   ██║   ██║   ██║█████╗  ██║     ██║   ██║██║ █╗ ██║`,
		`██╔══██║██║   ██║   ██║   ██║   ██║██╔══╝  ██║     ██║   ██║██║███╗██║`,
		`██║  ██║╚██████╔╝   ██║   ╚██████╔╝██║     ███████╗╚██████╔╝╚███╔███╔╝`,
		`╚═╝  ╚═╝ ╚═════╝    ╚═╝    ╚═════╝ ╚═╝     ╚══════╝ ╚═════╝  ╚══╝╚══╝ `,
	}
	titleColors := []string{"#B0E2FF", "#87CEEB", "#6BB3D9", "#4A9CC7", "#2E86B5", "#1C6EA4"}

	for i, line := range asciiTitle {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(titleColors[i])).Render(line))
		sb.WriteString("\n")
	}

	sb.WriteString("\n\n")
	switch m.screen {

	case types.ScreenLanding:

		sb.WriteString(titleStyle.Render("Pick a name for your workflow"))
		sb.WriteString("\n\n")
		sb.WriteString(m.landing.Value.View())

	case types.ScreenExecutor:
		sb.WriteString(titleStyle.Render("What Executor Would You Like To Use?"))
		sb.WriteString("\n\n")
		for i, choice := range m.executor.Choices {
			cursor := " "
			if m.executor.Cursor == i {
				cursor = ">"
			}
			checked := " "
			if m.executor.Selected == i {
				checked = "x"
			}
			s := fmt.Sprintf("%s [%s] %s", cursor, checked, choice)
			if m.executor.Cursor == i {
				sb.WriteString(selectedStyle.Render(s))
			} else {
				sb.WriteString(normalStyle.Render(s))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\nPress q to quit.\n")

	case types.ScreenExtractor:
		sb.WriteString(titleStyle.Render("What Extractor Would You Like To Use?"))
		sb.WriteString("\n\n")
		for i, choice := range m.extractor.Choices {
			cursor := " "
			if m.extractor.Cursor == i {
				cursor = ">"
			}
			checked := " "
			if m.extractor.Selected == i {
				checked = "x"
			}
			s := fmt.Sprintf("%s [%s] %s", cursor, checked, choice)
			if m.extractor.Cursor == i {
				sb.WriteString(selectedStyle.Render(s))
			} else {
				sb.WriteString(normalStyle.Render(s))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\nPress q to quit.\n")

	case types.ScreenDirectory:
		fmt.Fprintf(&sb, "Project Directory: %s\n\n", m.directory.Value.View())
		for i, choice := range m.directory.Choices {
			cursor := " "
			if !m.directory.FocusInput && m.directory.Cursor == i {
				cursor = ">"
			}
			fmt.Fprintf(&sb, "%s %s\n", cursor, choice)
		}
		sb.WriteString("\nPress tab to switch focus, enter to select.\n")

	case types.ScreenResult:
		sb.WriteString(titleStyle.Render("Workflow Overview:") + "\n\n")
		fmt.Fprintf(&sb, "%s\n\n", m.directory.Value.Value())
		sb.WriteString(outputStyle.Render(m.output))
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, sb.String())
	}

	content := containerStyle.Render(sb.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func main() {
	rootCmd.AddCommand(listCmd)
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		fmt.Scanln("An Error Ocurred")
		os.Exit(1)
	}
}

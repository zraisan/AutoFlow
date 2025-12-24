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
	"github.com/zraisan/AutoFlow/registry"
	"github.com/spf13/cobra"

	_ "github.com/zraisan/AutoFlow/executors"
	_ "github.com/zraisan/AutoFlow/extractors"
)

type Screen string

const (
	ScreenLanding   Screen = "landing"
	ScreenExecutor  Screen = "executor"
	ScreenExtractor Screen = "extractor"
	ScreenDirectory Screen = "directory"
	ScreenResult    Screen = "result"
)

type ErrMsg error

type Landing struct {
	Value textinput.Model
	Err   error
}

type Executor struct {
	Choices  []string
	Cursor   int
	Selected int
}

type Extractor struct {
	Choices  []string
	Cursor   int
	Selected int
}

type Directory struct {
	Value      textinput.Model
	Choices    []string
	Cursor     int
	Selected   int
	Err        error
	FocusInput bool
}

type Model struct {
	screen    Screen
	landing   Landing
	executor  Executor
	extractor Extractor
	directory Directory
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
		screen: ScreenLanding,
		landing: Landing{
			Value: lanti,
			Err:   nil,
		},
		executor: Executor{
			Choices:  registry.ExecutorNames(),
			Selected: -1,
		},
		extractor: Extractor{
			Choices:  registry.ExtractorNames(),
			Selected: -1,
		},
		directory: Directory{
			Value:      dirti,
			Err:        nil,
			FocusInput: true,
			Selected:   -1,
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
	case ScreenLanding:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.screen = ScreenExecutor
			}

		}

		m.landing.Value, cmd = m.landing.Value.Update(msg)

	case ScreenExecutor:
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
				m.screen = ScreenExtractor
			case "shift+tab":
				m.screen = ScreenLanding
			}

		}

	case ScreenExtractor:
		m.directory.Choices = m.directory.Choices[:0]
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
				m.screen = ScreenDirectory
				entries, _ := os.ReadDir(m.directory.Value.Value())
				for _, entry := range entries {
					if entry.IsDir() {
						m.directory.Choices = append(m.directory.Choices, entry.Name())
					}
				}
			case "shift+tab":
				m.screen = ScreenExecutor
			}
		}

	case ScreenDirectory:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "tab":
				m.directory.FocusInput = !m.directory.FocusInput
			case "shift+tab":
				m.screen = ScreenExtractor
				m.directory.FocusInput = true
				m.directory.Value.Focus()
			case "enter":
				if m.directory.FocusInput {
					m.output = GenerateWorkflow(m, m.directory.Value.Value())
					m.screen = ScreenResult
				}
				if !m.directory.FocusInput && len(m.directory.Choices) > 0 {
					m.directory.Selected = m.directory.Cursor
					currentPath := m.directory.Value.Value()
					selectedFolder := m.directory.Choices[m.directory.Selected]
					m.directory.Value.SetValue(currentPath + "/" + selectedFolder)
					m.output = GenerateWorkflow(m, m.directory.Value.Value())
					m.screen = ScreenResult
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

		case ErrMsg:
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

	case ScreenResult:
		if len(m.output) < 1 {
			return m, tea.Quit
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			case "shift+tab":
				m.screen = ScreenDirectory
			case "enter":
				newModel := initialModel()
				newModel.height = m.height
				newModel.width = m.width
				return newModel, textinput.Blink

			}
		}
	}
	return m, cmd
}

func GenerateWorkflow(m Model, directory string) string {
	extractor := registry.GetExtractor(m.extractor.Selected)
	result, err := extractor.Extract(directory)
	if err != nil {
		fmt.Printf("Extraction error: %v", err)
		os.Exit(1)
	}

	executor := registry.GetExecutor(m.executor.Selected)
	output, err := executor.Generate(result, directory, m.landing.Value.Value())
	if err != nil {
		fmt.Printf("Generation error: %v", err)
		os.Exit(1)
	}

	return output
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

	case ScreenLanding:

		sb.WriteString(titleStyle.Render("Pick a name for your workflow"))
		sb.WriteString("\n\n")
		sb.WriteString(m.landing.Value.View())

	case ScreenExecutor:
		sb.WriteString(titleStyle.Render("What Executor Would You Like To Use?"))
		sb.WriteString("\n\n")
		for i, choice := range m.executor.Choices {
			cursor := " "
			if m.executor.Cursor == i {
				cursor = ">"
			}
			s := fmt.Sprintf("%s %s", cursor, choice)
			if m.executor.Cursor == i {
				sb.WriteString(selectedStyle.Render(s))
			} else {
				sb.WriteString(normalStyle.Render(s))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\nPress q to quit.\n")

	case ScreenExtractor:
		sb.WriteString(titleStyle.Render("What Extractor Would You Like To Use?"))
		sb.WriteString("\n\n")
		for i, choice := range m.extractor.Choices {
			cursor := " "
			if m.extractor.Cursor == i {
				cursor = ">"
			}
			s := fmt.Sprintf("%s %s", cursor, choice)
			if m.extractor.Cursor == i {
				sb.WriteString(selectedStyle.Render(s))
			} else {
				sb.WriteString(normalStyle.Render(s))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\nPress q to quit.\n")

	case ScreenDirectory:
		fmt.Fprintf(&sb, "Project Directory: %s\n\n", m.directory.Value.View())
		for i, choice := range m.directory.Choices {
			cursor := " "
			if !m.directory.FocusInput && m.directory.Cursor == i {
				cursor = ">"
			}
			fmt.Fprintf(&sb, "%s %s\n", cursor, choice)
		}
		sb.WriteString("\nPress tab to switch focus, space to access, enter to select.\n")

	case ScreenResult:
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
		fmt.Println("An Error Ocurred")
		os.Exit(1)
	}
}

package types

import "github.com/charmbracelet/bubbles/textinput"

type Scripts struct {
	Lint    string
	Install string
	Test    string
	Build   string
	Deploy  string
}

type Extractor struct {
	Choices  []string
	Cursor   int
	Selected int
}

type Directory struct {
	Value    textinput.Model
	Choices  []string
	Cursor   int
	Selected int
	Err      error
}

type Executor struct {
	Choices  []string
	Cursor   int
	Selected int
}

type ErrMsg error

type Screen string

const (
	ScreenMain      Screen = "main"
	ScreenExecutor  Screen = "executor"
	ScreenExtractor Screen = "extractor"
)

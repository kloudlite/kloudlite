package input

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func Prompt(o Options) (string, error) {
	i := textinput.New()
	i.Focus()
	i.Prompt = o.Prompt
	i.Placeholder = o.Placeholder
	i.Width = o.Width
	i.PromptStyle = o.PromptStyle.ToLipgloss()
	i.CursorStyle = o.CursorStyle.ToLipgloss()
	i.CharLimit = o.CharLimit
	if o.Password {
		i.EchoMode = textinput.EchoPassword
		i.EchoCharacter = '•'
	}
	p := tea.NewProgram(model{
		textinput: i,
		aborted:   false,
	}, tea.WithOutput(os.Stderr))

	tm, err := p.StartReturningModel()
	if err != nil {
		return "", fmt.Errorf("failed to run input: %w", err)
	}
	m := tm.(model)

	if m.aborted {
		return "", fmt.Errorf("aborted the program")
	}

	return m.textinput.Value(), nil

}

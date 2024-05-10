package spinner

import (
	"github.com/briandowns/spinner"
)

type Sp struct {
	spinner *spinner.Spinner
	show    bool
}

func NewSpinnerClient(message string, show bool) *Sp {
	s := NewSpinner(message)

	return &Sp{
		spinner: s,
		show:    show,
	}
}

func (s *Sp) Start(msg ...string) {
	if len(msg) > 0 {
		s.Update(msg[0])
		return
	}
	s.spinner.Start()
}

func (s *Sp) Stop() {
	s.spinner.Stop()
}

func (s *Sp) Update(message string) {
	s.spinner.Stop()
	s.spinner = NewSpinner(message)
	s.spinner.Start()
}

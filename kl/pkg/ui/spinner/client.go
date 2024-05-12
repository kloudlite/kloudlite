package spinner

import (
	"github.com/briandowns/spinner"
)

type Sp struct {
	spinner *spinner.Spinner
	message string
	show    bool
}

func NewSpinnerClient(message string, show bool) *Sp {
	s := NewSpinner(message)

	return &Sp{
		spinner: s,
		show:    show,
		message: message,
	}
}

func (s *Sp) Start(msg ...string) func() {
	if len(msg) > 0 {
		s.Update(msg[0])
		return func() {
			s.spinner.Stop()
		}
	}

	if s.show {
		s.spinner.Start()
	}

	return func() {
		s.spinner.Stop()
	}
}

func (s *Sp) Stop() {
	s.message = ""
	s.spinner.Stop()
}

func (s *Sp) Update(message string) func() {
	om := s.message
	s.message = message

	s.spinner.Stop()
	s.spinner = NewSpinner(message)

	if s.show {
		s.spinner.Start()
	}

	return func() {
		if om == "" {
			s.spinner.Stop()
			return
		}

		s.message = om
		s.spinner.Stop()
		s.spinner = NewSpinner(om)
		if s.show {
			s.spinner.Start()
		}

	}
}

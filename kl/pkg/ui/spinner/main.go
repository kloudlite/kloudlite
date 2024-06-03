package spinner

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

type Spinner struct {
	spinner *spinner.Spinner
	message string
	verbose bool

	started bool
}

func (s *Spinner) start() {
	s.started = true
	if !s.verbose {
		s.spinner.Start()
		return
	}

	// functions.Logf("[%s] %s", text.Blue("started"), s.message)
}

func (s *Spinner) stop() {
	s.started = false
	if !s.verbose {
		s.spinner.Stop()
		return
	}

	// fmt.Print("\033c")
	// functions.Logf("[%s] %s", text.Blue("stopped"), s.message)
}

func (s *Spinner) Started() bool {
	return s.started
}

func (s *Spinner) Start(msg ...string) func() {
	s.start()

	if len(msg) > 0 && s.message != msg[0] {
		s.UpdateMessage(msg[0])
	}

	return s.stop
}
func (s *Spinner) Stop() {
	s.stop()
}

func (s *Spinner) UpdateMessage(msg string) func() {
	if !s.started {
		s.started = true
		return s.Start(msg)
	}

	om := s.message
	s.message = msg
	s.spinner.Suffix = fmt.Sprintf(" %s...", msg)

	if !s.verbose {
		s.spinner.Restart()
	} else {
		fn.Logf("[%s] %s", text.Blue("+"), s.message)
	}

	return func() {
		if om != "" {
			s.message = om
			s.spinner.Suffix = fmt.Sprintf(" %s...", om)
			if !s.verbose {
				s.spinner.Restart()
			} else {
				fn.Logf("[%s] %s", text.Yellow("-"), om)
			}
		} else {
			fmt.Print("\033c")
		}
	}
}

func newSpinner(msg string, verbose bool) *Spinner {
	sp := spinner.CharSets[11]

	s := spinner.New(sp, 100*time.Millisecond)

	s.Suffix = " loading please wait..."
	if msg != "" {
		s.Suffix = fmt.Sprintf(" %s...", msg)
	}

	return &Spinner{
		spinner: s,
		verbose: verbose,
	}
}

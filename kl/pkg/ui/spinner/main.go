package spinner

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	fn "github.com/kloudlite/kl/pkg/functions"
)

const inProgress = "⏳"
const done = "✅"

type Spinner struct {
	spinner *spinner.Spinner
	message string
	verbose bool
	quiet   bool

	started bool
}

func (s *Spinner) SetVerbose(verbose bool) {
	s.verbose = verbose
}

func (s *Spinner) SetQuiet(quiet bool) {
	s.quiet = quiet
}

func (s *Spinner) start() {
	if s.quiet {
		return
	}

	s.started = true
	if !s.verbose {
		s.spinner.Start()
		return
	}

	fn.Logf("[%s] %s\n", inProgress, s.message)
}

func (s *Spinner) stop() {
	if s.quiet {
		return
	}

	s.started = false
	if !s.verbose {
		s.spinner.Stop()
		return
	}

	fn.Logf("[%s] %s\n", done, s.message)
}

func (s *Spinner) Started() bool {
	return s.started
}

func (s *Spinner) Start(msg ...string) func() {
	if s.quiet {
		return s.stop
	}

	if len(msg) > 0 && s.message != msg[0] {
		s.message = msg[0]
		s.UpdateMessage(msg[0])
	}

	return s.stop
}
func (s *Spinner) Stop() {
	s.stop()
}

func (s *Spinner) UpdateMessage(msg string, verbose ...bool) func() {
	if s.quiet {
		return s.stop
	}

	if len(verbose) > 0 {
		s.verbose = verbose[0]
	}

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
		fn.Logf("[%s] %s\n", inProgress, s.message)
	}

	return func() {
		if om != "" {
			s.message = om
			s.spinner.Suffix = fmt.Sprintf(" %s...", om)
			if !s.verbose {
				s.spinner.Restart()
			} else {
				fn.Logf("[%s] %s\n", done, om)
			}
		} else {
			s.spinner.Stop()
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

package spinner

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func logf(format string, str ...interface{}) {
	if flags.IsQuiet {
		return
	}

	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf(fmt.Sprint(format), str...))
}

type sclient struct {
	spinner *spinner.Spinner
	message []string
	verbose bool
	quiet   bool

	started bool
}

type Spinner interface {
	UpdateMessage(msg string) func()
	Stop()
	Pause()
	Resume()
	SetVerbose(verbose bool)
	SetQuiet(quiet bool)
	IsRunning() bool
}

func (s *sclient) IsRunning() bool {
	return s.started
}

func (s *sclient) pushMessage(msg string) {
	if s.verbose && !s.quiet {
		logf("%s %s\n", text.Bold(text.Green(fmt.Sprintf("+ [%d]", len(s.message)))), msg)

	}

	s.message = append(s.message, msg)
}

func (s *sclient) popMessage() string {
	if len(s.message) == 0 {
		return "please wait..."
	}

	oresp := s.message[len(s.message)-1]
	if s.verbose && !s.quiet {
		logf("%s %s\n", text.Bold(text.Red(fmt.Sprintf("- [%d]", len(s.message)-1))), oresp)
	}

	s.message = s.message[:len(s.message)-1]
	if len(s.message) == 0 {
		return oresp
	}

	return s.message[len(s.message)-1]
}

func (s *sclient) isPopable() bool {
	return len(s.message) > 0
}

func (s *sclient) SetVerbose(verbose bool) {
	s.verbose = verbose
}

func (s *sclient) SetQuiet(quiet bool) {
	s.quiet = quiet
}

func (s *sclient) start() {
	if s.quiet {
		return
	}

	s.started = true
	if !s.verbose {
		s.spinner.Start()
		return
	}
}

func (s *sclient) stop() {
	if s.quiet {
		return
	}

	s.started = false
	if !s.verbose {
		s.spinner.Stop()
		return
	}
}

func (s *sclient) Stop() {
	s.message = []string{}
	s.stop()
}

func (s *sclient) Pause() {
	s.stop()
}

func (s *sclient) Resume() {
	s.start()
}

func (s *sclient) UpdateMessage(msg string) func() {
	if s.quiet {
		return s.stop
	}

	if !s.started {
		s.started = true
		s.start()
	}

	s.pushMessage(msg)
	s.spinner.Suffix = fmt.Sprintf(" %s...", msg)

	if !s.verbose {
		s.spinner.Restart()
	}

	return func() {
		if s.isPopable() {

			s2 := s.popMessage()
			s.spinner.Suffix = fmt.Sprintf(" %s...", s2)

			if !s.verbose {
				s.spinner.Restart()
			}

			if len(s.message) == 0 {
				s.stop()
			}
		} else {
			s.stop()
		}
	}
}

func newSpinner(msg string, verbose bool) Spinner {
	sp := spinner.CharSets[11]

	s := spinner.New(sp, 100*time.Millisecond)

	s.Suffix = " loading please wait..."
	if msg != "" {
		s.Suffix = fmt.Sprintf(" %s...", msg)
	}

	return &sclient{
		spinner: s,
		verbose: verbose,
	}
}

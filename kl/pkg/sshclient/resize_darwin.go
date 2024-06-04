package sshclient

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func handleResize(session *ssh.Session) {
	updateSize := func() {
		width, height, err := term.GetSize(int(os.Stdin.Fd()))
		if err != nil {
			width, height = 100, 30
		}

		session.WindowChange(height, width-2)
	}

	// Set up channel on which to send signal notifications.
	sigCh := make(chan os.Signal, 1)
	// Notify on SIGWINCH (window size change)
	signal.Notify(sigCh, syscall.SIGWINCH)

	// Get the initial terminal size.

	go func() {
		for range sigCh {
			updateSize()
		}
	}()

}

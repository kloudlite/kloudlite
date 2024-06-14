package sshclient

import (
	"golang.org/x/crypto/ssh"
)

func handleResize(session *ssh.Session) {
	// updateSize := func() {
	// 	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	// 	if err != nil {
	// 		width, height = 100, 30
	// 	}

	// 	session.WindowChange(height, width-2)
	// }

}

package console

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
)

func Open(s ssh.Session, cmd *exec.Cmd) error {
	ptyReq, winCh, isPty := s.Pty()
	if !isPty {
		_, err := io.WriteString(s, "No PTY requested.\n")
		if err != nil {
			err = fmt.Errorf("failed to write to console :%w", err)
			err2 := s.Exit(1)
			if err2 != nil {
				return fmt.Errorf("failed to close ssh session: %w", err2)
			}
			return err
		}
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("TERM=%s", ptyReq.Term))
	f, err := pty.Start(cmd)
	if err != nil {
		err = fmt.Errorf("failed to execute command via PTY: command:%s args:%s error:%w", cmd.Path, strings.Join(cmd.Args, " "), err)
		err2 := s.Exit(1)
		if err2 != nil {
			return fmt.Errorf("failed to close ssh session: %w", err2)
		}
		return err
	}

	go func() {
		for win := range winCh {
			err := pty.Setsize(f, &pty.Winsize{X: uint16(win.Width), Y: uint16(win.Height)})
			if err != nil {
				_, err = io.WriteString(s, fmt.Sprintf("failed to set window size to:%d x %d\n", win.Width, win.Height))
			}
		}
	}()

	done := make(chan error)

	go func() {
		_, err = io.Copy(f, s) // stdin
		done <- err
	}()

	go func() {
		_, err = io.Copy(s, f) // stdout
		done <- err
	}()

	// wait till connection is closed
	err = <-done
	return err
}

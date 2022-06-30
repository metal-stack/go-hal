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
	ptyReq, _, isPty := s.Pty()
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

	done := make(chan bool)
	go func() {
		_, err = io.Copy(f, s) // stdin
		_, _ = io.WriteString(s, fmt.Sprintf("failed to copy local stdin to remote:%v\n", err))
		done <- true
	}()

	go func() {
		_, err = io.Copy(s, f) // stdout
		_, _ = io.WriteString(s, fmt.Sprintf("failed to copy local stdout to remote:%v\n", err))
		done <- true
	}()

	// wait till connection is closed
	<-done
	if err != nil {
		return s.Exit(1)
	}

	return s.Exit(0)
}

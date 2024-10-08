package console

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

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

	var winSizeErr, stdinErr, stdoutErr error
	go func() {
		for win := range winCh {
			_, _ = io.WriteString(s, fmt.Sprintf("window size changed width:%d height:%d\n", win.Width, win.Height))
			err := pty.Setsize(f, &pty.Winsize{X: uint16(win.Width), Y: uint16(win.Height)}) // nolint:gosec
			if err != nil {
				winSizeErr = fmt.Errorf("failed to set window size:%w", err)
				_, _ = io.WriteString(s, winSizeErr.Error())
			}
		}
	}()

	done := make(chan bool)

	go func() {
		_, err = io.Copy(f, s) // stdin
		if err != nil && !errors.Is(err, io.EOF) && !strings.HasSuffix(err.Error(), syscall.EIO.Error()) {
			stdinErr = fmt.Errorf("failed to copy remote stdin to local %w", err)
		}
		done <- true
	}()

	go func() {
		_, err = io.Copy(s, f) // stdout
		if err != nil && !errors.Is(err, io.EOF) && !strings.HasSuffix(err.Error(), syscall.EIO.Error()) {
			stdoutErr = fmt.Errorf("failed to copy local stdout to remote: %w", err)
		}
		done <- true
	}()

	// wait till connection is closed
	<-done

	if winSizeErr != nil {
		// FIXME why calling s.Exit(1) in the error message ?
		err = fmt.Errorf("exit ssh session:%s error:%w", s.Exit(1), winSizeErr) // nolint:errorlint
	} else if stdinErr != nil {
		err = fmt.Errorf("exit ssh session:%s error:%w", s.Exit(1), stdinErr) // nolint:errorlint
	} else if stdoutErr != nil {
		err = fmt.Errorf("exit ssh session:%s error:%w", s.Exit(1), stdoutErr) // nolint:errorlint
	} else {
		err2 := s.Exit(0)
		if err2 != nil {
			err = fmt.Errorf("failed to exit ssh session:%w", err2)
		}
	}

	return err
}

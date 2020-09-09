package console

import (
	"fmt"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

func Open(s ssh.Session, cmd *exec.Cmd) error {
	ptyReq, winCh, isPty := s.Pty()
	if !isPty {
		_, err := io.WriteString(s, "No PTY requested.\n")
		if err != nil {
			err = errors.Wrap(err, "failed to write to console")
			err2 := s.Exit(1)
			if err2 != nil {
				return errors.Wrapf(err, "failed to close ssh session: %v", err2)
			}
			return err
		}
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("TERM=%s", ptyReq.Term))
	f, err := pty.Start(cmd)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute command via PTY: command:%s args:%s", cmd.Path, strings.Join(cmd.Args, " "))
		err2 := s.Exit(1)
		if err2 != nil {
			return errors.Wrapf(err, "failed to close ssh session: %v", err2)
		}
		return err
	}

	var winSizeErr, stdinErr, stdoutErr error
	go func() {
		for win := range winCh {
			err := setWinSize(f, win.Width, win.Height)
			if err != nil {
				winSizeErr = errors.Wrapf(err, "failed to set window size:%v", err)
			}
		}
	}()

	done := make(chan bool)

	go func() {
		_, err = io.Copy(f, s) // stdin
		if err != nil && err != io.EOF && !strings.HasSuffix(err.Error(), syscall.EIO.Error()) {
			stdinErr = errors.Wrap(err, "failed to copy remote stdin to local")
		}
		done <- true
	}()

	go func() {
		_, err = io.Copy(s, f) // stdout
		if err != nil && err != io.EOF && !strings.HasSuffix(err.Error(), syscall.EIO.Error()) {
			stdoutErr = errors.Wrap(err, "failed to copy local stdout to remote")
		}
		done <- true
	}()

	// wait till connection is closed
	<-done

	if winSizeErr != nil {
		err = errors.Wrapf(winSizeErr, "exit ssh session:%s", s.Exit(1))
	} else if stdinErr != nil {
		err = errors.Wrapf(stdinErr, "exit ssh session:%s", s.Exit(1))
	} else if stdoutErr != nil {
		err = errors.Wrapf(stdoutErr, "exit ssh session:%s", s.Exit(1))
	} else {
		err2 := s.Exit(0)
		if err2 != nil {
			err = errors.Wrapf(err2, "failed to exit ssh session")
		}
	}

	return err
}

func setWinSize(f *os.File, w, h int) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
	return err
}

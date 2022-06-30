package console

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

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
			err := setWinSize(f, win.Width, win.Height)
			if err != nil {
				_, _ = io.WriteString(s, fmt.Sprintf("failed to set window size from ssh: %dx%d to:%dx%d\n", ptyReq.Window.Width, ptyReq.Window.Height, win.Width, win.Height))
			}
		}
	}()

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
func setWinSize(f *os.File, w, h int) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
	return err
}

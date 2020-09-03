package outband

import (
	"fmt"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

type OutBand struct {
	Redfish  *redfish.APIClient
	IpmiTool ipmi.IpmiTool
	board    *api.Board
	ip       string
	ipmiPort int
	user     string
	password string
}

func New(r *redfish.APIClient, board *api.Board, ip string, ipmiPort int, user, password string) (*OutBand, error) {
	i, err := ipmi.New()
	if err != nil {
		return nil, err
	}

	return &OutBand{
		Redfish:  r,
		IpmiTool: i,
		board:    board,
		ip:       ip,
		ipmiPort: ipmiPort,
		user:     user,
		password: password,
	}, nil
}

func (ob *OutBand) Board() *api.Board {
	return ob.board
}

func (ob *OutBand) IPMIConnection() (string, int, string, string) {
	return ob.ip, ob.ipmiPort, ob.user, ob.password
}

func (ob *OutBand) Goipmi(f func(*ipmi.Client) error) error {
	client, err := ipmi.OpenClientConnection(ob.IPMIConnection())
	if err != nil {
		return err
	}
	defer client.Close()
	return f(client)
}

func (ob *OutBand) Console(s ssh.Session) error {
	_, err := io.WriteString(s, "Exit with '~.'\n")
	if err != nil {
		return errors.Wrap(err, "failed to write to console")
	}
	cmd, err := ob.IpmiTool.NewCommand("sol", "activate")
	if err != nil {
		return err
	}
	return OpenConsole(s, cmd)
}

func OpenConsole(s ssh.Session, cmd *exec.Cmd) error {
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

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		_, err = io.Copy(f, s) // stdin
		if err != nil && err != io.EOF && !strings.HasSuffix(err.Error(), syscall.EIO.Error()) {
			stdinErr = errors.Wrap(err, "failed to copy remote stdin to local")
		}
		wg.Done()
	}()

	go func() {
		_, err = io.Copy(s, f) // stdout
		if err != nil && err != io.EOF && !strings.HasSuffix(err.Error(), syscall.EIO.Error()) {
			stdoutErr = errors.Wrap(err, "failed to copy local stdout to remote")
		}
		wg.Done()
	}()

	// wait till connection is closed
	wg.Wait()

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

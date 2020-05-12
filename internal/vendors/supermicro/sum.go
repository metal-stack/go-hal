package supermicro

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

type sum struct {
	sum      string
	remote   bool
	ip       string
	user     string
	password string
}

func newSum(sumBin string) (*sum, error) {
	_, err := exec.LookPath(sumBin)
	if err != nil {
		return nil, fmt.Errorf("sum binary not present at:%s err:%w", sumBin, err)
	}
	return &sum{
		sum: sumBin,
	}, nil
}

func newRemoteSum(sumBin string, ip, user, password string) (*sum, error) {
	s, err := newSum(sumBin)
	if err != nil {
		return nil, err
	}
	s.remote = true
	s.ip = ip
	s.user = user
	s.password = password
	return s, nil
}

func (s *sum) execute(args ...string) (io.ReadCloser, error) {
	if s.remote {
		args = append(args, "-i", s.ip, "-u", s.user, "-p", s.password)
	}
	cmd := exec.Command(s.sum, args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not initiate sum command to get dmi data from ip:%s, err: %v", s.ip, err)
	}
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("could not start sum command to get dmi data from ip:%s, err: %v", s.ip, err)
	}
	go func() {
		err = cmd.Wait()
		if err != nil {
			log.Printf("wait for sum command failed ip:%s err: %v", s.ip, err)
		}
	}()
	return out, nil
}

func (s *sum) uuidRemote() (string, error) {
	out, err := s.execute("--no_banner", "--no_progress", "--journal_level", "0", "-c", "GetDmiInfo")
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "UUID") {
			return parseUUIDLine(l), nil
		}
	}
	return "", fmt.Errorf("could not find UUID in dmi data for ip:%s", s.ip)
}

const (
	uuidRegex = `([0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12})`
)

var (
	uuidRegexCompiled = regexp.MustCompile(uuidRegex)
)

func parseUUIDLine(l string) string {
	return strings.ToLower(uuidRegexCompiled.FindString(l))
}

package supermicro

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	log "github.com/inconshreveable/log15"
	"github.com/metal-stack/go-hal/internal/kernel"
	"github.com/pkg/errors"
	"golang.org/x/net/html/charset"
)

type boardModel int

const (
	// Bigtwin
	X11DPT_B boardModel = iota
	// S2 Storage
	X11SDV_8C_TP8F
	// S3
	X11DPU
	// N1 Firewall
	X11SDD_8C_F
)

var (
	boardModels = map[string]boardModel{
		// Bigtwin
		"X11DPT-B": X11DPT_B,
		// S2 Storage
		"X11SDV-8C-TP8F": X11SDV_8C_TP8F,
		// S3
		"X11DPU": X11DPU,
		// N1 Firewall
		"X11SDD-8C-F": X11SDD_8C_F,
	}

	// SUM does not complain or fail if more boot options are given than actually available
	uefiBootXMLFragmentTemplates = map[boardModel]string{
		X11DPT_B: `<?xml version="1.0" encoding="ISO-8859-1" standalone="yes"?>
<BiosCfg>
  <Menu name="Boot">
    <Setting name="Boot mode select" selectedOption="UEFI" type="Option"/>
    <Setting name="LEGACY to EFI support" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #1" order="1" selectedOption="UEFI_NETWORK_BOOT_OPTION" type="Option"/>
    <Setting name="Boot Option #2" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #3" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #4" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #5" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #6" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #7" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #8" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #9" order="1" selectedOption="Disabled" type="Option"/>
  </Menu>
  <Menu name="Security">
    <Menu name="SMC Secure Boot Configuration">
      <Setting name="Secure Boot" selectedOption="Enabled" type="Option"/>
    </Menu>
  </Menu>
</BiosCfg>`,
		X11DPU: `<?xml version="1.0" encoding="ISO-8859-1" standalone="yes"?>
<BiosCfg>
  <Menu name="Boot">
    <Setting name="Boot mode select" selectedOption="UEFI" type="Option"/>
    <Setting name="LEGACY to EFI support" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #1" order="1" selectedOption="UEFI_NETWORK_BOOT_OPTION" type="Option"/>
    <Setting name="Boot Option #2" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #3" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #4" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #5" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #6" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #7" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #8" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #9" order="1" selectedOption="Disabled" type="Option"/>
  </Menu>
  <Menu name="Security">
    <Menu name="SMC Secure Boot Configuration">
      <Setting name="Secure Boot" selectedOption="Enabled" type="Option"/>
    </Menu>
  </Menu>
</BiosCfg>`,
		X11SDV_8C_TP8F: `<?xml version="1.0" encoding="ISO-8859-1" standalone="yes"?>
<BiosCfg>
  <Menu name="Boot">
    <Setting name="Boot mode select" selectedOption="UEFI" type="Option"/>
    <Setting name="Legacy to EFI support" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #1" selectedOption="UEFI_NETWORK_BOOT_OPTION" type="Option"/>
    <Setting name="UEFI Boot Option #2" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #3" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #4" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #5" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #6" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #7" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #8" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #9" selectedOption="Disabled" type="Option"/>
  </Menu>
  <Menu name="Security">
    <Menu name="SMC Secure Boot Configuration">
      <Setting name="Secure Boot" selectedOption="Enabled" type="Option"/>
    </Menu>
  </Menu>
</BiosCfg>`,
		X11SDD_8C_F: `<?xml version="1.0" encoding="ISO-8859-1" standalone="yes"?>
<BiosCfg>
  <Menu name="Boot">
    <Setting name="Boot mode select" selectedOption="UEFI" type="Option"/>
    <Setting name="Legacy To EFI Support" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #1" selectedOption="UEFI_NETWORK_BOOT_OPTION" type="Option"/>
    <Setting name="Boot Option #2" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #3" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #4" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #5" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #6" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #7" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #8" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #9" selectedOption="Disabled" type="Option"/>
  </Menu>
</BiosCfg>`,
	}

	bootOrderXMLFragmentTemplates = map[boardModel]string{
		X11DPT_B: `<?xml version="1.0" encoding="ISO-8859-1" standalone="yes"?>
<BiosCfg>
  <Menu name="Boot">
    <Setting name="Boot Option #1" order="1" selectedOption="UEFI Hard Disk:BOOTLOADER_ID" type="Option"/>
    <Setting name="Boot Option #2" order="1" selectedOption="UEFI_NETWORK_BOOT_OPTION" type="Option"/>
    <Setting name="Boot Option #3" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #4" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #5" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #6" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #7" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #8" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #9" order="1" selectedOption="Disabled" type="Option"/>
  </Menu>
</BiosCfg>`,
		X11DPU: `<?xml version="1.0" encoding="ISO-8859-1" standalone="yes"?>
<BiosCfg>
  <Menu name="Boot">
    <Setting name="Boot Option #1" order="1" selectedOption="UEFI Hard Disk:BOOTLOADER_ID" type="Option"/>
    <Setting name="Boot Option #2" order="1" selectedOption="UEFI_NETWORK_BOOT_OPTION" type="Option"/>
    <Setting name="Boot Option #3" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #4" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #5" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #6" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #7" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #8" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #9" order="1" selectedOption="Disabled" type="Option"/>
  </Menu>
</BiosCfg>`,
		X11SDD_8C_F: `<?xml version="1.0" encoding="ISO-8859-1" standalone="yes"?>
<BiosCfg>
  <Menu name="Boot">
    <Setting name="Boot Option #1" order="1" selectedOption="UEFI Hard Disk:BOOTLOADER_ID" type="Option"/>
    <Setting name="Boot Option #2" order="1" selectedOption="UEFI_NETWORK_BOOT_OPTION" type="Option"/>
    <Setting name="Boot Option #3" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #4" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #5" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #6" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #7" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #8" order="1" selectedOption="Disabled" type="Option"/>
    <Setting name="Boot Option #9" order="1" selectedOption="Disabled" type="Option"/>
  </Menu>
</BiosCfg>`,
		X11SDV_8C_TP8F: `<?xml version="1.0" encoding="ISO-8859-1" standalone="yes"?>
<BiosCfg>
  <Menu name="Boot">
    <Setting name="UEFI Boot Option #1" selectedOption="UEFI Hard Disk:BOOTLOADER_ID" type="Option"/>
    <Setting name="UEFI Boot Option #2" selectedOption="UEFI_NETWORK_BOOT_OPTION" type="Option"/>
    <Setting name="UEFI Boot Option #3" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #4" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #5" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #6" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #7" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #8" selectedOption="Disabled" type="Option"/>
    <Setting name="UEFI Boot Option #9" selectedOption="Disabled" type="Option"/>
  </Menu>
</BiosCfg>`,
	}
)

type Menu struct {
	XMLName  xml.Name `xml:"Menu"`
	Name     string   `xml:"name,attr"`
	Settings []struct {
		XMLName        xml.Name `xml:"Setting"`
		Name           string   `xml:"name,attr"`
		Order          string   `xml:"order,attr,omitempty"`
		SelectedOption string   `xml:"selectedOption,attr"`
	} `xml:"Setting"`
	Menus []Menu `xml:"Menu"`
}

type BiosCfg struct {
	XMLName xml.Name `xml:"BiosCfg"`
	Menus   []Menu   `xml:"Menu"`
}

type sum struct {
	binary   string
	remote   bool
	ip       string
	user     string
	password string

	bootloaderID          string
	biosCfgXML            string
	biosCfg               BiosCfg
	boardModel            boardModel
	boardName             string
	uefiNetworkBootOption string
	secureBootEnabled     bool
}

func newSum(sumBin, boardName string) (*sum, error) {
	_, err := exec.LookPath(sumBin)
	if err != nil {
		return nil, fmt.Errorf("sum binary not present at:%s err:%w", sumBin, err)
	}
	sum := &sum{
		binary:    sumBin,
		boardName: boardName,
	}
	bm, ok := boardModels[boardName]
	if ok {
		sum.boardModel = bm
	} else {
		sum.boardModel = X11DPT_B
	}
	return sum, nil
}

func NewRemoteSum(sumBin, boardName string, ip, user, password string) (*sum, error) {
	s, err := newSum(sumBin, boardName)
	if err != nil {
		return nil, err
	}
	s.remote = true
	s.ip = ip
	s.user = user
	s.password = password
	return s, nil
}

// ConfigureBIOS updates BIOS to UEFI boot and disables CSM-module if required.
// If returns whether machine needs to be rebooted or not.
func (s *sum) ConfigureBIOS() (bool, error) {
	firmware := kernel.Firmware()
	log.Info("firmware", "is", firmware, "board", s.boardModel, "boardname", s.boardName)

	// We must not configure the Bios if UEFI is already activated and the board is one of the following.
	if firmware == kernel.EFI && (s.boardModel == X11SDV_8C_TP8F || s.boardModel == X11SDD_8C_F) {
		return false, nil
	}

	err := s.prepare()
	if err != nil {
		return false, err
	}
	log.Info("firmware", "is", firmware, "board", s.boardModel, "boardname", s.boardName, "secureboot", s.secureBootEnabled)

	// Secureboot can be set for specific bigtwins, called CSM Support in the bios
	// This is so far only possible on these machines, detection requires sum call which downloads the bios.xml
	if firmware == kernel.EFI && s.secureBootEnabled {
		return false, nil
	}

	fragment := uefiBootXMLFragmentTemplates[s.boardModel]
	fragment = strings.ReplaceAll(fragment, "UEFI_NETWORK_BOOT_OPTION", s.uefiNetworkBootOption)

	return true, s.changeBiosCfg(fragment)
}

// EnsureBootOrder ensures BIOS boot order so that boot from the given allocated OS image is attempted before PXE boot.
func (s *sum) EnsureBootOrder(bootloaderID string) error {
	s.bootloaderID = bootloaderID

	err := s.prepare()
	if err != nil {
		log.Warn("BIOS updates for this machine type are intentionally not supported, skipping EnsureBootOrder", "error", err)
		return nil
	}

	ok := s.bootOrderProperlySet()
	if ok {
		log.Info("sum", "message", "boot order is already configured")
		return nil
	}

	fragment := bootOrderXMLFragmentTemplates[s.boardModel]
	fragment = strings.ReplaceAll(fragment, "BOOTLOADER_ID", s.bootloaderID)
	fragment = strings.ReplaceAll(fragment, "UEFI_NETWORK_BOOT_OPTION", s.uefiNetworkBootOption)

	return s.changeBiosCfg(fragment)
}

func (s *sum) prepare() error {
	err := s.getCurrentBiosCfg()
	if err != nil {
		return err
	}

	err = s.unmarshalBiosCfg()
	if err != nil {
		return errors.Wrapf(err, "unable to unmarshal BIOS configuration:\n%s", s.biosCfgXML)
	}

	s.determineSecureBoot()

	return s.findUEFINetworkBootOption()
}

func (s *sum) getCurrentBiosCfg() error {
	biosCfgXML := "biosCfg.xml"
	_ = os.Remove(biosCfgXML)

	err := s.execute("-c", "GetCurrentBiosCfg", "--file", biosCfgXML)
	if err != nil {
		return errors.Wrapf(err, "unable to get BIOS configuration via:%s -c GetCurrentBiosCfg --file %s", s.binary, biosCfgXML)
	}

	bb, err := ioutil.ReadFile(biosCfgXML)
	if err != nil {
		return errors.Wrapf(err, "unable to read file:%s", biosCfgXML)
	}

	s.biosCfgXML = string(bb)
	return nil
}

func (s *sum) determineSecureBoot() {
	if s.boardModel == X11SDV_8C_TP8F || s.boardModel == X11SDD_8C_F { // secure boot option is not available in S2 BIOS
		return
	}
	for _, menu := range s.biosCfg.Menus {
		if menu.Name != "Security" {
			continue
		}
		for _, m := range menu.Menus {
			if m.Name != "SMC Secure Boot Configuration" {
				continue
			}
			for _, setting := range m.Settings {
				if setting.Name == "Secure Boot" {
					s.secureBootEnabled = setting.SelectedOption == "Enabled"
					return
				}
			}
		}
	}
}

func (s *sum) unmarshalBiosCfg() error {
	s.biosCfg = BiosCfg{}
	decoder := xml.NewDecoder(strings.NewReader(s.biosCfgXML))
	decoder.CharsetReader = charset.NewReaderLabel
	return decoder.Decode(&s.biosCfg)
}

func (s *sum) findUEFINetworkBootOption() error {
	for _, menu := range s.biosCfg.Menus {
		if menu.Name != "Boot" {
			continue
		}
		for _, setting := range menu.Settings {
			if strings.Contains(setting.SelectedOption, "UEFI Network") {
				s.uefiNetworkBootOption = setting.SelectedOption
				return nil
			}
		}
	}

	return fmt.Errorf("cannot find PXE boot option in BIOS configuration:\n%s\n", s.biosCfgXML)
}

func (s *sum) bootOrderProperlySet() bool {
	if !s.checkBootOptionAt(1, s.bootloaderID) {
		return false
	}
	if !s.checkBootOptionAt(2, s.uefiNetworkBootOption) {
		return false
	}
	for i := 2; i <= 9; i++ {
		if !s.checkBootOptionAt(i, "Disabled") {
			return false
		}
	}
	return true
}

func (s *sum) checkBootOptionAt(index int, bootOption string) bool {
	for _, menu := range s.biosCfg.Menus {
		if menu.Name != "Boot" {
			continue
		}
		for _, setting := range menu.Settings {
			switch s.boardModel {
			case X11DPT_B, X11DPU, X11SDD_8C_F:
				if setting.Order != "1" {
					continue
				}
				if setting.Name != fmt.Sprintf("Boot Option #%d", index) {
					continue
				}
			case X11SDV_8C_TP8F:
				if setting.Name != fmt.Sprintf("UEFI Boot Option #%d", index) {
					continue
				}
			}

			return strings.Contains(setting.SelectedOption, bootOption)
		}
	}

	return false
}

func (s *sum) changeBiosCfg(fragment string) error {
	biosCfgUpdateXML := "biosCfgUpdate.xml"
	err := ioutil.WriteFile(biosCfgUpdateXML, []byte(fragment), 0600)
	if err != nil {
		return err
	}

	return s.execute("-c", "ChangeBiosCfg", "--file", biosCfgUpdateXML)
}

func (s *sum) execute(args ...string) error {
	if s.remote {
		args = append(args, "-i", s.ip, "-u", s.user, "-p", s.password)
	}
	// #nosec G204
	cmd := exec.Command(s.binary, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid:    uint32(0),
			Gid:    uint32(0),
			Groups: []uint32{0},
		},
	}
	return cmd.Run()
}

func (s *sum) executeAsync(args ...string) (io.ReadCloser, error) {
	if s.remote {
		args = append(args, "-i", s.ip, "-u", s.user, "-p", s.password)
	}
	// #nosec G204
	cmd := exec.Command(s.binary, args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not initiate sum command to get dmi data from ip:%s, err: %w", s.ip, err)
	}
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("could not start sum command to get dmi data from ip:%s, err: %w", s.ip, err)
	}
	go func() {
		err = cmd.Wait()
		if err != nil {
			log.Info("wait for sum command failed ip", "ip", s.ip, "error", err)
		}
	}()
	return out, nil
}

func (s *sum) uuidRemote() (string, error) {
	out, err := s.executeAsync("--no_banner", "--no_progress", "--journal_level", "0", "-c", "GetDmiInfo")
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

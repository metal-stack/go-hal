package ipmi

import (
	"reflect"
	"strings"
	"testing"

	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/stretchr/testify/require"
)

// Output of root@ipmitest:~# ipmitool lan print
const lanPrint = `
Set in Progress         : Set Complete
Auth Type Support       : NONE MD2 MD5 PASSWORD
Auth Type Enable        : Callback : MD2 MD5 PASSWORD
                        : User     : MD2 MD5 PASSWORD
                        : Operator : MD2 MD5 PASSWORD
                        : Admin    : MD2 MD5 PASSWORD
                        : OEM      : MD2 MD5 PASSWORD
IP Address Source       : Static Address
IP Address              : 10.248.36.246
Subnet Mask             : 255.255.252.0
MAC Address             : 0c:c4:7a:ed:3e:27
SNMP Community String   : public
IP Header               : TTL=0x00 Flags=0x00 Precedence=0x00 TOS=0x00
BMC ARP Control         : ARP Responses Enabled, Gratuitous ARP Disabled
Default Gateway IP      : 10.248.36.1
Default Gateway MAC     : 30:b6:4f:c3:a0:c1
Backup Gateway IP       : 0.0.0.0
Backup Gateway MAC      : 00:00:00:00:00:00
802.1q VLAN ID          : Disabled
802.1q VLAN Priority    : 0
RMCP+ Cipher Suites     : 1,2,3,6,7,8,11,12
Cipher Suite Priv Max   : XaaaXXaaaXXaaXX
                        :     X=Cipher Suite Unused
                        :     c=CALLBACK
                        :     u=USER
                        :     o=OPERATOR
                        :     a=ADMIN
                        :     O=OEM
Bad Password Threshold  : Not Available
`
const lanPrint2 = "Set in Progress         : Set Complete\nAuth Type Support       : NONE MD2 MD5 PASSWORD \nAuth Type Enable        : Callback : MD2 MD5 PASSWORD \n                        : User     : MD2 MD5 PASSWORD \n                        : Operator : MD2 MD5 PASSWORD \n                        : Admin    : MD2 MD5 PASSWORD \n                        : OEM      : MD2 MD5 PASSWORD \nIP Address Source       : DHCP Address\nIP Address              : 192.168.2.53\nSubnet Mask             : 255.255.255.0\nMAC Address             : ac:1f:6b:73:c9:f0\nSNMP Community String   : public\nIP Header               : TTL=0x00 Flags=0x00 Precedence=0x00 TOS=0x00\nBMC ARP Control         : ARP Responses Enabled, Gratuitous ARP Disabled\nDefault Gateway IP      : 192.168.2.1\nDefault Gateway MAC     : 00:00:00:00:00:00\nBackup Gateway IP       : 0.0.0.0\nBackup Gateway MAC      : 00:00:00:00:00:00\n802.1q VLAN ID          : Disabled\n802.1q VLAN Priority    : 0\nRMCP+ Cipher Suites     : 1,2,3,6,7,8,11,12\nCipher Suite Priv Max   : XaaaXXaaaXXaaXX\n                        :     X=Cipher Suite Unused\n                        :     c=CALLBACK\n                        :     u=USER\n                        :     o=OPERATOR\n                        :     a=ADMIN\n                        :     O=OEM\nBad Password Threshold  : 0\nInvalid password disable: no\nAttempt Count Reset Int.: 0\nUser Lockout Interval   : 0\n"

const bmcInfo = `
Device ID                 : 32
Device Revision           : 1
Firmware Revision         : 7.00
IPMI Version              : 2.0
Manufacturer ID           : 674
Manufacturer Name         : Dell Inc.
Product ID                : 256 (0x0100)
Product Name              : Unknown (0x100)
Device Available          : yes
Provides Device SDRs      : yes
Additional Device Support :
    Sensor Device
    SDR Repository Device
    SEL Device
    FRU Inventory Device
    IPMB Event Receiver
    Bridge
    Chassis Device
Aux Firmware Rev Info     : 
    0x00
    0x03
    0x00
    0xb7
`

const fru = `
FRU Device Description : Builtin FRU Device (ID 0)                                                                                                                                                                                              
 Board Mfg Date        : Fri 15 Dec 2017 09:34:00 PM CET CET                                                                                                                                                                                    
 Board Mfg             : DELL                                                                                                                                                                                                                   
 Board Product         : PowerEdge R540                                                                                                                                                                                                         
 Board Serial          : CNFCP007CE00WG                                                                                                                                                                                                         
 Board Part Number     : 0N28XXA02                                                                                                                                                                                                              
 Product Manufacturer  : DELL                                                                                                                                                                                                                   
 Product Name          : PowerEdge R540                                                                                                                                                                                                         
 Product Version       : 01                                                                                                                                                                                                                     
 Product Serial        : DR8L3N2                                                                                                                                                                                                                
 Product Asset Tag     :                                                                                                                                                                                                                        
                                                                                                                                                                                                                                                
FRU Device Description : PS1 (ID 1)                                                                                                                                                                                                             
 Board Mfg Date        : Tue 14 Nov 2017 12:56:00 PM CET CET                                                                                                                                                                                    
 Board Mfg             : DELL                                                                                                                                                                                                                   
 Board Product         : PWR SPLY,750W,RDNT,DELTA                                                                                                                                                                                               
 Board Serial          : CNDED007BE9YEX                                                                                                                                                                                                         
 Board Part Number     : 05RHVVA02                                                                                                                                                                                                              
                                                                                                                                                                                                                                                
FRU Device Description : PS2 (ID 2)                                                                                                                                                                                                             
 Board Mfg Date        : Tue 14 Nov 2017 12:46:00 PM CET CET                                                                                                                                                                                    
 Board Mfg             : DELL                                                                                                                                                                                                                   
 Board Product         : PWR SPLY,750W,RDNT,DELTA                                                                                                                                                                                               
 Board Serial          : CNDED007BE9YEV                                                                                                                                                                                                         
 Board Part Number     : 05RHVVA02                                                                                                                                                                                                              
                                                                                                                                                                                                                                                
FRU Device Description : BP1 (ID 13)                                                                                                                                                                                                            
 Board Mfg Date        : Sat 29 Jul 2017 05:28:00 PM CEST CEST                                                                                                                                                                                  
 Board Mfg             : DELL                                                                                                                                                                                                                   
 Board Product         : DRIVE BACKPLANE                                                                                                                                                                                                        
 Board Serial          : CNIVC0077T0039                                                                                                                                                                                                         
 Board Part Number     : 08N0NGA00                                                                                                                                                                                                              

FRU Device Description : NDC (ID 4)                                                                                     
 Device not present (Timeout)                                                                                           

FRU Device Description : PERC1 (ID 10)                                                                                  
 Board Mfg Date        : Sun 24 Dec 2017 05:27:00 AM CET CET                                                            
 Board Mfg             : DELL                                                                                           
 Board Product         : Dell Storage Cntlr. H730P - Adp.                                                               
 Board Serial          : CNFCP007CL004R                                                                                 
 Board Part Number     : 0J14DCA03                                                                                      

FRU Device Description : OEM fru (ID 17)                                                                                

FRU Device Description : BP0 (ID 12)                                                                                    
 Device not present (Timeout)                                                                                           

FRU Device Description : OCP Mezz (ID 7)                                                                                
 Device not present (Timeout)                                                                                           

FRU Device Description : PERC2 (ID 11)                                                                                  
 Device not present (Parameter out of range)                                                                            
`

func Test_getLanConfig(t *testing.T) {
	tests := []struct {
		name      string
		cmdOutput string
		want      map[string]string
	}{
		{
			name:      "simple",
			cmdOutput: lanPrint,
			want: map[string]string{
				"IP Address":  "10.248.36.246",
				"Subnet Mask": "255.255.252.0",
				"MAC Address": "0c:c4:7a:ed:3e:27",
			},
		},
		{
			name:      "from real",
			cmdOutput: lanPrint2,
			want: map[string]string{
				"IP Address":  "192.168.2.53",
				"Subnet Mask": "255.255.255.0",
				"MAC Address": "ac:1f:6b:73:c9:f0",
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		i := Ipmitool{log: logger.New()}
		t.Run(tt.name, func(t *testing.T) {
			got := i.output2Map(tt.cmdOutput)
			for key, value := range tt.want {
				if got[key] != value {
					t.Errorf("getLanConfig() = %v, want %v", got[key], value)
				}
			}
		})
	}
}

func Test_getBmcInfog(t *testing.T) {
	tests := []struct {
		name      string
		cmdOutput string
		want      *BMCInfo
	}{
		{
			name:      "simple",
			cmdOutput: bmcInfo,
			want: &BMCInfo{
				FirmwareRevision: "7.00",
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		i := Ipmitool{log: logger.New()}
		t.Run(tt.name, func(t *testing.T) {
			got := i.output2Map(tt.cmdOutput)
			b := &BMCInfo{}
			from(b, got)
			require.Equal(t, tt.want, b)
		})
	}
}

func Test_getFru(t *testing.T) {
	tests := []struct {
		name      string
		cmdOutput string
		want      *Fru
	}{
		{
			name:      "simple",
			cmdOutput: fru,
			want: &Fru{
				BoardMfg:            "DELL",
				BoardPartNumber:     "0J14DCA03",
				ProductManufacturer: "DELL",
				ProductSerial:       "DR8L3N2",
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		i := Ipmitool{log: logger.New()}
		t.Run(tt.name, func(t *testing.T) {
			got := i.output2Map(tt.cmdOutput)
			b := &Fru{}
			from(b, got)
			require.Equal(t, tt.want, b)
		})
	}
}

func TestGetLanConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    LanConfig
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			i := &Ipmitool{command: "/bin/true"}
			got, err := i.GetLanConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLanConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLanConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLanConfig_From(t *testing.T) {
	i := Ipmitool{log: logger.New()}

	type fields struct {
		IP  string
		Mac string
	}
	tests := []struct {
		name   string
		fields fields
		input  map[string]string
	}{
		{
			name: "simple",
			fields: fields{
				IP:  "192.168.2.53",
				Mac: "ac:1f:6b:73:c9:f0",
			},
			input: i.output2Map(lanPrint2),
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			c := LanConfig{}
			from(&c, tt.input)
			if c.IP != tt.fields.IP {
				t.Errorf("IP is not as expected")
			}
			if c.Mac != tt.fields.Mac {
				t.Errorf("Mac is not as expected")
			}
		})
	}
}

func TestIpmitool_Run(t *testing.T) {
	type fields struct {
		command  string
		ip       string
		port     int
		user     string
		password string
		outband  bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "verify outband args are added before user args",
			fields: fields{
				command:  "echo",
				ip:       "1.2.3.4",
				port:     623,
				user:     "user",
				password: "password",
				outband:  true,
			},
			args:    []string{"lan", "print"},
			want:    "-I lanplus -H 1.2.3.4 -p 623 -U user -E lan print",
			wantErr: false,
		},
		{
			name: "verify inband has no outband args",
			fields: fields{
				command:  "echo",
				ip:       "1.2.3.4",
				port:     623,
				user:     "user",
				password: "password",
				outband:  false,
			},
			args:    []string{"lan", "print"},
			want:    "lan print",
			wantErr: false,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			i := &Ipmitool{
				command:  tt.fields.command,
				ip:       tt.fields.ip,
				port:     tt.fields.port,
				user:     tt.fields.user,
				password: tt.fields.password,
				outband:  tt.fields.outband,
			}
			got, err := i.Run(tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ipmitool.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("Ipmitool.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

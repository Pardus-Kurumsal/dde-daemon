package miracast

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"pkg.deepin.io/lib/strv"
	"strings"
)

type WirelessInfo struct {
	Name       string
	MacAddress string
	Ciphers    []string
	IFCModes   []string
	Commands   []string
}
type WirelessInfos []*WirelessInfo

func ListWirelessInfo() (WirelessInfos, error) {
	var envPath = os.Getenv("PATH")
	os.Setenv("PATH", "/sbin:"+envPath)
	defer os.Setenv("PATH", envPath)
	outputs, err := exec.Command("/bin/sh", "-c",
		"exec iw list").CombinedOutput()
	if err != nil {
		return nil, err
	}
	return parseIwOutputs(outputs), nil
}

func (infos WirelessInfos) ListMiracastDevice() WirelessInfos {
	var ret WirelessInfos
	for _, info := range infos {
		if !info.SupportedMiracast() {
			continue
		}
		ret = append(ret, info)
	}
	return ret
}

func (infos WirelessInfos) ListHotspotDevice() WirelessInfos {
	var ret WirelessInfos
	for _, info := range infos {
		if !info.SupportedHotspot() {
			continue
		}
		ret = append(ret, info)
	}
	return ret
}

func (infos WirelessInfos) Get(macAddress string) *WirelessInfo {
	for _, info := range infos {
		if info.MacAddress == macAddress {
			return info
		}
	}
	return nil
}

func (info *WirelessInfo) SupportedHotspot() bool {
	return strv.Strv(info.IFCModes).Contains("AP")
}

func (info *WirelessInfo) SupportedMiracast() bool {
	list := strv.Strv(info.IFCModes)
	return list.Contains("P2P-client") &&
		list.Contains("P2P-GO")
	// list.Contains("P2P-device")
}

func debugWirelessInfos() {
	infos, err := ListWirelessInfo()
	if err != nil {
		fmt.Println("Failed to list wireless devices:", err)
		return
	}

	for _, info := range infos {
		fmt.Println(info.Name)
		fmt.Println("\tMac Address\t:", info.MacAddress)
		fmt.Println("\tCiphers\t:", info.Ciphers)
		fmt.Println("\tInterface Modes\t:", info.IFCModes)
		fmt.Println("\tCommands\t:", info.Commands)
	}
}

func parseIwOutputs(contents []byte) WirelessInfos {
	lines := strings.Split(string(contents), "\n")
	length := len(lines)
	var infos WirelessInfos
	for i := 0; i < length; {
		line := lines[i]
		if len(line) == 0 {
			i += 1
			continue
		}

		line = strings.TrimSpace(line)
		if strings.Contains(line, "Wiphy phy") {
			infos = append(infos, new(WirelessInfo))
			name := strings.Split(line, "Wiphy ")[1]
			infos[len(infos)-1].Name = name
			infos[len(infos)-1].MacAddress = getMacAddressByFile(macAddressFile(name))
			i += 1
			continue
		}

		if strings.Contains(line, "Supported Ciphers:") {
			i, infos[len(infos)-1].Ciphers = getValues(i+1, &lines)
			continue
		}

		if strings.Contains(line, "Supported interface modes:") {
			i, infos[len(infos)-1].IFCModes = getValues(i+1, &lines)
			continue
		}

		if strings.Contains(line, "Supported commands:") {
			i, infos[len(infos)-1].Commands = getValues(i+1, &lines)
			continue
		}

		i += 1
	}
	return infos
}

func getValues(idx int, lines *[]string) (int, []string) {
	var values []string
	length := len(*lines)
	for ; idx < length; idx++ {
		value := strings.TrimSpace((*lines)[idx])
		if value[0] != '*' {
			break
		}
		values = append(values, strings.Split(value, "* ")[1])
	}
	return idx, values
}

func macAddressFile(name string) string {
	return "/sys/class/ieee80211/" + name + "/macaddress"
}

func getMacAddressByFile(file string) string {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(contents))
}

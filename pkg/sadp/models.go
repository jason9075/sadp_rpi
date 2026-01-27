package sadp

import (
	"encoding/xml"
)

// Probe 定義發送給設備的 XML 結構
type Probe struct {
	XMLName xml.Name `xml:"Probe"`
	Uuid    string   `xml:"Uuid"`
	Types   string   `xml:"Types"`
}

// Device 接收設備回傳的解析結果
type Device struct {
	IPv4Address     string `xml:"IPv4Address" json:"ipv4Address"`
	IPv4SubnetMask  string `xml:"IPv4SubnetMask" json:"ipv4SubnetMask"`
	IPv4Gateway     string `xml:"IPv4Gateway" json:"ipv4Gateway"`
	IPv4Port        int    `xml:"IPv4Port" json:"ipv4Port"`
	HttpPort        int    `xml:"HttpPort" json:"httpPort"`
	MAC             string `xml:"MAC" json:"mac"`
	DeviceID        string `xml:"DeviceID" json:"deviceId"`
	DeviceDescription string `xml:"DeviceDescription" json:"deviceDescription"`
	DeviceSN        string `xml:"DeviceSN" json:"deviceSN"`
	SoftwareVersion string `xml:"SoftwareVersion" json:"softwareVersion"`
	Activated       bool   `xml:"Activated" json:"activated"`
}

// ProbeMatch 定義設備回傳的 XML 包裝結構
type ProbeMatch struct {
	XMLName xml.Name `xml:"ProbeMatch"`
	Device
}

package main

// Config contains hardcoded configurations for the ELP test application
var Config = struct {
	NumDevices  int
	IntervalSec int
	TargetIP    string
	TargetPort  int
	BaseMAC     string
	APMacs      [5]string
}{
	NumDevices:  5,                   // Number of devices
	IntervalSec: 2,                   // Interval in seconds
	TargetIP:    "127.0.0.1",         // Target IP
	TargetPort:  8080,                // Target port
	BaseMAC:     "C4:CB:6B:23:00:01", // Starting MAC address
}

package main

import (
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AlpAdditionalInfo struct {
	Charging          bool
	InitialScan       bool
	MessagePending    bool
	SafetySwitch      bool
	SupplementaryScan bool
}

type AlpMac struct {
	AsHexString string
	BytesArray  [6]byte
}

type AlpHeader struct {
	Length    uint16
	Magic     string
	RawBytes  []byte
	Sequence  uint16
	Timestamp uint32
	Type      uint16
	Version   uint16
}

type AlpDeviceData struct {
	RawBytes        []byte
	DeviceMac       *AlpMac
	DeviceType      uint16
	MenuItem        uint8
	BatteryLevel    uint8
	MessageID       uint8
	AddnlInfo       *AlpAdditionalInfo
	RequestAck      bool
	ScanReason      uint16
	ScannedChannels uint16
}

type AlpRssiData struct {
	ApMac     *AlpMac
	Channel   uint16
	RawBytes  []byte
	Signal    int16
	TXPower   uint16
	Timestamp uint32
}

type AlpIrData struct {
	IRMAC                   *AlpMac
	IRTransmitterID         [4]byte
	IRTransmitterIDAsString string
	RawBytes                []byte
	Signal                  int16
	Timestamp               int
}

type XpertAlpParser struct {
	DeviceData    AlpDeviceData
	Header        AlpHeader
	StartDateTime time.Time
	MBytes        []byte
	MBytesFull    []byte
	IrDataArray   []AlpIrData
	RssiDataArray []AlpRssiData
	ResponseTime  float64
}

const (
	MaxSeqNum = 65000
)

type Config struct {
	NumDevices  int
	IntervalSec int
	TargetIP    string
	TargetPort  int
	BaseMAC     string
}

func initLogger() *os.File {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// Create log file with timestamp in name
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("udp_sender_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Set log output to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)
	log.SetFlags(0)

	return file
}

func parseFlags() *Config {
	config := &Config{}

	// Define command line flags
	flag.IntVar(&config.NumDevices, "devices", 1, "Number of devices to simulate")
	flag.IntVar(&config.IntervalSec, "interval", 2, "Interval between packets in seconds")
	flag.StringVar(&config.TargetIP, "ip", "127.0.0.1", "Target IP address")
	flag.IntVar(&config.TargetPort, "port", 8552, "Target port number")
	flag.StringVar(&config.BaseMAC, "mac", "C4:CB:6B:23:00:01", "Base MAC address")

	// Parse flags
	flag.Parse()

	return config
}

func GenerateAlpAdditionalInfo() *AlpAdditionalInfo {
	return &AlpAdditionalInfo{
		Charging:          false,
		InitialScan:       true,
		MessagePending:    false,
		SafetySwitch:      false,
		SupplementaryScan: true,
	}
}

func GenerateAlpMac(hexString string) *AlpMac {
	var macBytes [6]byte
	hw, _ := hex.DecodeString(hexString)
	copy(macBytes[:], hw)
	return &AlpMac{
		AsHexString: hexString,
		BytesArray:  macBytes,
	}
}

func GenerateAlpHeader(sequence uint16, timestamp uint32) *AlpHeader {
	return &AlpHeader{
		Length:    64,
		Magic:     "ELP\x00",
		RawBytes:  []byte("ALP_HEADER_RAW_BYTES"),
		Sequence:  sequence,
		Timestamp: timestamp,
		Type:      1,
		Version:   2,
	}
}

func GenerateAlpDeviceData(deviceMac *AlpMac) *AlpDeviceData {
	return &AlpDeviceData{
		RawBytes:        []byte("DEVICE_RAW_BYTES"),
		DeviceMac:       deviceMac,
		DeviceType:      1000,
		MenuItem:        0,
		BatteryLevel:    100,
		MessageID:       0,
		AddnlInfo:       GenerateAlpAdditionalInfo(),
		RequestAck:      true,
		ScanReason:      3,
		ScannedChannels: 0x03FF,
	}
}

func GenerateAlpRssiData() []*AlpRssiData {
	apMacs := []string{
		"C4CB6B500001",
		"C4CB6B500002",
		"C4CB6B500003",
		"C4CB6B500004",
		"C4CB6B500005",
	}
	var rssiData []*AlpRssiData
	for _, mac := range apMacs {
		rssiData = append(rssiData, &AlpRssiData{
			ApMac:     GenerateAlpMac(mac),
			Channel:   11,
			RawBytes:  []byte("RSSI_RAW_BYTES"),
			Signal:    int16(rand.Intn(31) + 60),
			TXPower:   20,
			Timestamp: uint32(time.Now().UnixMilli() % 0xFFFFFFFF),
		})
	}
	return rssiData
}

func GenerateAlpIrData() *AlpIrData {
	irMac := "C4CB6B500101"
	return &AlpIrData{
		IRMAC:                   GenerateAlpMac(irMac),
		IRTransmitterID:         [4]byte{0xFF, 0xBF, 0x00, 0x01},
		IRTransmitterIDAsString: "FFBF0001",
		RawBytes:                []byte("IR_RAW_BYTES"),
		Signal:                  -75,
		Timestamp:               int(time.Now().UnixMilli() % 0xFFFFFFFF),
	}
}

func GenerateMAC(baseMAC string, index int) [6]byte {
	var mac [6]byte
	fmt.Sscanf(baseMAC, "%x:%x:%x:%x:%x:%x", &mac[0], &mac[1], &mac[2], &mac[3], &mac[4], &mac[5])

	macInt := uint64(mac[0])<<40 | uint64(mac[1])<<32 | uint64(mac[2])<<24 |
		uint64(mac[3])<<16 | uint64(mac[4])<<8 | uint64(mac[5])

	// Increment the MAC address
	macInt += uint64(index)

	// Convert back to a 6-byte MAC address
	mac[0] = byte((macInt >> 40) & 0xFF)
	mac[1] = byte((macInt >> 32) & 0xFF)
	mac[2] = byte((macInt >> 24) & 0xFF)
	mac[3] = byte((macInt >> 16) & 0xFF)
	mac[4] = byte((macInt >> 8) & 0xFF)
	mac[5] = byte(macInt & 0xFF)

	return mac
}

func BuildELPPacket(seqNum uint16, deviceData *AlpDeviceData, rssiData []*AlpRssiData, irData []*AlpIrData) []byte {
	var packet []byte

	// Header construction
	header := GenerateAlpHeader(seqNum, uint32(time.Now().UnixMilli()%0xFFFFFFFF))

	// Construct Device Chunk
	deviceChunk := make([]byte, 0)
	deviceChunk = append(deviceChunk, deviceData.DeviceMac.BytesArray[:]...)
	deviceChunk = append(deviceChunk, Uint16ToBytes(deviceData.DeviceType)...)
	deviceChunk = append(deviceChunk, deviceData.MenuItem)
	deviceChunk = append(deviceChunk, deviceData.BatteryLevel)
	deviceChunk = append(deviceChunk, deviceData.MessageID)
	deviceChunk = append(deviceChunk, BuildAdditionalInfo(deviceData.AddnlInfo))

	// ScanReason (2 bytes) - value 3 (Periodic) with RequestAck
	scanReason := uint16(deviceData.ScanReason)
	if deviceData.RequestAck {
		scanReason |= 0x8000 // Set the highest bit for RequestAck
	}
	deviceChunk = append(deviceChunk, Uint16ToBytes(scanReason)...)
	deviceChunk = append(deviceChunk, Uint16ToBytes(deviceData.ScannedChannels)...)

	// Construct RSSI Chunks
	rssiChunks := make([]byte, 0)
	for _, rssi := range rssiData {
		rssiChunks = append(rssiChunks, rssi.ApMac.BytesArray[:]...)
		rssiChunks = append(rssiChunks, Uint16ToBytes(uint16(rssi.Channel))...)
		rssiChunks = append(rssiChunks, Int16ToBytes(rssi.Signal)...)
		rssiChunks = append(rssiChunks, Uint16ToBytes(rssi.TXPower)...)
		rssiChunks = append(rssiChunks, Uint32ToBytes(rssi.Timestamp)...)
	}

	// Construct IR Chunks
	irChunks := make([]byte, 0)
	for _, ir := range irData {
		irChunks = append(irChunks, []byte{0xFF, 0xBF, 0x00, 0x01}...)
		irChunks = append(irChunks, ir.IRTransmitterID[:]...)
		irChunks = append(irChunks, []byte{0x00, 0x00, 0x00, 0x00}...)
		irChunks = append(irChunks, Uint32ToBytes(uint32(ir.Timestamp))...)
	}

	// Calculate Length (sum of all chunks excluding the header)
	totalLength := len(deviceChunk) + len(rssiChunks) + len(irChunks)
	header.Length = uint16(totalLength)

	// Serialize Header
	packet = append(packet, []byte(header.Magic)...)
	packet = append(packet, Uint16ToBytes(header.Version)...)
	packet = append(packet, Uint16ToBytes(header.Type)...)
	packet = append(packet, Uint16ToBytes(header.Sequence)...)
	packet = append(packet, Uint32ToBytes(header.Timestamp)...)
	packet = append(packet, Uint16ToBytes(header.Length)...)

	// Append Device, RSSI, and IR Chunks
	packet = append(packet, deviceChunk...)
	packet = append(packet, rssiChunks...)
	packet = append(packet, irChunks...)

	return packet
}

func Uint16ToBytes(val uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, val)
	return buf
}

func Uint32ToBytes(val uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, val)
	return buf
}

func Int16ToBytes(val int16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(val))
	return buf
}

func BuildAdditionalInfo(info *AlpAdditionalInfo) uint8 {
	var result uint8
	if info.Charging {
		result |= 1 << 0
	}
	if info.SafetySwitch {
		result |= 1 << 1
	}
	if info.InitialScan {
		result |= 1 << 2
	}
	if info.SupplementaryScan {
		result |= 1 << 3
	}
	if info.MessagePending {
		result |= 1 << 4
	}
	return result
}

func logWithTime(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	log.Printf("[%s] %s", timestamp, message)
}

// SendPackets simulates a device sending packets
func SendPackets(deviceIndex int, baseMAC string, interval time.Duration, targetAddr string, wg *sync.WaitGroup) {
	defer wg.Done()
	mac := GenerateMAC(baseMAC, deviceIndex)
	conn, err := net.Dial("udp", targetAddr)
	if err != nil {
		logWithTime("[Device %d] Error: %v", deviceIndex, err)
		return
	}
	defer conn.Close()

	seqNum := uint16(1)

	for {
		deviceMac := &AlpMac{
			AsHexString: fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]),
			BytesArray:  mac,
		}
		deviceData := GenerateAlpDeviceData(deviceMac)

		rssiData := GenerateAlpRssiData()
		irData := []*AlpIrData{GenerateAlpIrData()}

		packet := BuildELPPacket(seqNum, deviceData, rssiData, irData)

		// Send the packet
		_, writeErr := conn.Write(packet)
		if writeErr != nil {
			logWithTime("[Device %d] Error sending packet: %v", deviceIndex, writeErr)
			break
		}

		logWithTime("[Device %d] Sent packet: Seq=%d, MAC=%X",
			deviceIndex, seqNum, mac)

		seqNum++
		if seqNum > MaxSeqNum {
			seqNum = 1
		}

		time.Sleep(interval)
	}
}

func main() {
	// Initialize logger
	logFile := initLogger()
	defer logFile.Close()

	// Parse command line flags
	config := parseFlags()

	// Log configuration
	logWithTime("Starting UDP sender with configuration:")
	logWithTime("Number of devices: %d", config.NumDevices)
	logWithTime("Interval: %d seconds", config.IntervalSec)
	logWithTime("Target: %s:%d", config.TargetIP, config.TargetPort)
	logWithTime("Base MAC: %s", config.BaseMAC)

	targetAddr := fmt.Sprintf("%s:%d", config.TargetIP, config.TargetPort)

	var wg sync.WaitGroup
	for i := 0; i < config.NumDevices; i++ {
		wg.Add(1)
		go func(deviceIndex int) {
			SendPackets(deviceIndex, config.BaseMAC, time.Duration(config.IntervalSec)*time.Second, targetAddr, &wg)
		}(i)
	}

	wg.Wait()
	log.Println("Simulation complete.")
}

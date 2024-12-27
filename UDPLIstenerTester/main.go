package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
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
	mac[5] += byte(index)
	for i := 5; i > 0; i-- {
		if mac[i] > 255 {
			mac[i] = 0
			mac[i-1]++
		}
	}
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
	additionalInfo := BuildAdditionalInfo(deviceData.AddnlInfo)
	deviceChunk = append(deviceChunk, additionalInfo)
	deviceChunk = append(deviceChunk, Uint16ToBytes(deviceData.ScanReason)...)
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

func logWithTime(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s\n", timestamp, message)
}

// SendPackets simulates a device sending packets
func SendPackets(deviceIndex int, baseMAC string, interval time.Duration, targetAddr string, wg *sync.WaitGroup) {
	defer wg.Done()
	mac := GenerateMAC(baseMAC, deviceIndex)
	conn, err := net.Dial("udp", targetAddr)
	if err != nil {
		logWithTime(fmt.Sprintf("[Device %d] Error: %v", deviceIndex, err))
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
			logWithTime(fmt.Sprintf("[Device %d] Error sending packet: %v", deviceIndex, writeErr))
			break
		}

		logWithTime(fmt.Sprintf("[Device %d] Sent packet: Seq=%d, MAC=%X", deviceIndex, seqNum, mac))

		seqNum++
		if seqNum > MaxSeqNum {
			seqNum = 1
		}

		time.Sleep(interval)
	}
}

func main() {
	numDevices := Config.NumDevices
	intervalSec := Config.IntervalSec
	targetIP := Config.TargetIP
	targetPort := Config.TargetPort
	targetAddr := fmt.Sprintf("%s:%d", targetIP, targetPort)
	baseMAC := Config.BaseMAC

	fmt.Printf("Starting simulation with %d devices, interval: %d seconds, target: %s\n", numDevices, intervalSec, targetAddr)

	var wg sync.WaitGroup
	for i := 0; i < numDevices; i++ {
		wg.Add(1)
		go func(deviceIndex int) {
			SendPackets(deviceIndex, baseMAC, time.Duration(intervalSec)*time.Second, targetAddr, &wg)
		}(i)
	}

	wg.Wait()
	fmt.Println("Simulation complete.")
}

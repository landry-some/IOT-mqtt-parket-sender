package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type NPITester struct {
	host        string
	port        string
	numTags     int
	numInfra    int
	interval    time.Duration
	connections map[string]net.Conn
	connLock    sync.Mutex
	stopChan    chan struct{}
}

type NPIPacket struct {
	PayloadLength uint16
	MessageType   byte
	SubSystemID   byte
	CommandID     byte
	Payload       []byte
	Checksum      byte
}

func initLogger() *os.File {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// Create log file with timestamp in name
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("npi_parser_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Set log output to file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	log.Printf("NPI Parser logging started")
	return file
}

func main() {
	// Initialize logger
	logFile := initLogger()
	defer logFile.Close()

	tester := &NPITester{
		connections: make(map[string]net.Conn),
		stopChan:    make(chan struct{}),
	}

	flag.StringVar(&tester.host, "host", "127.0.0.1", "Host to connect to")
	flag.StringVar(&tester.port, "port", "8888", "Port to connect to")
	flag.IntVar(&tester.numTags, "tags", 1, "Number of tags to simulate")
	flag.IntVar(&tester.numInfra, "infra", 1, "Number of infrastructure devices")
	interval := flag.Int("interval", 1, "Interval between packets in seconds")
	flag.Parse()

	tester.interval = time.Duration(*interval) * time.Second

	log.Printf("Starting NPI Parser Tester with settings:")
	log.Printf("Host: %s", tester.host)
	log.Printf("Port: %s", tester.port)
	log.Printf("Number of tags: %d", tester.numTags)
	log.Printf("Number of infrastructure devices: %d", tester.numInfra)
	log.Printf("Interval: %v", tester.interval)

	if err := tester.Start(); err != nil {
		log.Printf("Failed to start tester: %v", err)
		return
	}

	waitForInterrupt(tester)
}

func (t *NPITester) Start() error {
	for i := 0; i < t.numInfra; i++ {
		infraMAC := fmt.Sprintf("C4:CB:6B:50:00:%02X", i+1)
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", t.host, t.port))
		if err != nil {
			return fmt.Errorf("failed to connect to %s:%s: %v", t.host, t.port, err)
		}

		t.connLock.Lock()
		t.connections[infraMAC] = conn
		t.connLock.Unlock()

		go t.sendPackets(infraMAC, conn)
	}

	return nil
}

func (t *NPITester) sendPackets(infraMAC string, conn net.Conn) {
	sqn := uint16(1)
	bleSqn := byte(1)

	for {
		select {
		case <-t.stopChan:
			log.Printf("Stopping packet sender for infra %s", infraMAC)
			return
		default:
			for tagNum := 0; tagNum < t.numTags; tagNum++ {
				tagMAC := fmt.Sprintf("C4:CB:6B:23:00:%02X", tagNum+1)
				packet := t.createNPIPacket(infraMAC, tagMAC, sqn, bleSqn)

				// Add delimiter (SOF)
				if _, err := conn.Write([]byte{0xFE}); err != nil {
					log.Printf("Error writing SOF: %v", err)
					return
				}

				// Send packet
				if _, err := conn.Write(packet); err != nil {
					log.Printf("Error sending packet: %v", err)
					return
				}

				log.Printf("Sent packet from infra %s for tag %s (sqn: %d, bleSqn: %d)",
					infraMAC, tagMAC, sqn, bleSqn)

				// Optional: log packet bytes in hex
				log.Printf("Packet bytes: %X", packet)
			}

			sqn++
			if sqn > 65000 {
				log.Printf("Resetting sqn counter")
				sqn = 1
			}

			bleSqn++
			if bleSqn > 255 {
				log.Printf("Resetting bleSqn counter")
				bleSqn = 1
			}

			time.Sleep(t.interval)
		}
	}
}

func (t *NPITester) createNPIPacket(infraMAC, tagMAC string, sqn uint16, bleSqn byte) []byte {
	iBeaconPayload := createIBeaconPayload(bleSqn, sqn)
	advPayload := createAdvPayload(infraMAC, tagMAC, iBeaconPayload)

	payloadLength := uint16(len(advPayload))
	packet := make([]byte, payloadLength+5)

	// PayloadLength
	binary.LittleEndian.PutUint16(packet[0:2], payloadLength)

	// MessageType and SubSystemID combined in one byte
	packet[2] = (0x02 << 5) | 0x0F // MessageType=2, SubSystemID=15

	// CommandID must be 5 (LegacyAdvReport)
	packet[3] = 0x05

	// Payload
	copy(packet[4:], advPayload)

	// Calculate checksum
	checksum := packet[0]
	for i := 1; i < len(packet)-1; i++ {
		checksum ^= packet[i]
	}
	packet[len(packet)-1] = checksum

	return packet
}

func createAdvPayload(infraMAC, tagMAC string, iBeaconPayload []byte) []byte {
	payload := make([]byte, 14+len(iBeaconPayload))

	// Source MAC (infrastructure) - needs to be reversed to match GetMacFromBytes(ReverseBytes())
	infraBytes := macToBytes(infraMAC)
	reverseBytes(infraBytes)
	copy(payload[0:6], infraBytes)

	// Target MAC (tag) - needs to be reversed to match GetMacFromBytes(ReverseBytes())
	tagBytes := macToBytes(tagMAC)
	reverseBytes(tagBytes)
	copy(payload[6:12], tagBytes)

	// Source RSSI (parser will make it negative)
	payload[12] = byte(60 + (time.Now().UnixNano() % 31))

	// BLE_ADV_TYPE_IBEACON must be 1 to match parser check
	payload[13] = 0x01

	copy(payload[14:], iBeaconPayload)
	return payload
}

func createIBeaconPayload(bleSqn byte, wifiSqn uint16) []byte {
	payload := make([]byte, 30) // must match iBeaconPayloadLength

	// GAP Advertisement flags
	payload[0] = 0x02
	payload[1] = 0x01
	payload[2] = 0x06

	// iBeacon header must match "1AFF4C000215"
	iBeaconHeader := []byte{0x1A, 0xFF, 0x4C, 0x00, 0x02, 0x15}
	copy(payload[3:9], iBeaconHeader)

	// iBeacon UUID starts at index 9
	uuid := make([]byte, 21)

	// Fields must match parser's extraction points
	uuid[3] = bleSqn                                  // BLESQN
	binary.LittleEndian.PutUint16(uuid[4:6], wifiSqn) // WiFiSQN
	uuid[6] = 31                                      // DeviceType (W4)
	uuid[7] = 50                                      // Humidity

	// Temperature calculation to match parser's float32 conversion
	tempValue := uint16(25.5 * 256) // Example temperature
	binary.LittleEndian.PutUint16(uuid[8:10], tempValue)

	// Status byte with flags
	uuid[10] = 0x07 // StatusByte (SafetySwitch=1, Charger=1, Motion=1)
	uuid[11] = 0x01 // ReportReason
	uuid[12] = 100  // Battery

	copy(payload[9:], uuid)
	return payload
}

func macToBytes(mac string) []byte {
	bytes := make([]byte, 6)
	fmt.Sscanf(mac, "%02x:%02x:%02x:%02x:%02x:%02x",
		&bytes[0], &bytes[1], &bytes[2], &bytes[3], &bytes[4], &bytes[5])
	return bytes
}

func reverseBytes(b []byte) {
	for i := 0; i < len(b)/2; i++ {
		j := len(b) - i - 1
		b[i], b[j] = b[j], b[i]
	}
}

func waitForInterrupt(t *NPITester) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	log.Printf("Received signal %v, shutting down...", sig)
	close(t.stopChan)

	t.connLock.Lock()
	defer t.connLock.Unlock()

	for mac, conn := range t.connections {
		log.Printf("Closing connection for %s", mac)
		conn.Close()
	}

	log.Printf("Shutdown complete")
}

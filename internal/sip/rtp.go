package sip

import (
	"fmt"
		"math/rand"
   "encoding/binary"
   "net"
   "time"
)

type RTPPacket struct {
   Version        uint8
   Padding        bool
   Extension      bool
   CSRCCount      uint8
   Marker         bool
   PayloadType    uint8
   SequenceNumber uint16
   Timestamp      uint32
   SSRC           uint32
   Payload        []byte
}

func SendRTPStream(localIP string, localPort int, remoteIP string, remotePort int, duration time.Duration) error {
   conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", remoteIP, remotePort))
   if err != nil {
       return err
   }
   defer conn.Close()
   
   // RTP parameters
   ssrc := uint32(rand.Int31())
   sequenceNumber := uint16(rand.Intn(65535))
   timestamp := uint32(rand.Int31())
   
   // Send RTP packets (50 packets per second for 20ms intervals)
   ticker := time.NewTicker(20 * time.Millisecond)
   defer ticker.Stop()
   
   timeout := time.After(duration)
   
   for {
       select {
       case <-ticker.C:
           packet := createRTPPacket(sequenceNumber, timestamp, ssrc)
           conn.Write(packet)
           sequenceNumber++
           timestamp += 160 // 160 samples at 8kHz for 20ms
       case <-timeout:
           return nil
       }
   }
}

func createRTPPacket(seq uint16, ts uint32, ssrc uint32) []byte {
   packet := make([]byte, 12+160) // RTP header + 160 bytes of audio
   
   // RTP header
   packet[0] = 0x80 // Version 2, no padding, no extension, no CSRC
   packet[1] = 0    // Marker = 0, Payload type = 0 (PCMU)
   
   binary.BigEndian.PutUint16(packet[2:4], seq)
   binary.BigEndian.PutUint32(packet[4:8], ts)
   binary.BigEndian.PutUint32(packet[8:12], ssrc)
   
   // Fill with silence (0xFF for PCMU)
   for i := 12; i < len(packet); i++ {
       packet[i] = 0xFF
   }
   
   return packet
}

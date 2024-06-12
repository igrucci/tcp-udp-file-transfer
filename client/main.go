package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	fmt.Println("host tcp udp filepath time")

	input := bufio.NewScanner(os.Stdin)
	input.Split(bufio.ScanWords)

	input.Scan()
	host := input.Text()

	input.Scan()
	tcpPort := input.Text()

	input.Scan()
	udpPort := input.Text()

	udpPortLen := len(udpPort)

	input.Scan()
	filePath := input.Text()

	fileName := filepath.Base(filePath)

	input.Scan()
	timeForApprove, err := strconv.Atoi(input.Text())
	if err != nil {
		log.Println("Error parsing time ", err)
		return
	}

	tcpConn, err := net.Dial("tcp", host+":"+tcpPort)
	if err != nil {
		log.Println("Error dialing tcp ", err)
		return
	}

	defer tcpConn.Close()

	udpPortLenBytes := []byte{byte(udpPortLen)}
	data := append(udpPortLenBytes, []byte(udpPort)...)
	data = append(data, []byte(fileName)...)

	if _, err := tcpConn.Write(data); err != nil {
		log.Println("Error writing to tcp ", err)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file ", err)
		return
	}
	defer file.Close()

	udpConn, err := net.Dial("udp", host+":"+udpPort)
	if err != nil {
		log.Println("Error dialing udp ", err)
		return
	}

	defer udpConn.Close()

	buf := make([]byte, 200)

	packetID := 1

	for {
		bytesFile, err := file.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Println("Error reading file ", err)
			return
		}
		packet := append([]byte(fmt.Sprintf("%d:", packetID)), buf[:bytesFile]...)

		log.Println("send", packet, packetID)
		_, err = udpConn.Write(packet)

		if err != nil {
			log.Println("Error writing to udp", err)
			return
		}

		chanConf := make(chan interface{}, 1)
		go func() {
			confirmation := make([]byte, 1024)
			_, err = tcpConn.Read(confirmation)
			if err != nil {
				log.Println("Error reading tcp ", err)
				return
			}
			chanConf <- string(confirmation)

		}()
		select {
		case <-time.After(time.Duration(timeForApprove) * time.Millisecond):

			packet := append([]byte(fmt.Sprintf("%d:", packetID)), buf[:bytesFile]...)
			_, err = udpConn.Write(packet)
			if err != nil {
				log.Println("Error writing to udp ", err)
				return
			}
			log.Println("send again ", packetID)

		case _ = <-chanConf:
			log.Println("packet ", packetID, "delivered")
		}

		packetID++
	}
}

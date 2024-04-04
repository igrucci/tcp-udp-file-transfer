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
	input := bufio.NewScanner(os.Stdin)
	input.Split(bufio.ScanWords)

	input.Scan()
	host := input.Text()

	input.Scan()
	tcpPort := input.Text()

	input.Scan()
	udpPort := input.Text()

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

	_, err = tcpConn.Write([]byte(udpPort + fileName))
	if err != nil {
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

	buf := make([]byte, 100)

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
		log.Println(packet, packetID)
		_, err = udpConn.Write(packet)
		if err != nil {
			log.Println("Error writing to udp", err)
			return
		}

		timeout := time.After(time.Duration(timeForApprove) * time.Millisecond)

		select {
		case <-timeout:

			packet := append([]byte(fmt.Sprintf("%d:", packetID)), buf[:bytesFile]...)
			_, err = udpConn.Write(packet)
			if err != nil {
				log.Println("Error writing to udp ", err)
				return
			}
			log.Println("send again ", packetID)

		default:
			confirmation := make([]byte, 1024)
			_, err = tcpConn.Read(confirmation)
			if err != nil {
				log.Println("Error reading tcp ", err)
				return
			}
			log.Println(string(confirmation))
		}

		packetID++

	}
}

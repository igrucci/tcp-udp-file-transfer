package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func handleTCPConn(c net.Conn) {
	defer c.Close()
	log.Printf("Serving tcp  %s\n", c.RemoteAddr().String())

	buf := make([]byte, 1024)
	_, err := c.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return
		}
		log.Println("Error reading from tcp", err)
		return
	}

}

func handleUDPConn(udpPort string, tcpConn net.Conn, fileName string) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+udpPort)
	if err != nil {
		log.Println("Error resolving udp addr:", err)
		return
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("Error listening on udp port:", err)
		return
	}
	defer udpConn.Close()

	log.Println("udp server listening on port", udpPort)

	file, err := os.OpenFile("files/"+fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	buf := make([]byte, 1024)
	for {
		n, _, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error reading udp packet:", err)
			continue
		}
		packet := buf[:n]
		parts := strings.Split(string(packet), ":")
		packetId, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Println("Error parsing id:", err)
			continue
		}
		data := parts[1]

		fmt.Println("Received udp packet  : id:", packetId, data)

		_, err = file.WriteString(data)
		if err != nil {
			log.Println("Error writing to file:", err)
			return
		}

		_, err = tcpConn.Write([]byte("received"))
		if err != nil {
			fmt.Println("Error writing to tcp", err)
			return
		}
	}
}

func main() {

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening on tcp port ", err)
		return
	}

	for {
		tcpConn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting tcp connection", err)
			continue
		}
		go handleTCPConn(tcpConn)

		defer tcpConn.Close()
		fmt.Printf("Serving tcp  %s\n", tcpConn.RemoteAddr().String())

		buf := make([]byte, 1024)
		n, err := tcpConn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from tcp", err)
			return
		}
		udpPort := string(buf[:4])
		fileName := string(buf[4:n])
		if err := os.MkdirAll("files", 0777); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}

		go handleUDPConn(udpPort, tcpConn, fileName)

		if _, err := os.Create("files/" + fileName); err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
	}
}

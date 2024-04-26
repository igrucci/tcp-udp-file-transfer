package main

import (
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

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

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

		log.Println("Received udp packet  : id:", packetId, data)

		_, err = file.WriteString(data)

		if err != nil {
			log.Println("Error writing to file:", err)
			return
		}

		_, err = tcpConn.Write([]byte("received"))
		if err != nil {
			log.Println("Error writing to tcp", err)
			return
		}
	}
}

func main() {

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println("Error listening on tcp port ", err)
		return
	}

	for {
		tcpConn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting tcp connection", err)
			continue
		}

		go handleTCPConn(tcpConn)

		defer tcpConn.Close()
		log.Printf("Serving tcp  %s\n", tcpConn.RemoteAddr().String())

		buf := make([]byte, 1024)
		n, err := tcpConn.Read(buf)

		if err != nil {
			log.Println("Error reading from tcp", err)
			return
		}

		udpPortLen := int(buf[0])

		udpPort := string(buf[1 : udpPortLen+1])

		fileName := string(buf[udpPortLen+1 : n])

		if err := os.MkdirAll("files", 0777); err != nil {
			log.Println("Error creating directory:", err)
			return
		}
		if _, err := os.Create("files/" + fileName); err != nil {
			log.Println("Error creating file:", err)
			return
		}
		go handleUDPConn(udpPort, tcpConn, fileName)

	}
}

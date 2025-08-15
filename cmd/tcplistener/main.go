package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer close(out)
		defer f.Close()
		str := ""
		for {
			endBuffer := make([]byte, 8)
			n, err := f.Read(endBuffer)
			if err != nil {
				break
			}
			readStr := strings.Split(string(endBuffer[:n]), "\n")
			if len(readStr) > 1 {
				str = str + readStr[0]
				out <- str
				str = readStr[1]
			} else if len(readStr) == 1 {
				str = str + readStr[0]
			}
		}
	}()
	return out
}

func getFromFile(fileLocation string) *os.File {
	file, err := os.Open(fileLocation)
	if err != nil {
		log.Fatalln("error on the file read")
	}
	return file
}

func listenFromNetwork(network string, address string) net.Listener {
	Listen, err := net.Listen(network, address)
	if err != nil {
		log.Fatal("error on the file read")
	}
	return Listen
}

func main() {
	ports := ":42069"
	f := listenFromNetwork("tcp", ports)
	for {
		conn, err := f.Accept()
		if err != nil {
			log.Fatal("error on the file read")
			break
		}
		fmt.Printf("tcp network is connected to port %s\n", ports)
		for line := range getLinesChannel(conn) {
			if line == "exit" {
				f.Close()
				fmt.Printf("host request to close the connect;")
				os.Exit(0)
			}
			fmt.Printf("Read: %s\n", line)
		}
	}
	f.Close()

}

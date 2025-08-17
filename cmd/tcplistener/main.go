package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"boot.theprimeagen.tv/internal/request"
)

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
		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error on the file read")
			break
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
		if r.Headers.Size() > 0 {
			fmt.Printf("Headers:\n")
			r.Headers.ForEach(printHeaders)
		}
	}
	f.Close()

}

func printHeaders(key string, value string) {
	fmt.Printf("- %s:%s\n", key, value)
}

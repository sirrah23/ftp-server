package main

import "net"
import "fmt"

func main() {
	fmt.Println("Connecting now!")
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		// handle error
	}
	fmt.Println("Connected!")
	conn, err := ln.Accept()
	if err != nil {
		// handle error
		fmt.Println("Error!")
	}
	conn.Write([]byte("220 Hi There\n"))
	rcvB := make([]byte, 1024)
	for {
		_, err = conn.Read(rcvB)
		fmt.Println(string(rcvB))
		conn.Write([]byte("230 Logged In\n"))
	}
}

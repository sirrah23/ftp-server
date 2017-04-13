package main

import "net"
import "fmt"
import "strings"
import "strconv"

func main() {
	fmt.Println("Connecting now!")
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		// handle error
	}
	fmt.Println("Connected!")
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Println("Error!")
		}
		go connection_handler(conn)
	}
}

func connection_handler(conn net.Conn) {
	conn.Write([]byte("220 You are connected!\n"))
	rcvB := make([]byte, 1024)
	for {
		n, err := conn.Read(rcvB)
		if err != nil {
			//handle error
		}
		input := string(rcvB[:n])
		words := strings.Split(input, " ")
		fmt.Println(words[0])
		if len(words) > 1 {
			fmt.Println(strings.Compare(words[1], "anonymous\r\n"))
		}
		code := input_handler(words)
		fmt.Println(code)
		conn.Write([]byte(strconv.Itoa(code) + "\n")) //string(code) does not WORK PROPERLY AHHHHH
	}
}

func input_handler(words []string) int {
	if words[0] == "USER" {
		if strings.Compare(words[1], "anonymous\r\n") == 0 {
			return 230 //User logged in, proceed
		} else {
			return 530 //Not logged in
		}
	} else if strings.Compare(words[0], "SYST\r\n") == 0 {
		return 215 // System information
	} else {
		return 500 //Syntax error
	}
}

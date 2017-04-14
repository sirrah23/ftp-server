package main

import "net"
import "fmt"
import "strings"
import "strconv"
import "os"

type response struct {
	code int
	msg  string
}

func (r response) GenerateMsgStr() string {
	//TODO: Sprintf instead
	return strconv.Itoa(r.code) + " " + r.msg + "\n"
}

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
		response := input_handler(words)
		responseStr := response.GenerateMsgStr()
		fmt.Println(responseStr)
		conn.Write([]byte(responseStr)) //string(code) does not WORK PROPERLY AHHHHH
	}
}

func input_handler(input []string) response {
	if input[0] == "USER" {
		return login_handler(input)
	} else if strings.Compare(input[0], "SYST\r\n") == 0 { //TODO: Clean \r\n upfront
		return response{code: 215, msg: "Special FTP Server :)"}
	} else if strings.Compare(input[0], "PWD\r\n") == 0 {
		return pwd_handler(input)
	} else {
		return response{code: 500, msg: "Syntax error"}
	}
}

func login_handler(input []string) response {
	if strings.Compare(input[1], "anonymous\r\n") == 0 {
		return response{code: 230, msg: "Login successful"}
	} else {
		return response{code: 530, msg: "Login successful"}
	}
}

func pwd_handler(input []string) response {
	dir, err := os.Getwd()
	if err != nil {
		//handle error
	}
	return response{code: 257, msg: dir}
}

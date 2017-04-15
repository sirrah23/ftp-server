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

type ftp_dialer struct {
        data_sender net.Conn
}

func (f ftp_dialer) port_handler(input []string) response {
        if f.data_sender != nil {
                return response{code: 500, msg: "Port already specified for data retrieval"}
        }
        addr_str := port_address_str(input[1])
        fmt.Println(addr_str)
        conn, err := net.Dial("tcp", addr_str)
        if err != nil {
                fmt.Println(err)
                return response{code: 500, msg: "Cannot establish a port for data retrieval"}
        }
        f.data_sender = conn
        return response{code: 200, msg: "Port established"}
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
	conn.Write([]byte(successful_connection().GenerateMsgStr()))
	rcvB := make([]byte, 1024)
    dialer := ftp_dialer{data_sender: nil}
	for {
		n, err := conn.Read(rcvB)
		if err != nil {
			//handle error
		}
		input := string(rcvB[:n])
        fmt.Println(input)
		words := strings.Split(input, " ")
		response := input_handler(words, dialer)
		conn.Write([]byte(response.GenerateMsgStr()))
	}
}

func input_handler(input []string, dialer ftp_dialer) response {
	if input[0] == "USER" {
		return login_handler(input)
	} else if strings.Compare(input[0], "SYST\r\n") == 0 { //TODO: Clean \r\n upfront
		return response{code: 215, msg: "Special FTP Server :)"}
	} else if strings.Compare(input[0], "PWD\r\n") == 0 {
		return pwd_handler(input)
	} else if strings.Compare(input[0], "PORT") == 0 {
        return dialer.port_handler(input)
    } else {
		return response{code: 500, msg: "Syntax error"}
	}
}

func successful_connection() response {
	return response{code: 220, msg: " You are connected!"}
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

func port_address_str(host_port string) (addr_str string) {
        host_port_sp := strings.Split(host_port, ",") //h1, h2, h3, h4, p1, p2
        p1, _ := strconv.Atoi(host_port_sp[4])
        p1 *= 256
        p2, _ := strconv.Atoi(host_port_sp[5])
        port_str := strconv.Itoa(p1 + p2)
        addr_str = /*strings.Join(host_port_sp[:4],".") +*/":" + port_str
        return
}


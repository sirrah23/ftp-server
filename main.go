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

type data_connection struct {
	c chan int
}

func (r response) GenerateMsgStr() string {
	//TODO: Sprintf instead
	return strconv.Itoa(r.code) + " " + r.msg + "\r\n"
}

func passive_handler() (response, chan int) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		return response{code: 500, msg: "Cannot establish a port for data retrieval"}, nil
	}
	fmt.Println(ln.Addr().String())
	address := port_split_str(ln.Addr().String())
	c := make(chan int)
	go data_handler(ln, c)
	return response{code: 227, msg: "Entering Passive Mode (" + address + ")"}, c
}

func main() {
	fmt.Println("Connecting now!")
	ln, err := net.Listen("tcp", ":21")
	if err != nil {
		// handle error
		fmt.Println(err)
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
	var c chan int = nil
	var r response
	for {
		n, err := conn.Read(rcvB)
		if err != nil {
			//handle error
		}
		input := string(rcvB[:n])
		fmt.Println(input)
		words := strings.Split(input, " ")
		r, c = input_handler(words, c, conn)
		if r.code != -1 {
			conn.Write([]byte(r.GenerateMsgStr()))
		}
	}
}

func input_handler(input []string, c chan int, conn net.Conn) (response, chan int) {
	if input[0] == "USER" {
		return login_handler(input), nil
	} else if strings.Compare(input[0], "SYST\r\n") == 0 { //TODO: Clean \r\n upfront
		return response{code: 215, msg: "Special FTP Server :)"}, nil
	} else if strings.Compare(input[0], "PWD\r\n") == 0 {
		return pwd_handler(input), nil
	} else if strings.Compare(input[0], "PASV\r\n") == 0 {
		return passive_handler()
	} else if strings.Compare(input[0], "LIST\r\n") == 0 {
		conn.Write([]byte("150 File list send starting\r\n"))
		c <- 0
		<-c
		close(c)
		conn.Write([]byte("226 File list complete\r\n"))
		return response{code: -1, msg: ""}, nil
	} else {
		return response{code: 500, msg: "Syntax error"}, nil
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
	addr_str = strings.Join(host_port_sp[:4], ".") + ":" + port_str
	return
}

func data_handler(ln net.Listener, c chan int) {
	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Passive connection accepted!")
	}
	<-c
	conn.Write([]byte("file.txt\r\n"))
	conn.Close()
	c <- 0
	return
}

func port_split_str(address_str string) string {
	addr_port := strings.Split(address_str, ":")
	address := strings.Join(strings.Split(addr_port[0], "."), ",")
	port_int, _ := strconv.Atoi(addr_port[1])
	port16 := uint16(port_int)
	p1 := strconv.Itoa(int(port16 >> 8))
	p2 := strconv.Itoa(int(port16 << 8 >> 8))
	return address + "," + p1 + "," + p2
}

package main

import "net"
import "fmt"
import "strings"
import "strconv"
import "os"
import "io/ioutil"

type data_conn_info struct {
	code int
	info string
}

func GenerateMsgStr(code int, msg string) string {
	//TODO: Sprintf instead
	return strconv.Itoa(code) + " " + msg + "\r\n"
}

func passive_handler(conn net.Conn, data_conn_ch chan data_conn_info) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		response := GenerateMsgStr(500, "Cannot establish a port for data retrieval")
		conn.Write([]byte(response))
	}
	address := port_split_str(ln.Addr().String())
	go data_handler(ln, data_conn_ch)
	response := GenerateMsgStr(227, "Entering Passive Mode ("+address+")")
	conn.Write([]byte(response))
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
	successful_connection(conn)
	rcvB := make([]byte, 1024)
	data_conn_ch := make(chan data_conn_info)
	for {
		n, err := conn.Read(rcvB)
		if err != nil {
			//handle error
		}
		input := string(rcvB[:n])
		input = clean_CRLF(input)
		fmt.Println(input)
		words := strings.Split(input, " ")
		if words[0] == "QUIT" {
			end_connection(conn)
			return
		}
		input_handler(words, data_conn_ch, conn)
	}
}

func input_handler(input []string, data_conn_ch chan data_conn_info, conn net.Conn) {
	if input[0] == "USER" {
		login_handler(input, conn)
	} else if strings.Compare(input[0], "SYST") == 0 {
		syst_handler(conn)
	} else if strings.Compare(input[0], "PWD") == 0 {
		pwd_handler(input, conn)
	} else if strings.Compare(input[0], "CWD") == 0 {
		cwd_handler(input, conn)
	} else if strings.Compare(input[0], "PASV") == 0 {
		passive_handler(conn, data_conn_ch)
	} else if strings.Compare(input[0], "LIST") == 0 {
		ls_handler(conn, data_conn_ch)
	} else {
		syntax_err_handler(conn)
	}
}

func successful_connection(conn net.Conn) {
	response := GenerateMsgStr(220, "You are connected!")
	conn.Write([]byte(response))
}

func login_handler(input []string, conn net.Conn) {
	var response string
	if strings.Compare(input[1], "anonymous") == 0 {
		response = GenerateMsgStr(230, "Login successful!")
	} else {
		response = GenerateMsgStr(530, "Login unsuccessful")
	}
	conn.Write([]byte(response))
}

func pwd_handler(input []string, conn net.Conn) {
	dir, err := os.Getwd()
	if err != nil {
		//handle error
	}
	response := GenerateMsgStr(257, dir)
	conn.Write([]byte(response))
}

func cwd_handler(input []string, conn net.Conn) {
	dir := input[1]
	err := os.Chdir(dir)
	if err != nil {
		conn.Write([]byte(GenerateMsgStr(550, "Change directory failed")))
		return
	}
	conn.Write([]byte(GenerateMsgStr(250, "Success")))

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

func data_handler(ln net.Listener, data_conn_ch chan data_conn_info) {
	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Passive connection accepted!")
	}
	defer conn.Close()
	data_conn_req := <-data_conn_ch
	if data_conn_req.code == 0 {
		send_file_list(conn)
	} else {
		//fname := data_conn_req.info
		//send_file(fname)
	}
	data_conn_ch <- data_conn_info{}
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

func ls_handler(conn net.Conn, data_conn_ch chan data_conn_info) {
	var response string
	response = GenerateMsgStr(150, "File list send starting")
	conn.Write([]byte(response))
	data_conn_ch <- data_conn_info{code: 0, info: ""}
	<-data_conn_ch
	response = GenerateMsgStr(226, "File list send complete")
	conn.Write([]byte(response))
}

func get_files_dir(dir string) string {
	var file_list []string
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("Something went wrong during ls!")
		return ""
	}
	for _, file := range files {
		file_list = append(file_list, file.Name())
	}
	CRLF := "\r\n"
	return strings.Join(file_list, CRLF) + CRLF
}

func send_file_list(conn net.Conn) {
	file_list_str := get_files_dir(".")
	conn.Write([]byte(file_list_str))
}

func end_connection(conn net.Conn) {
	defer conn.Close()
	response := GenerateMsgStr(221, "Quitting!")
	conn.Write([]byte(response))
}

func syntax_err_handler(conn net.Conn) {
	response := GenerateMsgStr(500, "Syntax error")
	conn.Write([]byte(response))
}

func syst_handler(conn net.Conn) {
	response := GenerateMsgStr(215, "Special FTP Server :)")
	conn.Write([]byte(response))
}

func clean_CRLF(s string) string {
	CRLF := "\r\n"
	return strings.Replace(s, CRLF, "", -1)
}

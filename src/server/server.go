package server

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

func generateMsgStr(code int, msg string) string {
	//TODO: Sprintf instead
	return strconv.Itoa(code) + " " + msg + "\r\n"
}

func passiveHandler(conn net.Conn, data_conn_ch chan data_conn_info) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		response := generateMsgStr(500, "Cannot establish a port for data retrieval")
		conn.Write([]byte(response))
	}
	address := portSplitStr(ln.Addr().String())
	go dataHandler(ln, data_conn_ch)
	response := generateMsgStr(227, "Entering Passive Mode ("+address+")")
	conn.Write([]byte(response))
}

func connectionHandler(conn net.Conn) {
	successfulConnection(conn)
	rcvB := make([]byte, 1024)
	data_conn_ch := make(chan data_conn_info)
	for {
		n, err := conn.Read(rcvB)
		if err != nil {
			//handle error
		}
		input := string(rcvB[:n])
		input = cleanCRLF(input)
		fmt.Println(input)
		words := strings.Split(input, " ")
		if words[0] == "QUIT" {
			endConnection(conn)
			return
		}
		inputHandler(words, data_conn_ch, conn)
	}
}

func inputHandler(input []string, data_conn_ch chan data_conn_info, conn net.Conn) {
	if input[0] == "USER" {
		loginHandler(input, conn)
	} else if strings.Compare(input[0], "SYST") == 0 {
		systHandler(conn)
	} else if strings.Compare(input[0], "PWD") == 0 {
		pwdHandler(input, conn)
	} else if strings.Compare(input[0], "CWD") == 0 {
		cwdHandler(input, conn)
	} else if strings.Compare(input[0], "PASV") == 0 {
		passiveHandler(conn, data_conn_ch)
	} else if strings.Compare(input[0], "LIST") == 0 {
		lsHandler(conn, data_conn_ch)
	} else if strings.Compare(input[0], "RETR") == 0 {
		getHandler(input, conn, data_conn_ch)
	} else if strings.Compare(input[0], "TYPE") == 0 {
		typeHandler(conn)
	} else {
		syntaxErrHandler(conn)
	}
}

func successfulConnection(conn net.Conn) {
	response := generateMsgStr(220, "You are connected!")
	conn.Write([]byte(response))
}

func loginHandler(input []string, conn net.Conn) {
	var response string
	if strings.Compare(input[1], "anonymous") == 0 {
		response = generateMsgStr(230, "Login successful!")
	} else {
		response = generateMsgStr(530, "Login unsuccessful")
	}
	conn.Write([]byte(response))
}

func pwdHandler(input []string, conn net.Conn) {
	dir, err := os.Getwd()
	if err != nil {
		//handle error
	}
	response := generateMsgStr(257, dir)
	conn.Write([]byte(response))
}

// BUG: If two users are on separate threads and one of
//      them switches directories, they both switch directories...
func cwdHandler(input []string, conn net.Conn) {
	dir := input[1]
	err := os.Chdir(dir)
	if err != nil {
		conn.Write([]byte(generateMsgStr(550, "Change directory failed")))
		return
	}
	conn.Write([]byte(generateMsgStr(250, "Success")))

}

func portAddressStr(host_port string) (addr_str string) {
	host_port_sp := strings.Split(host_port, ",") //h1, h2, h3, h4, p1, p2
	p1, _ := strconv.Atoi(host_port_sp[4])
	p1 *= 256
	p2, _ := strconv.Atoi(host_port_sp[5])
	port_str := strconv.Itoa(p1 + p2)
	addr_str = strings.Join(host_port_sp[:4], ".") + ":" + port_str
	return
}

func dataHandler(ln net.Listener, data_conn_ch chan data_conn_info) {
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
		sendFileList(conn)
		data_conn_ch <- data_conn_info{code: 1, info: ""}
	} else {
		fname := data_conn_req.info
		fmt.Println(fname)
		data_conn_ch <- sendFile(conn, fname)
	}
}

func portSplitStr(address_str string) string {
	addr_port := strings.Split(address_str, ":")
	address := strings.Join(strings.Split(addr_port[0], "."), ",")
	port_int, _ := strconv.Atoi(addr_port[1])
	port16 := uint16(port_int)
	p1 := strconv.Itoa(int(port16 >> 8))
	p2 := strconv.Itoa(int(port16 << 8 >> 8))
	return address + "," + p1 + "," + p2
}

func lsHandler(conn net.Conn, data_conn_ch chan data_conn_info) {
	var response string
	response = generateMsgStr(150, "File list send starting")
	conn.Write([]byte(response))
	data_conn_ch <- data_conn_info{code: 0, info: ""}
	<-data_conn_ch
	response = generateMsgStr(226, "File list send complete")
	conn.Write([]byte(response))
}

func getHandler(input []string, conn net.Conn, data_conn_ch chan data_conn_info) {
	var response string
	if len(input[1]) <= 1 || len(input[1]) == 0 {
		response = generateMsgStr(450, "No file specified for retrieval")
		conn.Write([]byte(response))
		return
	}
	response = generateMsgStr(125, "Attempting file retrieval")
	conn.Write([]byte(response))
	data_conn_ch <- data_conn_info{code: 1, info: input[1]}
	data_conn_resp := <-data_conn_ch
	if data_conn_resp.code == -1 {
		response = generateMsgStr(550, "Could not retrieve file")
		conn.Write([]byte(response))
		return
	} else {
		response = generateMsgStr(250, "File sent successfully")
		conn.Write([]byte(response))
		return
	}
}

func getFilesDir(dir string) string {
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

func sendFileList(conn net.Conn) {
	file_list_str := getFilesDir(".")
	conn.Write([]byte(file_list_str))
}

func endConnection(conn net.Conn) {
	defer conn.Close()
	response := generateMsgStr(221, "Quitting!")
	conn.Write([]byte(response))
}

func syntaxErrHandler(conn net.Conn) {
	response := generateMsgStr(500, "Syntax error")
	conn.Write([]byte(response))
}

func systHandler(conn net.Conn) {
	response := generateMsgStr(215, "Special FTP Server :)")
	conn.Write([]byte(response))
}

func cleanCRLF(s string) string {
	CRLF := "\r\n"
	return strings.Replace(s, CRLF, "", -1)
}

func sendFile(conn net.Conn, fname string) data_conn_info {
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		return data_conn_info{code: -1, info: ""}
	}
	fdata, err := ioutil.ReadFile(fname)
	if err != nil {
		return data_conn_info{code: -1, info: ""}
	}
	n, err := conn.Write(fdata)
	fmt.Println(n)
	if err != nil {
		fmt.Println(err)
	}
	return data_conn_info{code: 1, info: ""}
}

func typeHandler(conn net.Conn) {
	response := generateMsgStr(200, "Type switch successful")
	conn.Write([]byte(response))
}

func Run() {
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
		go connectionHandler(conn)
	}
}

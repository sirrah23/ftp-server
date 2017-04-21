package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type dataConnInfo struct {
	code int
	info string
}

type ftpSession struct {
	dir string
}

func generateMsgStr(code int, msg string) string {
	//TODO: Sprintf instead
	return strconv.Itoa(code) + " " + msg + "\r\n"
}

func startFTPSession() ftpSession {
	dir, err := os.Getwd()
	if err != nil {
		//handle err
	}
	return ftpSession{dir: dir}
}

func (f *ftpSession) passiveHandler(conn net.Conn, data_conn_ch chan dataConnInfo) {
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
	ftpS := startFTPSession()
	rcvB := make([]byte, 1024)
	data_conn_ch := make(chan dataConnInfo)
	for {
		n, err := conn.Read(rcvB)
		if err != nil {
			//handle error
		}
		input := string(rcvB[:n])
		input = cleanCRLF(input)
		words := strings.Split(input, " ")
		if words[0] == "QUIT" {
			endConnection(conn)
			return
		}
		ftpS.inputHandler(words, data_conn_ch, conn)
	}
}

func (f *ftpSession) inputHandler(input []string, data_conn_ch chan dataConnInfo, conn net.Conn) {
	if input[0] == "USER" {
		f.loginHandler(input, conn)
	} else if strings.Compare(input[0], "SYST") == 0 {
		f.systHandler(conn)
	} else if strings.Compare(input[0], "PWD") == 0 {
		f.pwdHandler(input, conn)
	} else if strings.Compare(input[0], "CWD") == 0 {
		f.cwdHandler(input, conn)
	} else if strings.Compare(input[0], "PASV") == 0 {
		f.passiveHandler(conn, data_conn_ch)
	} else if strings.Compare(input[0], "LIST") == 0 {
		f.lsHandler(conn, data_conn_ch)
	} else if strings.Compare(input[0], "RETR") == 0 {
		f.getHandler(input, conn, data_conn_ch)
	} else if strings.Compare(input[0], "TYPE") == 0 {
		f.typeHandler(conn)
	} else {
		syntaxErrHandler(conn)
	}
}

func successfulConnection(conn net.Conn) {
	response := generateMsgStr(220, "You are connected!")
	conn.Write([]byte(response))
}

func (f *ftpSession) loginHandler(input []string, conn net.Conn) {
	var response string
	if strings.Compare(input[1], "anonymous") == 0 {
		response = generateMsgStr(230, "Login successful!")
	} else {
		response = generateMsgStr(530, "Login unsuccessful")
	}
	conn.Write([]byte(response))
}

func (f *ftpSession) pwdHandler(input []string, conn net.Conn) {
	response := generateMsgStr(257, f.dir)
	conn.Write([]byte(response))
}

func file_exists(file_name string) bool {
	if _, err := os.Stat(file_name); os.IsNotExist(err) {
		return false
	}
	return true
}

func follow_path(base_dir string, rel_path string) (string, bool) {
	path_to_follow := strings.Split(rel_path, "/")
	curr_dir := base_dir
	for _, p := range path_to_follow {
		if p == ".." {
			curr_dir = filepath.Dir(curr_dir)
		} else if p != "." {
			curr_dir = curr_dir + "/" + p
		}
		if !file_exists(curr_dir) {
			return "", false
		}
	}
	return curr_dir, true
}

// TODO: filepath.Clean(...) is a thing...look into it
func (f *ftpSession) cwdHandler(input []string, conn net.Conn) {
	//Switch directly to absolute directory - if it exists
	if filepath.IsAbs(input[1]) {
		if !file_exists(input[1]) {
			conn.Write([]byte(generateMsgStr(550, "Change directory failed")))
			return
		} else {
			f.dir = input[1]
			conn.Write([]byte(generateMsgStr(250, "Success")))
			return
		}
	}
	// Follow along the path of relative directory
	curr_dir, success := follow_path(f.dir, input[1])
	if success {
		f.dir = curr_dir
		conn.Write([]byte(generateMsgStr(250, "Success")))
		return
	} else {
		conn.Write([]byte(generateMsgStr(550, "Change directory failed")))
		return
	}
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

func dataHandler(ln net.Listener, data_conn_ch chan dataConnInfo) {
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
		sendFileList(conn, data_conn_req.info)
		data_conn_ch <- dataConnInfo{code: 1, info: ""}
	} else {
		fname := data_conn_req.info
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

func (f *ftpSession) lsHandler(conn net.Conn, data_conn_ch chan dataConnInfo) {
	var response string
	response = generateMsgStr(150, "File list send starting")
	conn.Write([]byte(response))
	data_conn_ch <- dataConnInfo{code: 0, info: f.dir}
	<-data_conn_ch
	response = generateMsgStr(226, "File list send complete")
	conn.Write([]byte(response))
}

func (f *ftpSession) getHandler(input []string, conn net.Conn, data_conn_ch chan dataConnInfo) {
	var response string
	if len(input[1]) <= 1 || len(input[1]) == 0 {
		response = generateMsgStr(450, "No file specified for retrieval")
		conn.Write([]byte(response))
		return
	}
	response = generateMsgStr(125, "Attempting file retrieval")
	conn.Write([]byte(response))
	var file_to_get string
	var success bool
	if filepath.IsAbs(input[1]) {
		file_to_get = input[1]
	} else {
		file_to_get, success = follow_path(f.dir, input[1])
		if !success {
			response = generateMsgStr(550, "Could not retrieve file")
			conn.Write([]byte(response))
			return
		}
	}
	data_conn_ch <- dataConnInfo{code: 1, info: file_to_get}
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
	files, err := ioutil.ReadDir(dir)
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

func sendFileList(conn net.Conn, file string) {
	file_list_str := getFilesDir(file)
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

func (f *ftpSession) systHandler(conn net.Conn) {
	response := generateMsgStr(215, "Special FTP Server :)")
	conn.Write([]byte(response))
}

func cleanCRLF(s string) string {
	CRLF := "\r\n"
	return strings.Replace(s, CRLF, "", -1)
}

func sendFile(conn net.Conn, fname string) dataConnInfo {
	if !file_exists(fname) {
		return dataConnInfo{code: -1, info: ""}
	}
	fdata, err := ioutil.ReadFile(fname)
	if err != nil {
		return dataConnInfo{code: -1, info: ""}
	}
	_, err = conn.Write(fdata)
	if err != nil {
		fmt.Println(err)
	}
	return dataConnInfo{code: 1, info: ""}
}

func (f *ftpSession) typeHandler(conn net.Conn) {
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

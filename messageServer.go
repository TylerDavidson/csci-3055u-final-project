package main

import (
	"fmt"
	"net"
	"os"
	"encoding/json"
)

type Message struct{
	UserId int
	ConversationId int
	Text string
}

const (
	CONN_HOST = "localhost"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

func main(){
	l, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)

	if err != nil{
		fmt.Println("Error listening")
		os.Exit(1)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil{
			fmt.Println("Error accepting")
			os.Exit(1)
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn){
	
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var msg Message

	for{
		err := decoder.Decode(&msg)

		if err != nil {
			fmt.Println("Error decoding message")
			conn.Close()
			return
		} else {
			fmt.Println(msg)
			encoder.Encode(msg)
		}
	}
}

/*
func handleRequest(conn net.Conn){
	buf := make([]byte, 1024)

	reqLen, err := conn.Read(buf)

	fmt.Println(reqLen)

	if err != nil {
		fmt.Println("Error reading")
	}

	conn.Write([]byte("Message recived."))

	conn.Close()
}
*/
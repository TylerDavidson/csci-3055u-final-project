package main

import (
	"fmt"
	"net"
	"os"
	"encoding/json"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

type Message struct{
	UserId int
	UserName string
	ConversationId int
	Text string
}

//struct for storing a user id and it's corresponding
//channel (for sending to handleRequest)
type clientConn struct{
	UserId int
	channel chan Message
}

//array to hold all active client connection channels
var clientConns []clientConn

func main(){
	l, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)

	if err != nil{
		fmt.Println("Error listening")
		os.Exit(1)
	}

	defer l.Close()

	//clientConns := []clientConn{}

	globalMessageChannel := make(chan Message, 10)

	go sendMessage(globalMessageChannel)

	//temp id var
	tempId := 0

	for {
		conn, err := l.Accept()
		if err != nil{
			fmt.Println("Error accepting")
			os.Exit(1)
		}

		decoder := json.NewDecoder(conn)
		encoder := json.NewEncoder(conn)

		var msg Message

		decoder.Decode(&msg)

		if err != nil {
			fmt.Println("Error decoding message")
			conn.Close()
		} else {
			//set user id
			if msg.UserId == 0 && msg.Text == "/new"{
				tempId += 1
				msg.UserId = tempId
				encoder.Encode(msg)

				fmt.Println(msg)

				//set up channels
				clientChannel := make(chan Message, 3)

				clientConns = append(clientConns, clientConn{msg.UserId, clientChannel})

				go handleRequest(conn, msg.UserId, clientChannel, globalMessageChannel)
			}
		}
		
	}
}

//Go function that handles sending messages to the other connections
//The channel that this function uses is passed to the handleRequest
// functions so that they can send messages on the channel to this function
func sendMessage(c chan Message){
	for{
		msg := <- c

		fmt.Println("Sending message:")
		fmt.Println(msg)

		for _, cc := range clientConns{
			if cc.UserId != msg.UserId{
				cc.channel <- msg
			}
		}
	}
}

//handles sending and receiving to the clinet
func handleRequest(conn net.Conn, UserId int, receiveChan chan Message, sendChan chan Message){
	
	//decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	connChannel := make(chan Message, 3)
	go handleConnReceive(conn, connChannel)

	for{
		select{
			case msg := <- connChannel:
				if msg.Text == "/Disconnect"{
					fmt.Println("Conn lost")
					//handle disconnect (potentially will cause problems)
					//clientConns = append(
					conn.Close()
					return
				} else {
					sendChan <- msg
				}
			case msg := <- receiveChan:
				encoder.Encode(msg)
		}
	}
}

func handleConnReceive(conn net.Conn, c chan Message){
	decoder := json.NewDecoder(conn)

	var msg Message

	for{
		err := decoder.Decode(&msg)

		if err != nil{
			c <- Message{0, "", 0, "/Disconnect"}
		} else {
			c <- msg
		}
	}
}
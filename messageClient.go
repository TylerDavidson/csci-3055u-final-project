package main

import (
	"fmt"
	"bufio"
	"os"
	"net"
	"os/exec"
	"strings"
	"strconv"
	"encoding/json"
)

type Message struct{
	UserId int
	UserName string
	ConversationId int
	Text string
}

type Size struct{
	height int
	width int
}

const (
	SERVER_CONN_TYPE = "tcp"
	SERVER_ADDRESS = "localhost:3333"

	CL_UP1 = "\033[0A"
	CL_DN1 = "\033[0B"

	CL_CLEAR = "\033[2J"
	CL_ORIGIN = "\033[0;0f"
	CL_ERASE_LINE = "\033[K"

	CL_WHITE = "\033[0;37m"
	CL_BLUE = "\033[1;34m"
	CL_RED = "\033[0;31m"
	CL_LIGHT_RED = "\033[1;31m"
)

var userId = 0
var userName = "Anon"

func getTerminalSize() Size{
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()

	//fmt.Println("out: " + string(out))

	if err != nil {
		fmt.Println("failed to get terminal size")
		os.Exit(1)
	}

	vals := strings.Split(string(out), " ")
	h, err := strconv.Atoi(vals[0])
	w, err := strconv.Atoi(vals[1])

	return Size{h, w}
}

func clListener(c chan string){
	scanner := bufio.NewScanner(os.Stdin)
	for{
		scanner.Scan()
		c <- scanner.Text()
	}
}

func printText(screenSize Size, text string, colour string){
	fmt.Print(setClCursorPos(screenSize.height - 1, 0) + colour +  text)
	fmt.Print(setClCursorPos(screenSize.height, 0) + CL_WHITE + "> ")
}

func setClCursorPos(line int, offset int) string{
	return "\033[" + strconv.Itoa(line) + ";" + strconv.Itoa(offset) + "f"
}

// go function (handles the command line)
func clWorker(c chan Message){
	screenSize := getTerminalSize()

	l := make(chan string)
	go clListener(l)

	for{
		select {
			case message := <- c: //main listener (incoming messages)
				fmt.Println()
				text := ""
				colour := CL_BLUE
				if message.UserId != 0 {
					text = message.UserName
					colour = CL_WHITE
					if message.UserName == "Anon" {
						text += " " + strconv.Itoa(message.UserId)
					}
					text += ": "
				}
				text += message.Text
				printText(screenSize, text, colour)

			case text := <- l: // command line listener (outgoing messages)
				printText(screenSize, "  " + text, CL_LIGHT_RED)
				message := Message{userId, userName, 0, text}
				c <- message
		}
	}
}

func serverConnection(c chan Message){
	send := make(chan Message)
	receive := make(chan Message)

	conn, err := net.Dial(SERVER_CONN_TYPE, SERVER_ADDRESS)

	if err != nil{
		//handle
	}

	go serverSendMsg(conn, send)
	go serverReceiveMsg(conn, receive)

	//request new userId from server
	send <- Message{userId, userName, 0, "/new"}
	//Send return message to main
	c <- <- receive

	for{
		select{
			case message := <- c:
				send <- message
			case message := <- receive:
				c <- message
		}
	}
}

func serverSendMsg(conn net.Conn, c chan Message){
	encoder := json.NewEncoder(conn)

	for{
		msg := <- c
		encoder.Encode(msg)
	}
}

func serverReceiveMsg(conn net.Conn, c chan Message){
	decoder := json.NewDecoder(conn)
	var msg Message

	for{
		err := decoder.Decode(&msg)

		if err != nil{
			fmt.Println("Connection to server lost")
			os.Exit(1)
		} else {
			c <- msg
		}
	}
}

func main(){
	
	fmt.Println(CL_CLEAR)

	clChannel := make(chan Message)
	serverChannel := make(chan Message)

	go clWorker(clChannel)
	go serverConnection(serverChannel)

	clChannel <- Message{0, "", 0, "Testing Go based messanger"}
	clChannel <- Message{0, "", 0, "  Type '/exit' to close"}

	for{
		select{
			case message := <- clChannel:
				if message.Text == "/exit"{
					fmt.Println("Good Bye")
					os.Exit(0)
				} else {
					serverChannel <- message
				}
			case message := <- serverChannel:
				if message.Text == "/new" {
					userId = message.UserId
				} else {
					clChannel <- message
				}
		}
	}
}
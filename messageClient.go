package main

import (
	"fmt"
	"os"
	"net"
	"os/exec"
	"strings"
	"strconv"
	"encoding/json"
	"unicode/utf8"
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
	CL_FORWARD1 = "\033[1C"
	CL_BACKWARD1 = "\033[1D"

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
var tempInput = ""

// go function (handles the command line)
func clWorker(c chan Message){
	screenSize := getTerminalSize()

	l := make(chan string)
	go clListener(l)

	for{
		select {
			case message := <- c: //main listener (incoming messages)
				printMessage(screenSize, message)
			case text := <- l: // command line listener (outgoing messages)
				message := Message{userId, userName, 0, text}

				if text[0] != '/'{
					printMessage(screenSize, message)
				}
				c <- message
		}
	}
}

func clListener(c chan string){
	var b []byte = make([]byte, 1)

	for{
		os.Stdin.Read(b)
		rune,_ := utf8.DecodeRune(b)
		if rune == '\n'{ //If enter key pressed
			c <- tempInput
			tempInput = ""
		} else if byte(rune) == byte(127) { //If backspace
			fmt.Print(CL_BACKWARD1 + CL_ERASE_LINE)
			tempInput = tempInput[:len(tempInput) - 1]
		} else if byte(rune) >= byte(32) && byte(rune) <= byte(126) { //If non-special char
			s := fmt.Sprintf("%c", rune)
			tempInput += s
			fmt.Print(s)
		}
	}
}

func printMessage(screenSize Size, message Message){
	text := message.Text

	if(message.UserId == 0){
		printText(screenSize, text, CL_BLUE)
	} else {
		text = message.UserName + " (" + strconv.Itoa(message.UserId) + "): " + text
		
		if(message.UserId != userId){
			printText(screenSize, text, CL_WHITE)
		} else {
			printText(screenSize, text, CL_LIGHT_RED)
		}
	}
}

func printText(screenSize Size, text string, colour string){
	fmt.Println()
	fmt.Print(setClCursorPos(screenSize.height - 1, 0) + colour +  text)
	fmt.Print(setClCursorPos(screenSize.height, 0) + CL_WHITE + "> " + tempInput + CL_ERASE_LINE)
}

func setClCursorPos(line int, offset int) string{
	return "\033[" + strconv.Itoa(line) + ";" + strconv.Itoa(offset) + "f"
}

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

	//Allow reading in single chars
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

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
					fmt.Println("\nGood Bye")
					return
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
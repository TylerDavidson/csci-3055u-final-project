package main

import (
	"fmt"
	"bufio"
	"os"
	"os/exec"
	"strings"
	"strconv"
)

type Message struct{
	id int
	text string
}

type Size struct{
	height int
	width int
}

const CL_UP1 = "\033[0A"
const CL_DN1 = "\033[0B"

const CL_CLEAR = "\033[2J"
const CL_ORIGIN = "\033[0;0f"
const CL_ERASE_LINE = "\033[K"

const CL_WHITE = "\033[0;37m"
const CL_BLUE = "\033[1;34m"
const CL_RED = "\033[0;31m"
const CL_LIGHT_RED = "\033[1;31m"

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


// go function (handles the command line)
func clWorker(c chan Message){
	screenSize := getTerminalSize()

	l := make(chan string)
	go clListener(l)

	for{
		select {
			case message := <- c:
				fmt.Println()
				printText(screenSize, message.text, CL_BLUE)
			case text := <- l:
				printText(screenSize, "  " + text, CL_LIGHT_RED)
				c <- Message{0, text}
		}
	}
}

func setClCursorPos(line int, offset int) string{
	return "\033[" + strconv.Itoa(line) + ";" + strconv.Itoa(offset) + "f"
}

func main(){
	
	fmt.Println(CL_CLEAR)

	clChannel := make(chan Message)

	go clWorker(clChannel)

	clChannel <- Message{0, "Testing Go"}
	clChannel <- Message{0, "  Type '/exit' to close"}

	for{
		message := <- clChannel

		if message.text == "/exit"{
			fmt.Println("Good Bye")
			os.Exit(0)
		}
	}
}
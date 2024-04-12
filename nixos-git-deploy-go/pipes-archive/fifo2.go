package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"syscall"
	//"time"
)

var pipeFile = "pipe.log"

func main() {
// Delete existing pipes
fmt.Println("Cleanup existing FIFO file")
os.Remove(filePath)

// Create pipe
fmt.Println("Creating " + filePath + " FIFO file")
err := syscall.Mkfifo(filePath, 0640)
if err != nil {
	fmt.Println("Failed to create pipe")
	panic(err)
}

// Open pipe for read only
fmt.Println("Starting read operation")
pipe, err := os.OpenFile(filePath, os.O_RDONLY, 0640)
if err != nil {
	fmt.Println("Couldn't open pipe with error: ", err)
}
defer pipe.Close()

// Read the content of named pipe
reader := bufio.NewReader(pipe)
fmt.Println("READER >> created")

// Infinite loop
for {
	line, err := reader.ReadBytes('\n')
	// Close the pipe once EOF is reached
	if err != nil {
		fmt.Println("FINISHED!")
		os.Exit(0)
	}

	// Remove new line char
	nline := string(line)
	nline = strings.TrimRight(nline, "\r\n")
	fmt.Printf("READER >> reading line: %+v\n", nline)

}
}


package main


import (
//	"bufio"
	"fmt"
	"log"
	"os"
//	"syscall"
	"time"
)

const pipeFile = "pipe.log"

func main(){
	scheduleWrite()
}

func scheduleWrite() {
	fmt.Println("start schedule writing.")
	f, err := os.OpenFile(pipeFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	i := 0
	for {
		fmt.Println("write string to named pipe file.")
		f.WriteString(fmt.Sprintf("test write times:%d\n", i))
		i++
		time.Sleep(time.Second)
	}
}
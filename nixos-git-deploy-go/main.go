package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	//"log"
	"time"
	//"syscall"
	"os/exec"
	"strconv"
	"encoding/json"
	"golang.org/x/sys/unix"
	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

var gitDirectory = "/home/spiderunderurbed/.config/nixos-git-deploy/"
var watchedFiles = make(map[string]bool)

type Config struct {
	UserAllowed string `json:"UserAllowed"`
	FirstTime   string `json:"FirstTime"`
}

// type Settings struct {
// 	UserAllowed: "n",
// 	firstTi 
// }
// Function to modify the file within the Git repository
func modifyFile(filename string) error {
	// Splits the filename by '/' to get all elements in a path
	parts := strings.Split(filename, "/")
	// Once you split it, you can find the ACTUAL filename
	fileName := parts[len(parts)-1]

	// Gets the absolute path of the coorosponding git file
	gitFilePath := filepath.Join(gitDirectory, fileName)

	// Read the content of the original file
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// TODO: See if go lets you append the newly added content to the file instead of overriting the whole thing with itself?
	modifiedContent := string(content)

	// Write the modified contents back to the file in the Git directory
	if err := ioutil.WriteFile(gitFilePath, []byte(modifiedContent), 0644); err != nil {
		return err
	}

	// Open the Git repository
	r, err := git.PlainOpen(gitDirectory)
	if err != nil {
		return err
	}

	// Get the worktree
	worktree, err := r.Worktree()
	if err != nil {
		return err
	}

	// Add the modified file to the Git staging area
	if _, err := worktree.Add(fileName); err != nil {
		return err
	}

	//fmt.Printf("File %s has been successfully modified in the Git repository\n", fileName)
	return nil
}

// Function to add files to Git repository
func addFilesToGit(files []string, r *git.Repository) error {
	worktree, err := r.Worktree()
	if err != nil {
		return err
	}

	for _, file := range files {
		// Check if the file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Println("File does not exist:", file)
			continue
		}

		// Determine the filename without the path
		fileName := filepath.Base(file)

		// Destination path in the Git directory
		destination := filepath.Join(gitDirectory, fileName)

		// Copy the file to the Git directory
		if err := copyFile(file, destination); err != nil {
			return err
		}

		// Add the file to the Git repository
		_, err := worktree.Add(fileName)
		if err != nil {
			return err
		}
	}

	return nil
}

// Function to watch for file changes
func watchChanges(filename string) {
	// Check if the file is already being watched
	if _, ok := watchedFiles[filename]; ok {
		fmt.Printf("File %s is already being watched\n", filename)
		return
	}

	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating watcher for file %s: %v\n", filename, err)
		return
	}
	defer watcher.Close()

	// Add the file to the watcher
	err = watcher.Add(filename)
	if err != nil {
		fmt.Printf("Error adding file %s to watcher: %v\n", filename, err)
		return
	}
	watchedFiles[filename] = true

	// Start watching for events
	fmt.Printf("Watching for changes in file: %s\n", filename)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				//fmt.Printf("File %s has been modified\n", filename)
				err := modifyFile(filename)
				if err != nil {
					fmt.Println("ERROR:", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Error watching file %s: %v\n", filename, err)
		}
	}
}

// Function to copy a file
func copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
func keepAlive(f *os.File, origin string) {
    i := 0
    for {
        // Write string to the named pipe file.
        _, err := f.WriteString(fmt.Sprintf("%s: test write times: %d\n", origin, i))
        if err != nil {
            fmt.Printf("Error writing to file: %v\n", err)
            return
        }
        i++
        time.Sleep(time.Second)
    }
}


func processChildArgs(args []string, messages chan string){
	//fmt.Println("child: " + strings.Join(args, " "))
	//fmt.Println(args)
	//messages <- "test"
}
func processParentArgs(args []string, messages chan string){
	//fmt.Println("parent: " + strings.Join(args, " "))
	if (args[0] == "watch"){
		messages <- "responding " + args[1]
		//fmt.Println("+"+args[1]+"+")
		go watchChanges(args[1])
	}
	//fmt.Println(args)
	//messages <- "test"
}

func Reader(pipeFile string, origin string, messages chan string) {
    // Open the named pipe for reading
    pipe, err := os.Open(pipeFile)
    if os.IsNotExist(err) {
        fmt.Println("Named pipe '%s' does not exist", pipeFile)
		return
    } else if os.IsPermission(err) {
        fmt.Println("Insufficient permissions to read named pipe '%s': %s", pipeFile, err)
		return
    } else if err != nil {
        fmt.Println("Error while opening named pipe '%s': %s", pipeFile, err)
		return
    }
    defer pipe.Close()

    // Infinite loop for reading from the named pipe
    //messages <- "We received"
    for {
        // Read from the named pipe
		data := make([]byte, 1024) // Read buffer size
		n, err := pipe.Read(data)
		if err != nil {
			fmt.Println("Error reading from named pipe '%s': %s", pipeFile, err)
			break
		}
		input := strings.TrimSpace(string(data[:n]))
		args := strings.Split(input, " ")
		//fmt.Println(args[0])
		if (args[0] == "child:"){
			args = args[1:] 
			//fmt.Println("child message" + args[0])
			processChildArgs(args, messages)
		} else if (args[0] == "parent:"){
			args = args[1:] 
			//fmt.Println("parent message" + args[0])
			processParentArgs(args, messages)
		}
		//#fmt.Println("data " + string(data[:n]))
        // Process the read data
        //processData(data[:n], origin, messages)
    }
}

func writer(pipeFile string, origin string, messages chan string) *os.File {
    // Open the file
    f, err := os.OpenFile(pipeFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
    if err != nil {
        fmt.Printf("Error opening file: %v\n", err)
        return nil // Return nil if there's an error
    }
    
    // Continuously wait for messages and write them to the file
	for msg := range messages {
		//fmt.Println("TEST")
		//fmt.Println(msg)
		//fmt.Printf(fmt.Sprintf("%s: %s\n", origin, msg))
        _, err := f.WriteString(fmt.Sprintf("%s: %s\n", origin, msg + "\n"))
        if err != nil {
            fmt.Printf("Error writing to file: %v\n", err)
            break // Break the loop if there's an error
        }
    }
    
    // Close the file before returning
    f.Close()
    
    // Return the opened file
    return f
} 

func runChildProcess() {
	// defer func() {
    //     if r := recover(); r != nil {
    //         fmt.Println("Recovered from panic:", r)
    //         // Add any cleanup or error handling logic here
    //     }
    // }()

	unix.Setpgid(0, 0)
	//var stdout bytes.Buffer
    // Function to be executed in child process
	// messages := make(chan string, 10000)
	//f := writer("detach.log", "child")
	// fmt.Println("TEST")

	messages := make(chan string, 10000)
	 go Reader("recede.log", "child", messages)
	 go writer("detach.log", "child", messages)

	//for {}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	//keepAlive(f, "parent")
    fmt.Println("Running in child process")
        // Sleep for 100 seconds
		//time.Sleep(100 * time.Second)
}

func cleanup(messages chan string) {
	// Handle SIGINT (Ctrl+C) signal to perform cleanup before exiting
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-c

	// Close the messages channel to stop writer goroutine
	close(messages)

	// Perform cleanup actions
	fmt.Println("Performing cleanup actions...")
	// Add your cleanup code here

	// Exit the program gracefully
	os.Exit(0)
}

func killProcess(pid int) error {
    // Find the process by its PID
    proc, err := os.FindProcess(pid)
    if err != nil {
        return fmt.Errorf("error finding process: %v", err)
    }
    
    // Check if the process is nil
    if proc == nil {
        return fmt.Errorf("process with PID %d not found", pid)
    }

    // Send SIGTERM signal to the process
    err = proc.Signal(syscall.SIGTERM)
    if err != nil {
        return fmt.Errorf("error sending SIGTERM signal: %v", err)
    }

    return nil
}
func main() {
    
    //Check if there are any command-line arguments
    if len(os.Args) > 1 && os.Args[1] == "child" {
        // This is the child process
        runChildProcess()
        return
    }

	reader := bufio.NewReader(os.Stdin)

	//type Settings struc {

		configFile := Config{
			UserAllowed: "y",
			FirstTime:   "y",
		}
	
		rawConfig, err := os.Open("./config.json")
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer rawConfig.Close()

		formattedConfig, err := ioutil.ReadAll(rawConfig)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
	
		err = json.Unmarshal(formattedConfig, &configFile)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}
	
		// Reset file cursor to beginning
		_, err = rawConfig.Seek(0, 0)
		if err != nil {
			fmt.Println("Error seeking file:", err)
			return
		}
	
		fomated_config, err := ioutil.ReadAll(rawConfig)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
	
		var settings Config
		err = json.Unmarshal(fomated_config, &settings)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}
	
		//fmt.Println(settings)
	//fmt.Print(settings)
	if (settings.FirstTime == "y"){
	//	print("user not allowed")
	configFile.FirstTime = "n"
	fmt.Print(" Hello! This is nixos-git-deploy.\n If allowed, we will spawn backround processes to\n watch for file changes if allowed, and a backround\n process so that if in the event of a crash or deletion\n of the main files the file watchers will be\n deleted, are you ok with this?[Y/n] ")

	userallow, _ := reader.ReadString('\n')
	userallow = strings.TrimSpace(userallow)

	if (userallow == "n"){
		//fmt.Println("User inputted ")
		configFile.UserAllowed = "n"
	}
		jsonData, err := json.Marshal(configFile)
		//fmt.Println(jsonData)
		if err != nil {
			fmt.Println("Error with JSON:", err)
		}

		err = ioutil.WriteFile("./config.json", []byte(jsonData), 0644)
		if err != nil {
			fmt.Println("Error with file:", err)
		}
		//err = ioutil.WriteFile("./config.json", []byte(json.Unmarshal([]bytejson.Marshal(configFile))), 0644)
	} else {

	}
	if _, err := os.Stat("detach.log"); err == nil {
		//fmt.Println("Named pipe", "detach.log", "already exists.")
	} else {
		err := syscall.Mkfifo("detach.log", 0600)
		if err != nil {
			fmt.Println("Error wit pipe file:", err)
		}
	}

	if _, err := os.Stat("recede.log"); err == nil {
		//fmt.Println("Named pipe", "recede.log", "already exists.")
	} else {
		err := syscall.Mkfifo("recede.log", 0600)
		if err != nil {
			fmt.Println("Error wit pipe file:", err)
		}
	}

	fmt.Print("\n")
    cmd := exec.Command("./nixos-git-deploy-go", "child")
	// cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	// Start the child process
	// err := cmd.Start()
	// if err != nil {
	// 	fmt.Println("Error starting child process:", err)
	// 	return
	// }

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

    err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting child process:", err)
		return
	}
	//go killProcess(cmd.Process.Pid)
    	if err != nil {
		fmt.Println("Error starting child process:", err)
		return
	}
	messages := make(chan string)
	
	go cleanup(messages)
	fmt.Println(strconv.Itoa(cmd.Process.Pid))
	
	go writer("recede.log", "parent", messages)
	go Reader("detach.log", "parent", messages)

	for {
		options := []string{"init", "apply", "status", "remove", "upgrade", "add-automatic", "add", "remote-init"}

		fmt.Println("What do you want to do?")
		for i, option := range options {
			fmt.Printf("%d. %s\n", i+1, option)
		}

		fmt.Print("Enter your choice (1-8): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		index := -1
		fmt.Sscanf(choice, "%d", &index)
		if index < 1 || index > len(options) {
			fmt.Println("Invalid choice, please try again.")
			continue
		}

		switch options[index-1] {
		case "init":
			if !ifDirectoryExists(gitDirectory + "/.git") {
				_, err := git.PlainInit(gitDirectory, false)
				if err != nil {
					fmt.Println("Error initializing git repository:", err)
					continue
				}
				fmt.Println("Initialized Git repository.")
				fmt.Print("Enter the remote (WE ONLY SUPPORT SSH): ")
				remote, _ := reader.ReadString('\n')
				remote = strings.TrimSpace(remote)
				r, err := git.PlainOpen(gitDirectory)
				if err != nil {
					fmt.Println("Error opening git repository:", err)
					continue
				}
				_, err = r.CreateRemote(&config.RemoteConfig{
					Name: "origin",
					URLs: []string{remote},
				})
				if err != nil {
					fmt.Println("Error adding remote:", err)
				}
			} else {
				fmt.Println("Git repository already initialized.")
			}
		case "remote-init":
			fmt.Print("Enter the remote (SSH URL): ")
			remote, _ := reader.ReadString('\n')
			remote = strings.TrimSpace(remote)
			r, err := git.PlainOpen(gitDirectory)
			if err != nil {
				fmt.Println("Error opening git repository:", err)
				continue
			}
			remoteConf := &config.RemoteConfig{
				Name: "origin",
				URLs: []string{remote},
			}
			err = r.DeleteRemote("origin")
			if err != nil {
				fmt.Println("Error deleting remote:", err)
				continue
			}
			_, err = r.CreateRemote(remoteConf)
			if err != nil {
				fmt.Println("Error adding remote:", err)
			}
		case "apply":
			// Add your logic for "apply" here
		case "remove":
			// Add your logic for "remove" here
		case "upgrade":
			// Add your logic for "upgrade" here
		case "status":
			// Add your logic for "status" here
		case "add-automatic":
			// Add logic for adding files
			fmt.Print("Enter the path of the file(s) you want to add (comma-separated): ")
			filesInput, _ := reader.ReadString('\n')
			filesInput = strings.TrimSpace(filesInput)
			files := strings.Split(filesInput, ",")
			//fmt.Println("test")
			if git, err := git.PlainOpen(gitDirectory); err == nil {
				go func() {
					if err := addFilesToGit(files, git); err != nil {
						fmt.Println("Error adding files to Git:", err)
					} else {
						fmt.Printf("Added %d file(s) to Git\n", len(files))
					}
				}()
				//fmt.Println("sending")
				// Start file watchers for added files in separate goroutines
				for _, file := range files {
					//fmt.Println("sending message")
					messages <- "watch " + file
					//messages <- "test " + file
					//go watchChanges(file)
					//_, err := f.WriteString(fmt.Sprintf(file, "parent"))
					if err != nil {
						fmt.Println("Error sending message:", err)
					} 
					//fmt.Println("finished")
				}
			} else {
				fmt.Println("Error opening Git repository:", err)
			}
		}
	}
}

// Function to create directory if it doesn't exist
func ensureDirectoryExists(directory string) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, 0755)
	}
}

// Function to check if directory exists
func ifDirectoryExists(directory string) bool {
	_, err := os.Stat(directory)
	return !os.IsNotExist(err)
}

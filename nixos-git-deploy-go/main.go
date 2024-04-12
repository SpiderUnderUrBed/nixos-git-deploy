package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"log"
	"time"
	//"syscall"
	"os/exec"
	//"strconv"
	"encoding/json"
	"golang.org/x/sys/unix"
	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

var gitDirectory = "/home/spiderunderurbed/.config/nixos-git-deploy/"
var watchedFiles = make(map[string]bool)

type Config struct {
	UserAllowed string `json:"userallowed"`
	FirstTime   string `json:"firstTime"`
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
func Reader(pipeFile string){
// Delete existing pipes
//fmt.Println("Cleanup existing FIFO file")
os.Remove(pipeFile)

// Create pipe
//fmt.Println("Creating " + pipeFile + " FIFO file")
err := syscall.Mkfifo(pipeFile, 0640)
if err != nil {
	fmt.Println("Failed to create pipe")
	panic(err)
}

// Open pipe for read only
//fmt.Println("Starting read operation")
pipe, err := os.OpenFile(pipeFile, os.O_RDONLY, 0640)
if err != nil {
	fmt.Println("Couldn't open pipe with error: ", err)
}
defer pipe.Close()

// Read the content of named pipe
reader := bufio.NewReader(pipe)
//fmt.Println("READER >> created")

// Infinite loop
for {
	line, err := reader.ReadBytes('\n')
	// Close the pipe once EOF is reached
	if err != nil {
		fmt.Println("exited!")
		os.Exit(0)
	}

	// Remove new line char
	nline := string(line)
	nline = strings.TrimRight(nline, "\r\n")
	//	fmt.Printf("READER >> reading line: %+v\n", nline)
}
}

func writer(pipeFile string){
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
func runChildProcess() {
	unix.Setpgid(0, 0)
    // Function to be executed in child process
	go writer("pipe.log")
    fmt.Println("Running in child process")
        // Sleep for 100 seconds
    time.Sleep(100 * time.Second)
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

	//}
	configFile := Config{
		UserAllowed: "n",
		FirstTime:   "y",
	}

	rawConfig, err := os.Open("config.json")
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

	fomated_config, _ := ioutil.ReadAll(rawConfig)
	var settings Config 
	err = json.Unmarshal(fomated_config, &settings)
	
	fmt.Print(" Hello! This is nixos-git-deploy.\n If allowed, we will spawn backround processes to\n watch for file changes if allowed, and a backround\n process so that if in the event of a crash or deletion\n of the main files the file watchers will be\n deleted, are you ok with this?[Y/n] ")

	userallow, _ := reader.ReadString('\n')
	userallow = strings.TrimSpace(userallow)
	if (userallow == "n"){
		jsonData, err := json.Marshal(configFile)
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

	fmt.Print("\n")
    cmd := exec.Command("./nixos-git-deploy-go", "child")
    cmd.Start()
    //fmt.Println(strconv.Itoa(cmd.Process.Pid))

	go Reader("pipe.log")

	//fmt.Sscanf("testing", "testing")
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

			if git, err := git.PlainOpen(gitDirectory); err == nil {
				go func() {
					if err := addFilesToGit(files, git); err != nil {
						fmt.Println("Error adding files to Git:", err)
					} else {
						fmt.Printf("Added %d file(s) to Git\n", len(files))
					}
				}()

				// Start file watchers for added files in separate goroutines
				for _, file := range files {
					go watchChanges(file)
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
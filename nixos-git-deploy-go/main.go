package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	//"os/signal"
	"strings"
	"syscall"
	"time"
	//"sync"
	"path/filepath"
	"strconv"
	//"github.com/foresthoffman/reap"
	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

const (
    UID = 501
    GUID = 100
    )
var gitDirectory = "/home/spiderunderurbed/.config/nixos-git-deploy/"
var watchedFiles = make(map[string]bool)

func modifyFile(filename string) error {
	parts := strings.Split(filename, "/")
	fileName := parts[len(parts)-1]
	gitFilePath := filepath.Join(gitDirectory, fileName)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	modifiedContent := string(content)
	if err := ioutil.WriteFile(gitFilePath, []byte(modifiedContent), 0644); err != nil {
		return err
	}
	r, err := git.PlainOpen(gitDirectory)
	if err != nil {
		return err
	}
	worktree, err := r.Worktree()
	if err != nil {
		return err
	}
	if _, err := worktree.Add(fileName); err != nil {
		return err
	}
	return nil
}

func addFilesToGit(files []string, r *git.Repository) error {
	worktree, err := r.Worktree()
	if err != nil {
		return err
	}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Println("File does not exist:", file)
			continue
		}
		fileName := filepath.Base(file)
		destination := filepath.Join(gitDirectory, fileName)
		if err := copyFile(file, destination); err != nil {
			return err
		}
		_, err := worktree.Add(fileName)
		if err != nil {
			return err
		}
	}
	return nil
}

func watchChanges(filename string) {
	if _, ok := watchedFiles[filename]; ok {
		fmt.Printf("File %s is already being watched\n", filename)
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating watcher for file %s: %v\n", filename, err)
		return
	}
	defer watcher.Close()
	err = watcher.Add(filename)
	if err != nil {
		fmt.Printf("Error adding file %s to watcher: %v\n", filename, err)
		return
	}
	watchedFiles[filename] = true
	fmt.Printf("Watching for changes in file: %s\n", filename)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
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

func runChildProcess() {
	// Function to be executed in child process
	fmt.Println("Running in child process")
		// Sleep for 100 seconds
	time.Sleep(100 * time.Second)
}

func main() {
    // Check if there are any command-line arguments
	if len(os.Args) > 1 && os.Args[1] == "child" {
		// This is the child process
		runChildProcess()
		return
	}

	parentPID := strconv.Itoa(os.Getppid())
	fmt.Println("Parent process running with PID: " + parentPID)

	// if parentPID == "1" {
	// 	// Child process code
	// 	runChildProcess()
	// 	return
	// }

	var cred = &syscall.Credential{UID, GUID, []uint32{}, true} // Provide empty slices for Groups and SupplementaryGroups
	// the Noctty flag is used to detach the process from the parent tty
	var sysproc = &syscall.SysProcAttr{Credential: cred, Noctty: true, Setpgid: true}
	var attr = os.ProcAttr{
		Dir:   ".",
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, nil, nil},
		Sys:   sysproc,
	}
	process, err := os.StartProcess("./nixos-git-deploy-go", []string{"./nixos-git-deploy-go", "child"}, &attr)
	if err == nil {
		// Print the PID of the process
		fmt.Println("PID:", process.Pid)
		
		// It is not clear from docs, but Release actually detaches the process
		err = process.Release()
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println(err.Error())
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		options := []string{"init", "apply", "status", "remove", "upgrade", "add", "remote-init"}
		fmt.Println("What do you want to do?")
		for i, option := range options {
			fmt.Printf("%d. %s\n", i+1, option)
		}
		fmt.Print("Enter your choice (1-7): ")
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
		case "add":
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
				for _, file := range files {
					go watchChanges(file)
				}
			} else {
				fmt.Println("Error opening Git repository:", err)
			}
		}
	}
}

func ensureDirectoryExists(directory string) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, 0755)
	}
}

func ifDirectoryExists(directory string) bool {
	_, err := os.Stat(directory)
	return !os.IsNotExist(err)
}
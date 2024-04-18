package main

import (
        "bufio"
        "fmt"
        //"io"
        "io/ioutil"
        "os"
        "os/signal"
        "path/filepath"
        "strings"
        "syscall"
        // "log"
        "time"
        //"syscall"
        "encoding/json"
        "os/exec"
        "strconv"
        // "bytes"
        // "nixos-git-deploy-go/lib/"
        "nixos-git-deploy-go/lib/add"
        "nixos-git-deploy-go/lib/aged"

        // "filippo.io/age"
        // "filippo.io/age/armor"
        "golang.org/x/sys/unix"
        //"github.com/fsnotify/fsnotify"
        "github.com/go-git/go-git/v5"
        "github.com/go-git/go-git/v5/config"
        "github.com/go-git/go-git/v5/plumbing/object"
        "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

var gitDirectory = "/home/spiderunderurbed/.config/nixos-git-deploy/"

// var watchedFiles = make(map[string]bool)

type Config struct {
        UserAllowed  string   `json:"UserAllowed"`
        FirstTime    string   `json:"FirstTime"`
        FilesToWatch []string `json:"filesToWatch"`
        trackedFiles []string `json:"trackedFiles"`
		encryptedFiles []string `json:"encryptedFiles"`
}

// type Settings struct {
//      UserAllowed: "n",
//      firstTi
// }
// Function to modify the file within the Git repository

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

func processChildArgs(args []string, messages chan string) {
        fmt.Println(args)
}
func processParentArgs(args []string, messages chan string) {
        //fmt.Println("parent: " + strings.Join(args, " "))
        if args[0] == "watch" {
                messages <- "responding " + args[1]
                //fmt.Println("+"+args[1]+"+")
                go add.WatchChanges(args[1])
        } else if args[0] == "new" {
                expectedPID, err := strconv.Atoi(args[1])
                if err != nil {
                        fmt.Println("Error:", err)
                        return
                }
                currentPID := os.Getpid()
                if currentPID != expectedPID {
                        //fmt.Printf("Current PID %d does not match expected PID %d. Terminating...\n", currentPID, expectedPID)
                        os.Exit(1)
                } else {
                        //fmt.Println("Persist")
                }
        }
        //fmt.Println(args)
        //messages <- "test"
}

func Reader(pipeFile string, origin string, messages chan string, settings Config) {
        for {
                // Open the named pipe for reading
                pipe, err := os.Open(pipeFile)
                if os.IsNotExist(err) {
                        fmt.Println("Named pipe '%s' does not exist", pipeFile)
                        continue
                } else if os.IsPermission(err) {
                        fmt.Println("Insufficient permissions to read named pipe '%s': %s", pipeFile, err)
                        continue
                } else if err != nil {
                        fmt.Println("Error while opening named pipe '%s': %s", pipeFile, err)
                        continue
                }
                defer pipe.Close()

                // Infinite loop for reading from the named pipe
                //messages <- "We received"
                for {
                        // Read from the named pipe
                        data := make([]byte, 1024) // Read buffer size
                        n, err := pipe.Read(data)
                        if err != nil {
                                //fmt.Println("Error reading from named pipe '%s': %s", pipeFile, err)
                                break
                        }
                        input := strings.TrimSpace(string(data[:n]))
                        args := strings.Split(input, " ")
                        //fmt.Println(args[0])
                        if args[0] == "child:" {
                                args = args[1:]
                                //fmt.Println("child message" + args[0])
                                processChildArgs(args, messages)
                        } else if args[0] == "parent:" {
                                args = args[1:]
                                //fmt.Println("parent message" + args[0])
                                processParentArgs(args, messages)
                        }
                }
        }
}

func writer(pipeFile string, origin string, messages chan string, settings Config) *os.File {
        // Open the file
        f, err := os.OpenFile(pipeFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
        if err != nil {
                fmt.Printf("Error opening file: %v\n", err)
                return nil // Return nil if there's an error
        }

        // Continuously wait for messages and write them to the file
        for msg := range messages {
                _, err := f.WriteString(fmt.Sprintf("%s: %s\n", origin, msg+"\n"))
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

        unix.Setpgid(0, 0)

        //fmt.Println("STARTED CHILD PROCESS")

        configFile := Config{
                // UserAllowed: "y",
                // FirstTime:   "y",
                // FilesToWatch: nil,
                // trackedFiles: nil
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
        //fmt.Println("survived here")
        //fmt.Println(len(settings.FilesToWatch))
        for i := 0; i < len(settings.FilesToWatch); i++ {
                // Check if the file exists
                if _, err := os.Stat(settings.FilesToWatch[i]); os.IsNotExist(err) {
                        //fmt.Printf("File %s does not exist. Removing from settings.\n", settings.FilesToWatch[i])
                        // Remove the file from settings
                        settings.FilesToWatch = append(settings.FilesToWatch[:i], settings.FilesToWatch[i+1:]...)
                        //continue // Skip to the next iteration
                }

                // Start watching the file in a goroutine
                go add.WatchChanges(settings.FilesToWatch[i])
        }
        //fmt.Println(settings.FilesToWatch)
        jsonData, err := json.Marshal(settings)
        //fmt.Println(jsonData)
        if err != nil {
                fmt.Println("Error with JSON:", err)
        }

        err = ioutil.WriteFile("./config.json", []byte(jsonData), 0644)
        if err != nil {
                fmt.Println("Error with file:", err)
        }

        messages := make(chan string, 10000)
        go Reader("recede.log", "child", messages, settings)
        go writer("detach.log", "child", messages, settings)

        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt, syscall.SIGTERM)
        <-c
        fmt.Println("Exited")
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

        configFile := Config{
                UserAllowed:  "y",
                FirstTime:    "y",
                FilesToWatch: nil,
				encryptedFiles: nil, 
				trackedFiles: nil,

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
        if settings.FirstTime == "y" {
                //      print("user not allowed")
				configFile.FirstTime = "n"
				fmt.Print("Hello! This is nixos-git-deploy.\n Here are all the commands \n POST: it will update all your files which you are tracking \n WATCH: It will look for any file changes even after script death and update your files \n CLEAN: clean any files that arnt being tracked or watched \n status, shows you the git status of the repo and files that are being tracked and watched (not by git) \n apply: will apply all your changes to your repo starting with a pull request sent to catch errors early, then you enter details for the commit, then it pushes \n remote-init: will add a remote \n age: will add age encryption to your file you want to push to your remotes. \n destination: changes the default git repo destination \n init: initilizes the repo \n upgrade, pulls your changes \n merge: 3 way merge between the remote, your target and source \n")
				fmt.Print("Enter to continue: ")
				reader.ReadString('\n')
                fmt.Print("If allowed, we will spawn backround processes to\n watch for file changes if allowed, and a backround\n process so that if in the event of a crash or deletion\n of the main files the file watchers will be\n deleted, are you ok with this?[Y/n] ")
                userallow, _ := reader.ReadString('\n')
                userallow = strings.TrimSpace(userallow)

                if userallow == "n" {
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

        // cmd.Stdout = os.Stdout
        // cmd.Stderr = os.Stderr

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
       // fmt.Println(strconv.Itoa(cmd.Process.Pid))

        go func() {
                messages <- "new " + strconv.Itoa(cmd.Process.Pid)
        }()

        go writer("recede.log", "parent", messages, settings)
        go Reader("detach.log", "parent", messages, settings)

        messages <- "new " + strconv.Itoa(cmd.Process.Pid)

        for {
                options := []string{" init", " apply", " status", " remove", " upgrade", " watch", " clean", " post", " add", "remote-init", "age", "destination", "merge"}

                fmt.Println("What do you want to do?")
                for i, option := range options {
                        fmt.Printf("%d. %s\n", i+1, option)
                }

                fmt.Print("Enter your choice (1-11): ")
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

                        // Delete the remote if it exists
                        if err := r.DeleteRemote("origin"); err != nil && err != git.ErrRemoteNotFound {
                                fmt.Println("Error deleting remote:", err)
                                continue
                        }

                        // Create the new remote
                        _, err = r.CreateRemote(remoteConf)
                        if err != nil {
                                fmt.Println("Error adding remote:", err)
                                continue
                        }

                        // Print the remotes after adding or modifying them
                        remotes, err := r.Remotes()
                        if err != nil {
                                fmt.Println("Error getting remotes:", err)
                                continue
                        }
                        fmt.Println("Remotes:")
                        for _, remote := range remotes {
                                fmt.Printf("Name: %s, URLs: %v\n", remote.Config().Name, remote.Config().URLs)
                        }
                case "destination":

                case "age":
						// jsonData, _ := json.Marshal(configFile)

                        fmt.Print("Enter the path of the file(s) you want to add (comma-separated): ")
                        filesInput, _ := reader.ReadString('\n')
                        filesInput = strings.TrimSpace(filesInput)
                        files := strings.Split(filesInput, ",")
                        for _, file := range files {
                                aged.Encrypt(file)
								// configFile.encryptedFiles.append(configFile.encryptedFiles, file)
								configFile.encryptedFiles = append(configFile.encryptedFiles, file)
                        }
						jsonData, _ := json.Marshal(configFile)
						ioutil.WriteFile("./config.json", []byte(jsonData), 0644)
                case "apply":
                        // Open the Git repository
                        repo, err := git.PlainOpen(gitDirectory)
                        if err != nil {
                                fmt.Printf("Error opening repository: %v\n", err)
                                return
                        }

                        // Get the worktree of the repository
                        worktree, err := repo.Worktree()
                        if err != nil {
                                fmt.Printf("Error getting worktree: %v\n", err)
                                return
                        }

                        // Create SSH authentication method
                        auth, err := ssh.NewPublicKeysFromFile("git", "/home/spiderunderurbed/.ssh/id_rsa_5", "sK4mmi@=")
                        if err != nil {
                                fmt.Printf("Error creating SSH authentication: %v\n", err)
                                return
                        }
                        // Pull changes from the remote repository
                        err = worktree.Pull(&git.PullOptions{
                                RemoteName: "origin",
                                Auth:       auth,
                        })
                        if err != nil && err != git.NoErrAlreadyUpToDate {
                                fmt.Printf("Error pulling changes: %v\n", err)
                                return
                        }

                        // Prompt for commit message, username, and email
                        fmt.Print("Give a commit message: ")
                        msg, _ := reader.ReadString('\n')
                        msg = strings.TrimSpace(msg)
                        fmt.Print("Whats your username: ")
                        username, _ := reader.ReadString('\n')
                        username = strings.TrimSpace(username)
                        fmt.Print("Give a commit email: ")
                        email, _ := reader.ReadString('\n')
                        email = strings.TrimSpace(email)

                        // Create a new commit with the added changes
                        commit, err := worktree.Commit(msg, &git.CommitOptions{
                                Author: &object.Signature{
                                        Name:  username,
                                        Email: email,
                                        When:  time.Now(),
                                },
                        })
                        if err != nil {
                                fmt.Printf("Error creating commit: %v\n", err)
                                return
                        }
                        fmt.Println("Commit:", commit)

                        // Push the changes to the remote repository
                        err = repo.Push(&git.PushOptions{
                                RemoteName: "origin",
                                Auth:       auth,
                                RefSpecs:   []config.RefSpec{"refs/heads/master:refs/heads/master"},
                        })
                        if err != nil {
                                fmt.Printf("Error pushing changes: %v\n", err)
                                return
                        }

                        fmt.Println("Changes pushed successfully")
                case "remove":
                        // Add your logic for "remove" here
                case "clean":
                        var trackedFiles []string
                        //trackedFiles = append(settings.FilesToWatch)
                        for _, path := range append(settings.trackedFiles, settings.FilesToWatch...) {
                                fileName := filepath.Base(path)
                                trackedFiles = append(trackedFiles, fileName)
                        }
                        //fmt.Println(trackedFiles)

                        d, err := os.Open(gitDirectory)
                        if err != nil {
                                fmt.Println("Error opening directory:", err)
                                return
                        }
                        defer d.Close()

                        // Read the files in the directory
                        files, err := d.Readdir(-1)
                        if err != nil {
                                fmt.Println("Error reading directory:", err)
                                return
                        }

                        // Loop over the files
                        for _, file := range files {
                                // Check if it's a regular file
                                if file.Mode().IsRegular() {
                                        //fmt.Println(file.Name())
                                        index := indexOf(trackedFiles, file.Name())
                                        // fmt.Println(index, file.Name())
                                        if index == -1 {
                                                filePath := filepath.Join(gitDirectory, file.Name())
                                                err := os.Remove(filePath)
                                                if err != nil {
                                                        // Handle the error if the file cannot be removed
                                                        fmt.Printf("Error deleting file: %v\n", err)
                                                        return
                                                }
                                        }
                                }
                        }
                        // for i := 0; i < len(settings.trackedFiles); i++ {
                        //      // Check if the file exists
                        //      // if _, err := os.Stat(settings.trackedFiles[i]); os.IsNotExist(err) {
                        //      //      continue;
                        //      // }

                        // }
                case "post":

                case "upgrade":
                        // Add your logic for "upgrade" here
                case "add":
                        // Add your logic for "add" here
                        fmt.Print("Enter the path of the file(s) you want to add (comma-separated): ")
                        filesInput, _ := reader.ReadString('\n')
                        filesInput = strings.TrimSpace(filesInput)
                        files := strings.Split(filesInput, ",")

                        if git, err := git.PlainOpen(gitDirectory); err == nil {
                                go func() {
                                        if err := add.AddFilesToGit(files, git); err != nil {
                                                fmt.Println("Error adding files to Git:", err)
                                        } else {
                                                //fmt.Printf("Added %d file(s) to Git\n", len(files))
                                        }
                                }()
                        }

                        for _, file := range files {
                                configFile.trackedFiles = append(configFile.trackedFiles, file)
                                jsonData, err := json.Marshal(configFile)
                                //fmt.Println(jsonData)
                                if err != nil {
                                        fmt.Println("Error with JSON:", err)
                                }

                                err = ioutil.WriteFile("./config.json", []byte(jsonData), 0644)
                                if err != nil {
                                        fmt.Println(err)
                                }
                                add.ModifyFile(file)

                        }

                case "status":
                        // Add your logic for "status" here
                case "watch":
                        // Add logic for adding files
                        fmt.Print("Enter the path of the file(s) you want to add (comma-separated): ")
                        filesInput, _ := reader.ReadString('\n')
                        filesInput = strings.TrimSpace(filesInput)
                        files := strings.Split(filesInput, ",")
                        //fmt.Println("test")
                        if git, err := git.PlainOpen(gitDirectory); err == nil {
                                go func() {
                                        if err := add.AddFilesToGit(files, git); err != nil {
                                                fmt.Println("Error adding files to Git:", err)
                                        } else {
                                                //fmt.Printf("Added %d file(s) to Git\n", len(files))
                                        }
                                }()
                                //fmt.Println("sending")
                                // Start file watchers for added files in separate goroutines
                                for _, file := range files {
                                        // Append the file to the configFile's FilesToWatch slice
                                        configFile.FilesToWatch = append(configFile.FilesToWatch, file)
                                        jsonData, err := json.Marshal(configFile)
                                        //fmt.Println(jsonData)
                                        if err != nil {
                                                fmt.Println("Error with JSON:", err)
                                        }

                                        err = ioutil.WriteFile("./config.json", []byte(jsonData), 0644)
                                        if err != nil {
                                                fmt.Println("Error with file:", err)
                                        }
                                        // Send the watch message to the channel
                                        messages <- "watch " + file

                                        // Handle errors if any
                                        if err != nil {
                                                fmt.Println("Error sending message:", err)
                                        }
                                }
                        } else {
                                fmt.Println("Error opening Git repository:", err)
                        }
                }
        }
}
func indexOf(array []string, target string) int {
        for i, item := range array {
                if item == target {
                        return i // Return the index if found
                }
        }
        return -1 // Return -1 if not found
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
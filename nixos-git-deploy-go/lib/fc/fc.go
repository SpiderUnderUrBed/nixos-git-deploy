package fc

import (
	"strings"
	"io"
	"io/fs"
	"io/ioutil"
	"fmt"
	"os"
	"log"
	"bufio"

	"path/filepath"

	"nixos-git-deploy-go/lib/core"
	
	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v5"
	//"github.com/go-git/go-git/v5/config"
)
var gitDirectory = "/home/spiderunderurbed/.config/nixos-git-deploy/"
var watchedFiles = make(map[string]bool)

func ModifyFile(filename string) error {
// 	var gitDirectory = "/home/spiderunderurbed/.config/nixos-git-deploy/"
// // var watchedFiles = make(map[string]bool)

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



// Function to watch for file changes
func WatchChanges(filename string) {
// 	// var gitDirectory = "/home/spiderunderurbed/.config/nixos-git-deploy/"
// var watchedFiles = make(map[string]bool)

	// Check if the file is already being watched
	fmt.Println(filename)
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
	//fmt.Printf("Watching for changes in file: %s\n", filename)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				//fmt.Printf("File %s has been modified\n", filename)
				err := ModifyFile(filename)
				if err != nil {
					//fmt.Println("ERROR:", err)
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
func CopyFile(src, dest string) error {
	// var gitDirectory = "/home/spiderunderurbed/.config/nixos-git-deploy/"
// var watchedFiles = make(map[string]bool)

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
type FileInfoExtra struct {
	FileInfo   fs.FileInfo
	Dir  string
}

 func PartialMergeTakeTwoLists(arr1 []fs.FileInfo, arr2 []fs.FileInfo){
 	for _, fileInfo := range arr1 {
 		if (fileInfo.Name() != ".git"){
 			_, isMatched := core.ContainsFSName(arr2, fileInfo.Name())
 			if !isMatched {
 				fmt.Println("Added: " + fileInfo.Name())
 			}
 		}
 	}
 	for _, fileInfo := range arr2 {
 		if (fileInfo.Name() != ".git"){
 			_, isMatched := core.ContainsFSName(arr2, fileInfo.Name())
 			if !isMatched {
 				fmt.Println("Removed: " + fileInfo.Name())
 			}
 		}
 	}
 }
func FileEnsureStrings(file string, contentList []string) any {
	for _, e := range contentList {
		content := e
		hasLine := false

		f, err := os.OpenFile(file, os.O_RDWR, 0644)
		if err != nil {
				log.Fatal(err)
		}

		// read the file line by line using scanner
		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
		// do something with a line
		// fmt.Printf("line: %s\n", scanner.Text())
		if (strings.Contains(scanner.Text(), content)){
		hasLine = true
		}
		}

		if err := scanner.Err(); err != nil {
		log.Fatal(err)
		}
		if (hasLine == false){
		if _, err := f.Write([]byte("\n " + content)); err != nil {
		log.Fatal(err)
		}
		if err := f.Close(); err != nil {
		log.Fatal(err)
		}
		}
	}
return "Success"
}
package egit

import (
	"os"
	"fmt"
	"path/filepath"

	"nixos-git-deploy-go/lib/fc"
	"github.com/go-git/go-git/v5"
)
var gitDirectory = "/home/spiderunderurbed/.config/nixos-git-deploy/"
// Function to add files to Git repository
func AddFilesToGit(files []string, r *git.Repository) error {
	// 	var gitDirectory = "/home/spiderunderurbed/.config/nixos-git-deploy/"
	// // var watchedFiles = make(map[string]bool)
	
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
			if err := fc.CopyFile(file, destination); err != nil {
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
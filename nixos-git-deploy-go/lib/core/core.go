package core

import (
	"os"
  //  "fmt"
        //"fs"
  //      "nixos-git-deploy-go"
	"strings"
	//"io/fs"
        //"reflect"
)

//  func (f FileInfoExtra) Name() string {
// 	return func() string {
// 		if f.name != "" {
// 			return f.name
// 		}
// 		return f.FileInfo.Name()
// 	}()
//  }


// var FileInfoExtra 
    
func IntersectArrays(arr1 []string, arr2 []string) []string {
    intersection := make([]string, 0)
    
    // Create a map to store unique elements from arr1
    elements := make(map[string]bool)
    for _, str := range arr1 {
        elements[str] = true
    }
    
    // Check if elements from arr2 exist in the map
    for _, str := range arr2 {
        if elements[str] {
            intersection = append(intersection, str)
        }
    }
    
    return intersection
}

func IndexOf(array []string, target string) int {
        for i, item := range array {
                if item == target {
                        return i // Return the index if found
                }
        }
        return -1 // Return -1 if not found
}

// Function to create directory if it doesn't exist
func EnsureDirectoryExists(directory string) {
        if _, err := os.Stat(directory); os.IsNotExist(err) {
                os.MkdirAll(directory, 0755)
        }
}

// Function to check if directory exists
func IfDirectoryExists(directory string) bool {
        _, err := os.Stat(directory)
        return !os.IsNotExist(err)
}

func Unique(arr []string) []string {
    occurred := map[string]bool{}
    result := []string{}
    //var lastIndex int

    for _, e := range arr {
        // check if already the mapped
        // variable is set to true or not
        if !occurred[e] {
            occurred[e] = true
            // Append to result slice.
            result = append(result, e)
           // lastIndex = i
        }
    }

    return result
}
func UniqueWithEncryption(slice []string) []string {
	unique := make(map[string]bool)
	windowStart := 0

	for windowEnd := 0; windowEnd < len(slice); windowEnd++ {
		trimmed := strings.TrimSuffix(slice[windowEnd], ".encrypted")
		if !unique[trimmed] {
			unique[trimmed] = true
			// If the element is unique, move the window start
			// and update the value at the window start index
			if windowStart != windowEnd {
				slice[windowStart] = slice[windowEnd]
			}
			windowStart++
		}
	}

	// Resize the slice to remove duplicates
	return slice[:windowStart]
}

func FileExists(fileName string) bool {
    _, err := os.Stat(fileName)
    return !os.IsNotExist(err)
}
//type FileInfo interface {
//        fs.FileInfo | FileInfoExtra
//}
//func ContainsFSName[FS fs.FileInfo | FileInfoExtra](slice []FS, item string) (any, bool) {
//func ContainsFSName[FS FileInfo](slice []FS, item string) (any, bool) { 
    type FileInfo interface {
        Name() string
        Dir() *string
    }
    
    type FileInfoExtra[FS FileInfo] struct {
        FileInfo FS
        Dir      string
    }
    
    func (f FileInfoExtra[FS]) Name() string {
        return f.FileInfo.Name()
    }
    
    func ContainsFSName[FS FileInfo](slice []FS, item string) (FileInfoExtra[FS], bool) {
        for _, s := range slice {
            if s.Name() == item {
                return FileInfoExtra[FS]{FileInfo: s, Dir: *s.Dir()}, true
            }
        }
        return FileInfoExtra[FS]{}, false
    }
    

        // // If the file is not found in the slice, create FileInfo from os.Stat
        // file, err := os.Stat(item)
        // if err != nil {
        //     // If there was an error, return false indicating file not found
        //     return FileInfoExtra[FS]{}, false
        // }
        // // Check if os.FileInfo satisfies the FileInfo interface
        // var fileInfo FileInfo = file
        // // Perform type assertion to ensure compatibility with FS
        // fileInfoFS, ok := fileInfo.(FS)
        // if !ok {
        //     // If fileInfo doesn't satisfy FS, return false
        //     return FileInfoExtra[FS]{}, false
        // }
        // return FileInfoExtra[FS]{FileInfo: fileInfoFS}, true
            // Check if FS type has Dir() method
            // fmt.Println(s)
            // if dirFunc, ok := interface{}(s).(interface{ Dir() string }); ok {
            //     fmt.Print("Dir")
            //     return FileInfoExtra[FS]{FileInfo: s, Dir: dirFunc.Dir()}, true
            // }
            // If FS type doesn't have Dir() method, return FileInfo without Dir
func Contains(array []string, target string) bool {
        for _, item := range array {
                if item == target {
                        return true
                }
        }
        return false
}
func Remove(slice []string, s int) []string {
    return append(slice[:s], slice[s+1:]...)
}
// func PartialMergeTakeTwoLists(arr1 []fs.FileInfo, arr2 []fs.FileInfo){

// }
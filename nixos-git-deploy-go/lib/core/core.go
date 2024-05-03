package core

import (
	"os"
        "fmt"
	"strings"
	"io/fs"
)

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
func ContainsFSName(slice []fs.FileInfo, item string) bool {
	for _, s := range slice {
		//fmt.Println(s.Name() + " " + item)
		if s.Name() == item {
			return true
		}
	}
	return false
}
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
func PartialMergeTakeTwoLists(arr1 []fs.FileInfo, arr2 []fs.FileInfo){
	for _, fileInfo := range arr1 {
		if (fileInfo.Name() != ".git"){
			if !ContainsFSName(arr2, fileInfo.Name()) {
				fmt.Println("Added: " + fileInfo.Name())
			}
		}
	}
	for _, fileInfo := range arr2 {
		if (fileInfo.Name() != ".git"){
			if !ContainsFSName(arr1, fileInfo.Name()) {
				fmt.Println("Removed: " + fileInfo.Name())
			}
		}
	}
}
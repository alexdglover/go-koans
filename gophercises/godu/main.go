package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
)

func main() {
	// parse CLI args
	humanReadable := flag.Bool("h", false, "show sizes in human readable format")

	flag.Parse()
	var path string
	path = flag.Arg(0)

	fmt.Printf("humanReadable is %t, path is %s\n", *humanReadable, path)

	getDirectoryDiskUsageInfo(path)
	//fmt.Println(result)

}

type directorySizeInfo struct {
	totalSize int64
	files     []fs.DirEntry
}

func getDirectoryDiskUsageInfo(path string) directorySizeInfo {
	summary := directorySizeInfo{
		totalSize: 0,
		files:     make([]fs.DirEntry, 0),
	}

	pathObject, err := os.Open(path)
	if err != nil {
		fmt.Println("error opening path")
	}
	pathFile, err := pathObject.Stat()
	if err != nil {
		fmt.Println("error getting path file info")
	}
	// handle single file use case
	if pathFile.IsDir() == false {
		fmt.Printf("%d %s\n", pathFile.Size(), path)
		summary.totalSize = pathFile.Size()
		return summary
	}

	if string(path[len(path)-1:]) != "/" {
		path = fmt.Sprintf("%s/", path)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Unable to list directory at path %s", path)
	}

	for _, file := range files {
		// get size of files
		info, err := file.Info()
		if err != nil {
			fmt.Println("error getting file info")
		}

		if info.IsDir() != true {
			// if it's not a directory, assume it's a file
			summary.totalSize += info.Size()
			fmt.Printf("%d %s%s\n", info.Size(), path, info.Name())
		} else if info.IsDir() == true {
			fmt.Printf("%s is a dir! recursing...\n", info.Name())
			result := getDirectoryDiskUsageInfo(fmt.Sprintf("%s%s", path, info.Name()))
			//// a directory, regardless of its contents, is always 4096 bytes or 4KB
			//summary.totalSize += 4096
			summary.totalSize += result.totalSize
		}
	}

	// a directory, regardless of its contents, is always 4096 bytes or 4KB
	summary.totalSize += 4096
	fmt.Printf("%d %s\n", summary.totalSize, path)

	return summary

	/*
		example du -h output
		24K	 ./gophercises/godu/.idea
		2.0K ./gophercises/godu
		28K	 ./gophercises/quiz-game/.idea
		2.2M ./gophercises/quiz-game
		4.2M ./gophercises
		7.1M .
	*/
}

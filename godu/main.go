package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"time"
)

func main() {
	// parse CLI args
	humanReadable := flag.Bool("h", false, "show sizes in human readable format")
	summarize := flag.Bool("s", false, "display only a total for each argument")

	flag.Parse()
	var path string
	path = flag.Arg(0)

	g := godu{
		humanReadable: *humanReadable,
	}

	resultsChannel := make(chan directorySizeInfo)

	go g.getDirectoryDiskUsageInfo(path, resultsChannel)

	select {
	case result := <-resultsChannel:
		g.reportStatistics(result.path, result.totalSize)
		if *summarize == false {
			for _, file := range result.files {
				g.reportStatistics(file.Name(), file.Size())
			}
		}
	}
}

type godu struct {
	humanReadable bool
}

type directorySizeInfo struct {
	totalSize int64
	path      string
	files     []fs.FileInfo
}

// A struct with the same fields as os.FileInfo interface
type mutableFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (f mutableFileInfo) Name() string {
	return f.name
}

func (f mutableFileInfo) Size() int64 {
	return f.size
}

func (f mutableFileInfo) Mode() os.FileMode {
	return 1
}

func (f mutableFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f mutableFileInfo) IsDir() bool {
	return f.isDir
}

func (f mutableFileInfo) Sys() any {
	return nil
}

func (g godu) getDirectoryDiskUsageInfo(path string, results chan directorySizeInfo) {
	dsi := directorySizeInfo{
		totalSize: 0,
		path:      path,
		files:     make([]fs.FileInfo, 0),
	}

	pathObject, err := os.Open(path)
	if err != nil {
		fmt.Println("error opening path")
	}
	fileInfo, err := pathObject.Stat()
	if err != nil {
		fmt.Println("error getting path file info")
	}
	// handle single file use case
	if fileInfo.IsDir() == false {
		dsi.totalSize = fileInfo.Size()
		dsi.files = append(dsi.files, fileInfo)
		results <- dsi
		return
	}

	// the directory itself takes up a minimum of 4096 bytes, so we need to account for the
	// directory in the total size
	dsi.totalSize += fileInfo.Size()

	if string(path[len(path)-1:]) != "/" {
		path += "/"
	}

	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Unable to list directory at path %s", path)
	}

	// if we're not handling a single file, likely going to recurse through multiple directories.
	// make a new channel to pass to child goroutines. also create a variable to track the number
	// of msgs expected, since our un-coordinated goroutines can't close the channel
	recursiveResultsChannel := make(chan directorySizeInfo)
	msgsExpected := 0

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			fmt.Println("error getting file info")
		}

		if info.IsDir() != true {
			// if it's not a directory, assume it's a file
			fileRecord := mutableFileInfo{
				name:  path + info.Name(),
				size:  info.Size(),
				isDir: false,
			}
			dsi.totalSize += info.Size()
			dsi.files = append(dsi.files, fileRecord)
		} else if info.IsDir() == true {
			go g.getDirectoryDiskUsageInfo(path+info.Name(), recursiveResultsChannel)
			msgsExpected += 1
		}
	}

	for msgsExpected > 0 {
		result := <-recursiveResultsChannel
		dsi.totalSize += result.totalSize
		for _, file := range result.files {
			dsi.files = append(dsi.files, file)
		}
		directoryRecord := mutableFileInfo{
			name:  result.path,
			size:  result.totalSize,
			isDir: true,
		}
		dsi.files = append(dsi.files, directoryRecord)
		msgsExpected -= 1
	}

	results <- dsi

}

func (g godu) reportStatistics(path string, size int64) {
	var sizeAsString string
	if g.humanReadable {
		sizeAsString = ByteCountIEC(size)
	} else {
		sizeAsString = strconv.FormatInt(size, 10)
	}
	fmt.Printf("%s\t\t%s\n", sizeAsString, path)
}
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

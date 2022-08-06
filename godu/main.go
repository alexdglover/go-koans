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

	//fmt.Printf("humanReadable is %t, path is %s\n", *humanReadable, path)

	g := godu{
		humanReadable: *humanReadable,
	}
	g.getDirectoryDiskUsageInfo(path)
	//fmt.Println(result)

}

type godu struct {
	humanReadable bool
}

type directorySizeInfo struct {
	totalSize int64
	files     []fs.DirEntry
}

func (g godu) getDirectoryDiskUsageInfo(path string) directorySizeInfo {
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
		//fmt.Printf("%d %s\n", pathFile.Size(), path)
		g.reportStatistics(path, pathFile.Size())
		summary.totalSize = pathFile.Size()
		return summary
	}

	if string(path[len(path)-1:]) != "/" {
		path += "/"
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
			g.reportStatistics(path+info.Name(), info.Size())
		} else if info.IsDir() == true {
			result := g.getDirectoryDiskUsageInfo(path + info.Name())
			summary.totalSize += result.totalSize
		}
	}

	// a directory, regardless of its contents, is always 4096 bytes or 4KB
	summary.totalSize += 4096
	g.reportStatistics(path, summary.totalSize)

	return summary
}

func (g godu) reportStatistics(path string, size int64) {
	var sizeAsString string
	if g.humanReadable {
		sizeAsString = ByteCountIEC(size)
	} else {
		sizeAsString = string(size)
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

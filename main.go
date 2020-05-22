package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"watcher/cmd"
)

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	os.Chdir(dir)
	runtime.GOMAXPROCS(1)
	cmd.Execute()
}

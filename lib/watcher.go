package lib

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Args struct {
	Cmd   string
	Match string
	Args  []string
	Dirs  []string
}

var process *os.Process
var timeID *time.Timer

var lock sync.Mutex

func WatchDir(watch *fsnotify.Watcher, dir string) error {

	dirpath, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	//通过Walk来遍历目录下的所有子目录
	return filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		//判断是否为目录，监控目录,目录下文件也在监控范围内，不需要加
		if info.IsDir() {
			path, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			err = watch.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

//启动进程

func Watch(args *Args) {

	timeID = nil
	//创建一个监控对象
	go startProcess(*args)

	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watch.Close()
	//添加要监控的文件

	for _, v := range args.Dirs {
		err = WatchDir(watch, v)
		if err != nil {
			log.Fatal(err)
		}
	}

	//我们另启一个goroutine来处理监控对象的事件
	go func() {

		for {
			select {
			case ev := <-watch.Events:
				{
					if ev.Op&fsnotify.Create == fsnotify.Create {
						file, err := os.Stat(ev.Name)
						if err == nil && file.IsDir() {
							watch.Add(ev.Name)
						}
					}

					if ev.Op&fsnotify.Remove == fsnotify.Remove {
						//如果删除文件是目录，则移除监控
						fi, err := os.Stat(ev.Name)
						if err == nil && fi.IsDir() {
							watch.Remove(ev.Name)
						}
					}

					//我们只需关心文件的修改
					if ev.Op&fsnotify.Write == fsnotify.Write {
						if args.Match != "" {
							args.Match = strings.ToLower(args.Match)
							var bytes = []byte(strings.ToLower(path.Ext(ev.Name)))
							m, _ := regexp.Match("^("+args.Match+")$", bytes[1:])
							if !m {
								break
							}
						}

						log.Println(ev.Name, "modifyed")
						restartProcess(args)
					}
				}
			case err := <-watch.Errors:
				{
					fmt.Println("error : ", err)
					return
				}
			}
		}
	}()

	//循环
	select {}
}

func restartProcess(args *Args) {
	if timeID != nil {
		timeID.Stop()
		timeID.Reset(3 * time.Second)
		return
	}

	timeID = time.AfterFunc(3*time.Second, func() {
		if process != nil {
			fmt.Printf("pid:%d close\n", process.Pid)
			process.Kill()
		}

		go startProcess(*args)
	})
}

func startProcess(a Args) error {
	lock.Lock()
	defer lock.Unlock()
	log.Printf("start:%s %s\n", a.Cmd, strings.Join(a.Args, " "))

	cmd := exec.Command(a.Cmd, a.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	err := cmd.Start()
	if err != nil {
		log.Printf("process error:%s\n", err.Error())
		return err
	}

	process = cmd.Process
	err = cmd.Wait()
	log.Printf("Wait error:%s\n", err.Error())
	return err
}

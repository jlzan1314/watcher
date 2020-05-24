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
	"syscall"
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

var cmd *exec.Cmd
var pid int

var ModifiedFiles *sync.Map

func WatchDir(watch *fsnotify.Watcher, dir string) error {

	dirpath, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	//通过Walk来遍历目录下的所有子目录
	return filepath.Walk(dirpath, func(path2 string, info os.FileInfo, err error) error {
		//判断是否为目录，监控目录,目录下文件也在监控范围内，不需要加
		if info.IsDir() {
			absPath, err2 := filepath.Abs(path2)

			if err2 != nil {
				return err2
			}
			err2 = watch.Add(absPath)
			if err2 != nil {
				return err2
			}
		}
		return nil
	})
}

//启动进程

func Watch(args *Args) {

	ModifiedFiles = &sync.Map{}

	timeID = nil
	//创建一个监控对象
	go StartProcess(*args)

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

						ModifiedFiles.Store(ev.Name, 1)
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

		ModifiedFiles.Range(func(key interface{}, value interface{}) bool {
			if k, ok := key.(string); ok {
				log.Println(k + " modifyed")
			}
			ModifiedFiles.Delete(key)
			return true
		})

		if pid != 0 {
			fmt.Printf("pid:%d close\n", pid)
			syscall.Kill(pid, syscall.SIGTERM)
			for {
				if err := syscall.Kill(pid, 0); err != nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			pid = 0
		}

		go StartProcess(*args)
	})
}

func initCmd(a Args) *exec.Cmd {
	log.Printf("start:%s %s\n", a.Cmd, strings.Join(a.Args, " "))
	cmd := exec.Command(a.Cmd, a.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	// user, err := user.Lookup("root")
	// if err == nil {
	// 	log.Printf("uid=%s,gid=%s", user.Uid, user.Gid)
	// 	uid, _ := strconv.Atoi(user.Uid)
	// 	gid, _ := strconv.Atoi(user.Gid)
	// 	cmd.SysProcAttr = &syscall.SysProcAttr{}
	// 	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	// }

	return cmd
}

func StartProcess(a Args) error {

	cmd := initCmd(a)

	if err := cmd.Start(); err != nil {
		log.Printf("process error:%s\n", err.Error())
		return err
	}

	log.Printf("start process id:%d\n", cmd.Process.Pid)

	pid = cmd.Process.Pid

	if err := cmd.Wait(); err != nil {
		log.Printf("process wait error:%s\n", err.Error())
		return err
	}

	log.Printf("stoped process id:%d, res:%s\n", cmd.Process.Pid, cmd.ProcessState.String())

	return nil
}

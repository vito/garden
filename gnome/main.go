package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"./barrier"
	"./daemon"
)

var runPath = flag.String(
	"run",
	"./run",
	"where to put the gnome daemon .sock file",
)

var rootPath = flag.String(
	"root",
	"./root",
	"root filesystem for the container",
)

var libPath = flag.String(
	"lib",
	"./lib",
	"directory containing hooks",
)

var title = flag.String(
	"title",
	"garden gnome",
	"title for the container gnome daemon",
)

var continueAsChild = flag.Bool(
	"continue",
	false,
	"(internal) continue execution as containerized daemon",
)

func createContainerizedProcess() (int, error) {
	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()

	pid, _, err := syscall.RawSyscall(
		syscall.SYS_CLONE,
		CLONE_NEWNS|CLONE_NEWUTS|CLONE_NEWIPC|CLONE_NEWPID|CLONE_NEWNET,
		0,
		0,
	)

	if err != 0 {
		return 0, err
	}

	return int(pid), nil
}

func childContinue(runPath string) {
	fmt.Println("pid:", os.Getpid())
	sid, err := syscall.Setsid()
	fmt.Println("pid, sid, err:", os.Getpid(), sid, err)

	if err != nil {
		panic(err)
	}

	childBarrier := &barrier.Barrier{path.Join(runPath, "child-barrier")}

	daemon := daemon.New(path.Join(runPath, "wshd.sock"))

	err = childBarrier.Signal()

	fmt.Println("CHILD -> PARENT", err)

	if err != nil {
		panic(err)
	}

	os.Stdin.Close()
	os.Stdout.Close()
	os.Stderr.Close()

	daemon.Start()

	select {}
}

func main() {
	flag.Parse()

	fullRunPath, err := filepath.Abs(*runPath)
	if err != nil {
		log.Fatalln(err)
	}

	fullRunPath, err = filepath.EvalSymlinks(fullRunPath)
	if err != nil {
		log.Fatalln(err)
	}

	if *continueAsChild {
		childContinue(fullRunPath)
		return
	}

	fullLibPath, err := filepath.Abs(*libPath)
	if err != nil {
		log.Fatalln(err)
	}

	fullLibPath, err = filepath.EvalSymlinks(fullLibPath)
	if err != nil {
		log.Fatalln(err)
	}

	fullRootPath, err := filepath.Abs(*rootPath)
	if err != nil {
		log.Fatalln(err)
	}

	fullRootPath, err = filepath.EvalSymlinks(fullRootPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = syscall.Unshare(CLONE_NEWNS)
	if err != nil {
		log.Fatalln(err)
	}

	err = exec.Command(path.Join(*libPath, "hook-parent-before-clone.sh")).Run()
	if err != nil {
		log.Fatalln(err)
	}

	parentBarrier, err := barrier.New(path.Join(fullRunPath, "parent-barrier"))
	if err != nil {
		log.Fatalln(err)
	}

	childBarrier, err := barrier.New(path.Join(fullRunPath, "child-barrier"))
	if err != nil {
		log.Fatalln(err)
	}

	pid, err := createContainerizedProcess()

	fmt.Println("created", pid, err)

	if pid == 0 {
		err := parentBarrier.Wait()

		fmt.Println("GOT PARENT -> CHILD", err)

		err = exec.Command(path.Join(fullLibPath, "hook-child-before-pivot.sh")).Run()

		fmt.Println("hook-child-before-pivot", err)

		if err != nil {
			log.Fatalln(err)
		}

		err = os.Chdir(fullRootPath)

		fmt.Println("chdir", err)

		if err != nil {
			log.Fatalln(err)
		}

		err = os.MkdirAll("mnt", 0700)

		fmt.Println("mkdir-all", err)

		if err != nil {
			log.Fatalln(err)
		}

		err = syscall.PivotRoot(".", "mnt")

		fmt.Println("PIVOOT", err)

		if err != nil {
			log.Fatalln(err)
		}

		err = os.Chdir("/")
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("chdir /", err)

		err = exec.Command(path.Join("/mnt", fullLibPath, "hook-child-after-pivot.sh")).Run()

		fmt.Println("hook-child-after-pivot", err)

		if err != nil {
			log.Fatalln(err)
		}

		err = syscall.Exec("/sbin/wshd", []string{"/sbin/wshd", "--continue", "--run", path.Join("/mnt", fullRunPath)}, []string{})

		if err != nil {
			log.Fatalln(err)
		}
	}

	os.Setenv("PID", fmt.Sprintf("%d", pid))

	err = exec.Command(path.Join(*libPath, "hook-parent-after-clone.sh")).Run()
	if err != nil {
		log.Fatalln(err)
	}

	err = parentBarrier.Signal()
	fmt.Println("PARENT -> CHILD", err)

	err = childBarrier.Wait()
	fmt.Println("GOT CHILD -> PARENT", err)

	println("OK DONE")

	os.Exit(0)
}

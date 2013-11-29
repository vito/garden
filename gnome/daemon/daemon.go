package daemon

import (
	"bytes"
	"encoding/gob"
	"net"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"log/syslog"
	"log"

	"../protocol"
)

type Daemon struct {
	socketPath string
	logger *syslog.Writer
}

func New(socketPath string) *Daemon {
	logger, err := syslog.Dial("unixgram", "/mnt/dev/log", syslog.LOG_KERN, "garden")
	if err != nil {
		panic(err)
	}

	log.SetOutput(logger)

	return &Daemon{
		socketPath: socketPath,
	}
}

func (d *Daemon) Start() error {
	listener, err := net.Listen("unix", d.socketPath)
	if err != nil {
		return err
	}

	go d.handleConnections(listener)
	//go d.handleChildExits()

	return nil
}

func (d *Daemon) handleConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("daemon error accepting connection:", err)
			continue
		}

		go d.serveConnection(conn.(*net.UnixConn))
	}
}

func (d *Daemon) serveConnection(conn *net.UnixConn) {
	//defer conn.Close()

	decoder := gob.NewDecoder(conn)

	var requestMessage protocol.RequestMessage

	err := decoder.Decode(&requestMessage)
	if err != nil {
		log.Println("failed reading request:", err)
		return
	}

	fmt.Println("request:", requestMessage)

	response := protocol.ResponseMessage{}

	res := new(bytes.Buffer)

	encoder := gob.NewEncoder(res)

	err = encoder.Encode(response)
	if err != nil {
		log.Println("failed writing response:", err)
		return
	}

	stdinOut, stdinIn, err := os.Pipe()
	if err != nil {
		log.Println("failed making stdin pipe", err)
		return
	}

	stdoutOut, stdoutIn, err := os.Pipe()
	if err != nil {
		log.Println("failed making stdout pipe", err)
		return
	}

	stderrOut, stderrIn, err := os.Pipe()
	if err != nil {
		log.Println("failed making stderr pipe", err)
		return
	}

	statusOut, statusIn, err := os.Pipe()
	if err != nil {
		log.Println("failed making status pipe", err)
		return
	}

	os.Setenv("PATH", "/sbin:/bin:/usr/sbin:/usr/bin")

	cmd := exec.Command(requestMessage.Argv[0], requestMessage.Argv[1:]...)

	cmd.Stdin = stdinOut
	cmd.Stdout = stdoutIn
	cmd.Stderr = stderrIn

	rights := syscall.UnixRights(
		int(stdinIn.Fd()),
		int(stdoutOut.Fd()),
		int(stderrOut.Fd()),
		int(statusOut.Fd()),
	)

	_, _, err = conn.WriteMsgUnix(res.Bytes(), rights, nil)
	if err != nil {
		log.Println("failed sending unix message:", err)
		return
	}

	err = cmd.Run()

	log.Println("command exited:", err)

	exitStatus := 255

	if cmd.ProcessState != nil {
		exitStatus = int(cmd.ProcessState.Sys().(syscall.WaitStatus) % 255)
	}

	log.Println("exit status:", exitStatus)

	fmt.Fprintf(statusIn, "%d\n", exitStatus)
}

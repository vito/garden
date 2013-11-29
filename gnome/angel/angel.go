package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"syscall"

	"../protocol"
)

var socketPath = flag.String(
	"socket",
	"run/wshd.sock",
	"path to gnome socket file",
)

var user = flag.String(
	"user",
	"root",
	"user to run the command as",
)

func main() {
	flag.Parse()

	conn, err := net.Dial("unix", *socketPath)
	if err != nil {
		log.Fatalln(err)
	}

	request := protocol.RequestMessage{
		TTY:  false,
		Argv: flag.Args(),
		User: *user,
	}

	encoder := gob.NewEncoder(conn)

	err = encoder.Encode(request)
	if err != nil {
		log.Fatalln("failed writing request:", err)
	}

	var b [2048]byte
	var oob [2048]byte

	n, oobn, _, _, err := conn.(*net.UnixConn).ReadMsgUnix(b[:], oob[:])
	if err != nil {
		log.Fatalln("failed to read unix msg:", err, n, oobn)
	}

	scms, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		log.Fatalln("failed to parse socket control message:", err)
	}

	if len(scms) < 1 {
		log.Fatalln("no socket control messages sent")
	}

	scm := scms[0]

	fds, err := syscall.ParseUnixRights(&scm)
	if err != nil {
		log.Fatalln("failed to parse unix rights", err)
		return
	}

	if len(fds) != 4 {
		log.Fatalln("invalid number of fds; need 4, got", len(fds))
	}

	stdin := os.NewFile(uintptr(fds[0]), "stdin")
	stdout := os.NewFile(uintptr(fds[1]), "stdout")
	stderr := os.NewFile(uintptr(fds[2]), "stderr")
	status := os.NewFile(uintptr(fds[3]), "status")

	err = syscall.SetNonblock(int(os.Stdin.Fd()), false)
	if err != nil {
		log.Fatalln("failed setting fd nonblock:", err)
	}

	err = syscall.SetNonblock(int(os.Stdout.Fd()), false)
	if err != nil {
		log.Fatalln("failed setting fd nonblock:", err)
	}

	err = syscall.SetNonblock(int(os.Stderr.Fd()), false)
	if err != nil {
		log.Fatalln("failed setting fd nonblock:", err)
	}

	for _, fd := range fds {
		err := syscall.SetNonblock(fd, false)
		if err != nil {
			log.Fatalln("failed setting fd nonblock:", err, fd)
		}
	}

	done := make(chan bool)

	go func() {
		io.Copy(stdin, os.Stdin)
		os.Stdin.Close()
		stdin.Close()
		done <- true
	}()

	go func() {
		io.Copy(os.Stdout, stdout)
		stdout.Close()
		os.Stdout.Close()
		done <- true
	}()

	go func() {
		io.Copy(os.Stderr, stderr)
		stderr.Close()
		os.Stderr.Close()
		done <- true
	}()

	<-done
	<-done
	<-done

	var exitStatus int

	_, err = fmt.Fscanf(status, "%d\n", &exitStatus)
	if err != nil {
		log.Fatalln("error reading status:", err)
	}

	os.Exit(exitStatus)
}

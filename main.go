package main

import (
	"flag"
	"log"
	"path"

	"github.com/vito/garden/backend"
	"github.com/vito/garden/backend/fakebackend"
	"github.com/vito/garden/backend/linuxbackend"
	"github.com/vito/garden/backend/linuxbackend/resource_pool"
	"github.com/vito/garden/server"
)

var socketFilePath = flag.String(
	"socket",
	"/tmp/warden.sock",
	"where to put the wardern server .sock file",
)

var backendName = flag.String(
	"backend",
	"linux",
	"which backend to use (linux or fake)",
)

var rootPath = flag.String(
	"root",
	"",
	"directory containing backend-specific scripts (i.e. ./linux/create.sh)",
)

var depotPath = flag.String(
	"depot",
	"",
	"directory in which to store containers",
)

var rootFSPath = flag.String(
	"rootfs",
	"",
	"directory of the rootfs for the containers",
)

func main() {
	flag.Parse()

	var backend backend.Backend

	switch *backendName {
	case "linux":
		if *rootPath == "" {
			log.Fatalln("must specify -root with linux backend")
		}

		if *depotPath == "" {
			log.Fatalln("must specify -depot with linux backend")
		}

		if *rootFSPath == "" {
			log.Fatalln("must specify -rootfs with linux backend")
		}

		pool := resource_pool.New(*rootFSPath)

		backend = linuxbackend.New(path.Join(*rootPath, "linux"), *depotPath, *rootFSPath, pool)
	case "fake":
		backend = fakebackend.New()
	}

	err := backend.Setup()
	if err != nil {
		log.Fatalln("failed to setup backend:", err)
	}

	wardenServer := server.New(*socketFilePath, backend)

	err = wardenServer.Start()
	if err != nil {
		log.Fatalln("failed to start:", err)
	}

	select {}
}

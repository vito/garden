package main

import (
	"flag"
	"log"
	"net"
	"path"
	"strings"
	"time"
	"encoding/json"
	"os"

	"github.com/hashicorp/serf/command/agent"

	"github.com/vito/garden/backend"
	"github.com/vito/garden/backend/fake_backend"
	"github.com/vito/garden/backend/linux_backend"
	"github.com/vito/garden/backend/linux_backend/linux_container_pool"
	"github.com/vito/garden/backend/linux_backend/network_pool"
	"github.com/vito/garden/backend/linux_backend/port_pool"
	"github.com/vito/garden/backend/linux_backend/quota_manager"
	"github.com/vito/garden/backend/linux_backend/uid_pool"
	"github.com/vito/garden/command_runner"
	"github.com/vito/garden/command_runner/remote_command_runner"
	"github.com/vito/garden/server"
)

type WardenMember struct {
	Addr            string
	AvailableMemory uint64
	AvailableDisk   uint64
}

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

var remoteHost = flag.String(
	"remoteHost",
	"",
	"machine to use for the Linux backend",
)

var remotePort = flag.Int(
	"remotePort",
	22,
	"SSH port of the remote machine",
)

var disableQuotas = flag.Bool(
	"disableQuotas",
	false,
	"disable disk quotas",
)

var debug = flag.Bool(
	"debug",
	false,
	"show low-level command output",
)

var serfAgentRPCAddr = flag.String(
	"serfAgentRPCAddr",
	"127.0.0.1:7373",
	"local serf agent's RPC address",
)

var serfMembers = flag.String(
	"serfMembers",
	"",
	"join a serf cluster",
)

func main() {
	flag.Parse()

	var serfClient *agent.RPCClient
	var err error

	if *serfMembers != "" {
		interval := 1 * time.Second

		for {
			serfClient, err = agent.NewRPCClient(*serfAgentRPCAddr)
			if err == nil {
				break
			}

			log.Println(
				"failed to reach serf agent at",
				*serfAgentRPCAddr,
				"trying again in",
				interval,
			)

			time.Sleep(interval)
		}
	}

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

		uidPool := uid_pool.New(10000, 256)

		_, ipNet, err := net.ParseCIDR("10.254.0.0/22")
		if err != nil {
			panic(err)
		}

		networkPool := network_pool.New(ipNet)

		// TODO: base on ephemeral port range
		portPool := port_pool.New(61000, 6501)

		var runner command_runner.CommandRunner

		runner = command_runner.New(*debug)

		if *remoteHost != "" {
			runner = remote_command_runner.New(
				"root",
				*remoteHost,
				uint32(*remotePort),
				"/host",
				runner,
			)
		}

		quotaManager, err := quota_manager.New(*depotPath, *rootPath, runner)
		if err != nil {
			panic(err)
		}

		if *disableQuotas {
			quotaManager.Disable()
		}

		pool := linux_container_pool.New(
			path.Join(*rootPath, "linux"),
			*depotPath,
			*rootFSPath,
			uidPool,
			networkPool,
			portPool,
			runner,
			quotaManager,
		)

		backend = linux_backend.New(pool)
	case "fake":
		backend = fake_backend.New()
	}

	log.Println("setting up backend")

	err = backend.Setup()
	if err != nil {
		log.Fatalln("failed to set up backend:", err)
	}

	log.Println("starting server; listening on", *socketFilePath)

	wardenServer := server.New(*socketFilePath, backend)

	err = wardenServer.Start()
	if err != nil {
		log.Fatalln("failed to start:", err)
	}

	hostname, _ := os.Hostname()

	if serfClient != nil {
		addrs := strings.Split(*serfMembers, ",")
		joinedCount, err := serfClient.Join(addrs, false)
		if err != nil {
			log.Fatalln("failed to join serf cluster:", err)
		}

		log.Println("joined", joinedCount, "members of serf cluster")

		for {
			containers, err := backend.Containers()
			if err != nil {
				log.Println("failed to get containers:", err)
				continue
			}

			member := WardenMember{
				Addr: hostname,
				AvailableMemory: uint64(128 * len(containers)),
				AvailableDisk:   uint64(1024 * len(containers)),
			}

			json, err := json.Marshal(member)
			if err != nil {
				log.Println("error marshaling member:", err)
				continue
			}

			log.Println("emitting warden.capacity:", member)

			err = serfClient.UserEvent("warden.capacity", json, false)
			if err != nil {
				log.Println("failed emitting warden.capacity:", err)
			}

			time.Sleep(10 * time.Second)
		}
	}

	select {}
}

package resource_pool

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"strconv"
	"sync"
	"time"
)

type ResourcePool struct {
	rootPath  string
	depotPath string

	defaultRootFSPath string

	nextContainer int64

	sync.RWMutex
}

func New(rootPath, depotPath, defaultRootFSPath string) *ResourcePool {
	return &ResourcePool{
		rootPath:  rootPath,
		depotPath: depotPath,

		defaultRootFSPath: defaultRootFSPath,

		nextContainer: time.Now().UnixNano(),
	}
}

func (p *ResourcePool) Setup() error {
	setup := exec.Command(path.Join(p.rootPath, "setup.sh"))

	setup.Env = []string{
		"POOL_NETWORK=10.254.0.0/24",
		"ALLOW_NETWORKS=",
		"DENY_NETWORKS=",
		"CONTAINER_ROOTFS_PATH=" + p.defaultRootFSPath,
		"CONTAINER_DEPOT_PATH=" + p.depotPath,
		"CONTAINER_DEPOT_MOUNT_POINT_PATH=/",
		"DISK_QUOTA_ENABLED=true",

		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	}

	out, err := setup.CombinedOutput()
	if err != nil {
		fmt.Println("error setting up:", string(out))
		return err
	}

	return nil
}

func (p *ResourcePool) Acquire() (*Resources, error) {
	p.Lock()

	resources := &Resources{
		containerID: p.generateContainerID(),

		depotPath:  p.depotPath,
		rootFSPath: p.defaultRootFSPath,
	}

	defer p.Unlock()

	create := exec.Command(
		path.Join(p.rootPath, "create.sh"),
		resources.ContainerPath(),
	)

	create.Env = resources.Env()

	out, err := create.CombinedOutput()
	if err != nil {
		log.Println("error creating container:", string(out), create.ProcessState)
		return nil, err
	}

	return resources, nil
}

func (p *ResourcePool) Release(resources *Resources) error {
	destroy := exec.Command(
		path.Join(p.rootPath, "destroy.sh"),
		resources.ContainerPath(),
	)

	destroy.Env = resources.Env()

	out, err := destroy.CombinedOutput()
	if err != nil {
		log.Println("error destroying container:", string(out), destroy.ProcessState)
		return err
	}

	return nil
}

func (p *ResourcePool) generateContainerID() string {
	p.nextContainer++

	containerID := []byte{}

	var i uint
	for i = 0; i < 11; i++ {
		containerID = strconv.AppendInt(
			containerID,
			(p.nextContainer>>(55-(i+1)*5))&31,
			32,
		)
	}

	return string(containerID)
}

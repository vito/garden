package linuxbackend

import (
	"fmt"
	"log"
	"os/exec"
	"path"

	"github.com/vito/garden/backend"
	"github.com/vito/garden/backend/linuxbackend/resource_pool"
)

type LinuxBackend struct {
	rootPath     string
	depotPath    string
	rootFSPath   string
	resourcePool *resource_pool.ResourcePool

	containers map[string]*LinuxContainer
}

type UnknownHandleError struct {
	Handle string
}

func (e UnknownHandleError) Error() string {
	return "unknown handle: " + e.Handle
}

func New(rootPath, depotPath, rootFSPath string, resourcePool *resource_pool.ResourcePool) *LinuxBackend {
	return &LinuxBackend{
		rootPath:     rootPath,
		depotPath:    depotPath,
		rootFSPath:   rootFSPath,
		resourcePool: resourcePool,

		containers: make(map[string]*LinuxContainer),
	}
}

func (b *LinuxBackend) Setup() error {
	setup := exec.Command(path.Join(b.rootPath, "setup.sh"))

	setup.Env = []string{
		"POOL_NETWORK=10.254.0.0/24",
		"ALLOW_NETWORKS=",
		"DENY_NETWORKS=",
		"CONTAINER_ROOTFS_PATH=" + b.rootFSPath,
		"CONTAINER_DEPOT_PATH=" + b.depotPath,
		"CONTAINER_DEPOT_MOUNT_POINT_PATH=/",
		"DISK_QUOTA_ENABLED=true",

		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	}

	out, err := setup.CombinedOutput()
	if err != nil {
		fmt.Println("error setting up:", string(out))
		return err
	}

	println("done")

	return nil
}
func (b *LinuxBackend) Create(spec backend.ContainerSpec) (backend.Container, error) {
	resources, err := b.resourcePool.Acquire()
	if err != nil {
		return nil, err
	}

	create := exec.Command(
		path.Join(b.rootPath, "create.sh"),
		path.Join(b.depotPath, resources.ContainerID()),
	)

	create.Env = resources.Env()

	out, err := create.CombinedOutput()
	if err != nil {
		log.Println("error creating container:", string(out), create.ProcessState)
		return nil, err
	}

	start := exec.Command(
		path.Join(b.depotPath, resources.ContainerID(), "start.sh"),
	)

	start.Env = resources.Env()

	out, err = start.CombinedOutput()
	if err != nil {
		log.Println("error starting container:", string(out), start.ProcessState)
		return nil, err
	}

	container := &LinuxContainer{
		Spec:      spec,
		Resources: resources,
	}

	b.containers[container.Handle()] = container

	return container, nil
}

func (b *LinuxBackend) Destroy(handle string) error {
	container, found := b.containers[handle]
	if !found {
		return UnknownHandleError{handle}
	}

	destroy := exec.Command(
		path.Join(b.rootPath, "destroy.sh"),
		path.Join(b.depotPath, container.Resources.ContainerID()),
	)

	destroy.Env = container.Resources.Env()

	out, err := destroy.CombinedOutput()
	if err != nil {
		log.Println("error destroying container:", string(out), destroy.ProcessState)
		return err
	}

	delete(b.containers, container.Handle())

	return nil
}

func (b *LinuxBackend) Containers() (containers []backend.Container, err error) {
	for _, c := range b.containers {
		containers = append(containers, c)
	}

	return containers, nil
}

func (b *LinuxBackend) Lookup(handle string) (backend.Container, error) {
	container, found := b.containers[handle]
	if !found {
		return nil, UnknownHandleError{handle}
	}

	return container, nil
}

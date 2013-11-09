package linuxbackend

import (
	"github.com/vito/garden/backend"
	"github.com/vito/garden/backend/linuxbackend/resource_pool"
)

type LinuxBackend struct {
	resourcePool *resource_pool.ResourcePool

	containers map[string]*LinuxContainer
}

type UnknownHandleError struct {
	Handle string
}

func (e UnknownHandleError) Error() string {
	return "unknown handle: " + e.Handle
}

func New(resourcePool *resource_pool.ResourcePool) *LinuxBackend {
	return &LinuxBackend{
		resourcePool: resourcePool,

		containers: make(map[string]*LinuxContainer),
	}
}

func (b *LinuxBackend) Create(spec backend.ContainerSpec) (backend.Container, error) {
	resources, err := b.resourcePool.Acquire()
	if err != nil {
		return nil, err
	}

	container := NewLinuxContainer(spec, resources)

	err = container.Start()
	if err != nil {
		return nil, err
	}

	b.containers[container.Handle()] = container

	return container, nil
}

func (b *LinuxBackend) Destroy(handle string) error {
	container, found := b.containers[handle]
	if !found {
		return UnknownHandleError{handle}
	}

	err := b.resourcePool.Release(container.Resources)
	if err != nil {
		return nil
	}

	delete(b.containers, handle)

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

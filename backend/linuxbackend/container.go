package linuxbackend

import (
	"github.com/vito/garden/backend"
	"github.com/vito/garden/backend/linuxbackend/resource_pool"
)

type LinuxContainer struct {
	Spec      backend.ContainerSpec
	Resources *resource_pool.Resources
}

func (c *LinuxContainer) Handle() string {
	if c.Spec.Handle != "" {
		return c.Spec.Handle
	}

	return c.Resources.ContainerID()
}

func (c *LinuxContainer) Stop(bool, bool) error {
	return nil
}

func (c *LinuxContainer) Info() (backend.ContainerInfo, error) {
	return backend.ContainerInfo{}, nil
}

func (c *LinuxContainer) CopyIn(src, dst string) error {
	return nil
}

func (c *LinuxContainer) CopyOut(src, dst, owner string) error {
	return nil
}

func (c *LinuxContainer) LimitBandwidth(backend.BandwidthLimits) (backend.BandwidthLimits, error) {
	return backend.BandwidthLimits{}, nil
}

func (c *LinuxContainer) LimitDisk(backend.DiskLimits) (backend.DiskLimits, error) {
	return backend.DiskLimits{}, nil
}

func (c *LinuxContainer) LimitMemory(backend.MemoryLimits) (backend.MemoryLimits, error) {
	return backend.MemoryLimits{}, nil
}

func (c *LinuxContainer) Spawn(backend.JobSpec) (uint32, error) {
	return 0, nil
}

func (c *LinuxContainer) Stream(uint32) (<-chan backend.JobStream, error) {
	return nil, nil
}

func (c *LinuxContainer) Link(uint32) (backend.JobResult, error) {
	return backend.JobResult{}, nil
}

func (c *LinuxContainer) Run(backend.JobSpec) (backend.JobResult, error) {
	return backend.JobResult{}, nil
}

func (c *LinuxContainer) NetIn(uint32, uint32) (uint32, uint32, error) {
	return 0, 0, nil
}

func (c *LinuxContainer) NetOut(string, uint32) error {
	return nil
}

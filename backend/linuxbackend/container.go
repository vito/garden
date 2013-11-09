package linuxbackend

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path"
	"sync"
	"syscall"

	"github.com/vito/garden/backend"
	"github.com/vito/garden/backend/linuxbackend/resource_pool"
)

type LinuxContainer struct {
	Spec      backend.ContainerSpec
	Resources *resource_pool.Resources

	nextJobID uint32
	jobs      map[uint32]*exec.Cmd

	sync.RWMutex
}

type UnknownJobIDError struct {
	JobID uint32
}

func (e UnknownJobIDError) Error() string {
	return fmt.Sprintf("unknown job id: %d", e.JobID)
}

func NewLinuxContainer(spec backend.ContainerSpec, resources *resource_pool.Resources) *LinuxContainer {
	return &LinuxContainer{
		Spec:      spec,
		Resources: resources,

		jobs: make(map[uint32]*exec.Cmd),
	}
}

func (c *LinuxContainer) Handle() string {
	if c.Spec.Handle != "" {
		return c.Spec.Handle
	}

	return c.Resources.ContainerID()
}

func (c *LinuxContainer) Start() error {
	start := exec.Command(path.Join(c.Resources.ContainerPath(), "start.sh"))
	start.Env = c.Resources.Env()

	out, err := start.CombinedOutput()
	if err != nil {
		log.Println("error starting container:", string(out), start.ProcessState)
		return err
	}

	return nil
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

func (c *LinuxContainer) Spawn(spec backend.JobSpec) (uint32, error) {
	wshPath := path.Join(c.Resources.ContainerPath(), "bin", "wsh")
	wshdSocketPath := path.Join(c.Resources.ContainerPath(), "run", "wshd.sock")

	var user string

	if spec.Priveleged {
		user = "root"
	} else {
		user = "vcap"
	}

	cmd := exec.Command(wshPath, "--socket", wshdSocketPath, "--user", user, "/bin/bash")

	cmd.Stdin = bytes.NewBufferString(spec.Script)
	cmd.Env = []string{} // TODO: resource limits

	err := cmd.Start()
	if err != nil {
		return 0, err
	}

	c.Lock()
	defer c.Unlock()

	jobID := c.nextJobID

	c.nextJobID++

	c.jobs[jobID] = cmd

	return jobID, nil
}

func (c *LinuxContainer) Stream(uint32) (<-chan backend.JobStream, error) {
	return nil, nil
}

func (c *LinuxContainer) Link(jobID uint32) (backend.JobResult, error) {
	c.RLock()
	cmd, found := c.jobs[jobID]
	c.RUnlock()

	result := backend.JobResult{}

	if !found {
		return result, UnknownJobIDError{jobID}
	}

	err := cmd.Wait()
	if err != nil {
		result.ExitStatus = uint32(cmd.ProcessState.Sys().(syscall.WaitStatus))
	}

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

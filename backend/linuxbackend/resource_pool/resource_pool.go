package resource_pool

import (
	"strconv"
	"sync"
	"time"
)

type ResourcePool struct {
	defaultRootFSPath string

	nextContainer int64

	sync.RWMutex
}

func New(defaultRootFSPath string) *ResourcePool {
	return &ResourcePool{
		defaultRootFSPath: defaultRootFSPath,

		nextContainer: time.Now().UnixNano(),
	}
}

func (p *ResourcePool) Acquire() (*Resources, error) {
	p.Lock()
	defer p.Unlock()

	return &Resources{
		containerID: p.generateContainerID(),

		rootFS: p.defaultRootFSPath,
	}, nil
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

package resource_pool

import (
	"path"
)

type Resources struct {
	depotPath  string
	rootFSPath string

	containerID string
}

func (r *Resources) ContainerID() string {
	return r.containerID
}

func (r *Resources) ContainerPath() string {
	return path.Join(r.depotPath, r.containerID)
}

func (r *Resources) Env() []string {
	return []string{
		"id=" + r.containerID,
		"rootfs_path=" + r.rootFSPath,
		"allow_nested_warden=false",
		"container_iface_mtu=1500",

		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",

		// "network_host_ip" => host_ip.to_human,
		// "network_container_ip" => container_ip.to_human,
		// "network_netmask" => self.class.network_pool.pooled_netmask.to_human,
		// "user_uid" => uid,
	}
}

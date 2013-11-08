package resource_pool

type Resources struct {
	containerID string
	rootFS      string
}

func (r *Resources) ContainerID() string {
	return r.containerID
}

func (r *Resources) Env() []string {
	return []string{
		"id=" + r.containerID,
		"rootfs_path=" + r.rootFS,
		"allow_nested_warden=false",
		"container_iface_mtu=1500",

		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",

		// "id" => container_id,
		// "network_host_ip" => host_ip.to_human,
		// "network_container_ip" => container_ip.to_human,
		// "network_netmask" => self.class.network_pool.pooled_netmask.to_human,
		// "user_uid" => uid,
		// "rootfs_path" => container_rootfs_path,
		// "allow_nested_warden" => Server.config.allow_nested_warden?.to_s,
		// "container_iface_mtu" => container_iface_mtu,
	}
}

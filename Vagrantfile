# vim: set ft=ruby

Vagrant.configure("2") do |config|
  config.vm.hostname = "garden"

  config.vm.box = "garden"

  config.vm.network "private_network", ip: "192.168.50.5"

  5.times do |i|
    config.vm.network "forwarded_port", guest: 7012 + i, host: 7012 + i
  end

  config.vm.synced_folder ENV["GOPATH"], "/go"

  config.vm.provider :virtualbox do |v, override|
    v.customize ["modifyvm", :id, "--memory", 3*1024]
    v.customize ["modifyvm", :id, "--cpus", 4]
  end

  config.vm.provider :vmware_fusion do |v, override|
    override.vm.box_url = "./boxes/packer_vmware-iso_vmware.box"
    v.vmx["memsize"] = 3 * 1024
    v.vmx["numvcpus"] = "4"
  end
end

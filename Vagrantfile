Vagrant.require_version ">= 2.2.19"
Vagrant.configure("2") do |config|
  config.vm.define "xprog" # name in vagrant CLI
  config.vm.box = "debian/bullseye64" # debian/11
  config.vm.hostname = "xprog"
  config.vm.synced_folder ".", "/vagrant", disabled: true
  config.vm.synced_folder ".", "/home/vagrant/xprog"
  config.vm.box_check_update = false

  config.vm.provider "virtualbox" do |v|
    v.name = "xprog" # name in VirtualBox GUI
    v.linked_clone = true
    v.check_guest_additions = false
    v.memory = 2048
    v.cpus = 2
  end
end

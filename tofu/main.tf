terraform {
  required_providers {
    proxmox = {
      source  = "bpg/proxmox"
      version = "~> 0.78"
    }
  }
}

provider "proxmox" {
  insecure = true
}

# --- Containers ---

resource "proxmox_virtual_environment_container" "dev_ct" {
  node_name     = "srv2"
  vm_id         = 110
  unprivileged  = true
  start_on_boot = false

  console {
    enabled   = true
    tty_count = 2
    type      = "tty"
  }

  cpu {
    architecture = "amd64"
    cores        = 2
  }

  memory {
    dedicated = 8192
    swap      = 1024
  }

  disk {
    datastore_id = "local-lvm"
    size         = 32
  }

  features {
    nesting = true
  }

  initialization {
    hostname = "dev-ct"

    ip_config {
      ipv4 {
        address = "10.0.10.110/24"
        gateway = "10.0.10.1"
      }
    }
  }

  network_interface {
    name     = "eth0"
    bridge   = "vmbr10"
    firewall = true
  }

  operating_system {
    template_file_id = "local:vztmpl/debian-13-standard_13.6-1_amd64.tar.zst"
    type             = "debian"
  }

  lifecycle {
    ignore_changes = [operating_system[0].template_file_id]
  }
}

resource "proxmox_virtual_environment_container" "minecraft_ct" {
  node_name    = "srv2"
  vm_id        = 111
  unprivileged = true

  console {
    enabled   = true
    tty_count = 2
    type      = "tty"
  }

  cpu {
    architecture = "amd64"
    cores        = 2
  }

  memory {
    dedicated = 8192
    swap      = 1024
  }

  disk {
    datastore_id = "local-lvm"
    size         = 64
  }

  features {
    nesting = true
  }

  initialization {
    hostname = "minecraft-ct"

    ip_config {
      ipv4 {
        address = "10.0.10.111/24"
        gateway = "10.0.10.1"
      }
    }
  }

  mount_point {
    path   = "/mnt/srv1-storage"
    volume = "/mnt/storage"
    backup = false
  }

  network_interface {
    name     = "eth0"
    bridge   = "vmbr10"
    firewall = true
  }

  operating_system {
    template_file_id = "local:vztmpl/debian-13-standard_13.6-1_amd64.tar.zst"
    type             = "debian"
  }

  lifecycle {
    ignore_changes = [operating_system[0].template_file_id]
  }
}

resource "proxmox_virtual_environment_container" "nginx_ct" {
  node_name    = "srv2"
  vm_id        = 120
  unprivileged = true

  console {
    enabled   = true
    tty_count = 2
    type      = "tty"
  }

  memory {
    dedicated = 512
    swap      = 512
  }

  disk {
    datastore_id = "local-lvm"
  }

  features {
    nesting = true
  }

  initialization {
    hostname = "nginx-ct"

    ip_config {
      ipv4 {
        address = "10.0.10.120/24"
        gateway = "10.0.10.1"
      }
    }
  }

  network_interface {
    name     = "eth0"
    bridge   = "vmbr10"
    firewall = true
  }

  operating_system {
    template_file_id = "local:vztmpl/debian-13-standard_13.6-1_amd64.tar.zst"
    type             = "debian"
  }

  lifecycle {
    ignore_changes = [operating_system[0].template_file_id]
  }
}

resource "proxmox_virtual_environment_container" "monitoring_ct" {
  node_name    = "srv2"
  vm_id        = 129
  unprivileged = true

  console {
    enabled   = true
    tty_count = 2
    type      = "tty"
  }

  memory {
    dedicated = 2048
    swap      = 512
  }

  disk {
    datastore_id = "local-lvm"
    size         = 8
  }

  features {
    nesting = true
  }

  initialization {
    hostname = "monitoring-ct"

    ip_config {
      ipv4 {
        address = "10.0.10.129/24"
        gateway = "10.0.10.1"
      }
    }
  }

  network_interface {
    name     = "eth0"
    bridge   = "vmbr10"
    firewall = true
  }

  operating_system {
    template_file_id = "local:vztmpl/debian-13-standard_13.6-1_amd64.tar.zst"
    type             = "debian"
  }

  lifecycle {
    ignore_changes = [operating_system[0].template_file_id]
  }
}

# --- Virtual Machines ---

resource "proxmox_virtual_environment_vm" "pbs_vm" {
  node_name     = "srv2"
  vm_id         = 113
  name          = "pbs-vm"
  scsi_hardware = "virtio-scsi-single"
  on_boot       = true

  agent {
    enabled = true
    timeout = "15m"
    type    = "virtio"
  }

  cpu {
    cores   = 2
    sockets = 1
    type    = "x86-64-v2-AES"
  }

  memory {
    dedicated = 4096
  }

  operating_system {
    type = "l26"
  }
}

resource "proxmox_virtual_environment_vm" "torrent" {
  node_name     = "srv2"
  vm_id         = 114
  name          = "torrent"
  scsi_hardware = "virtio-scsi-single"
  on_boot       = true

  agent {
    enabled = true
    timeout = "15m"
    type    = "virtio"
  }

  cpu {
    cores   = 2
    sockets = 1
    type    = "x86-64-v2-AES"
  }

  memory {
    dedicated = 4096
  }

  operating_system {
    type = "l26"
  }
}

resource "proxmox_virtual_environment_vm" "steam_vm" {
  node_name     = "srv2"
  vm_id         = 119
  name          = "steam-vm"
  started       = false
  on_boot       = false
  scsi_hardware = "virtio-scsi-single"

  agent {
    enabled = true
    timeout = "15m"
    type    = "virtio"
  }

  cpu {
    cores   = 4
    sockets = 1
    type    = "x86-64-v2-AES"
  }

  memory {
    dedicated = 32768
  }

  hostpci {
    device = "hostpci0"
    id     = "0000:65:00"
    rombar = true
    xvga   = true
  }

  serial_device {
    device = "socket"
  }

  usb {
    host = "1-7"
  }

  usb {
    host = "1-8"
  }

  operating_system {
    type = "l26"
  }
}

resource "proxmox_virtual_environment_vm" "postgres_vm" {
  node_name     = "srv2"
  vm_id         = 121
  name          = "postgres-vm"
  scsi_hardware = "virtio-scsi-single"
  on_boot       = true

  agent {
    enabled = true
    timeout = "15m"
    type    = "virtio"
  }

  cpu {
    cores   = 4
    sockets = 1
    type    = "x86-64-v2-AES"
  }

  memory {
    dedicated = 12288
  }

  operating_system {
    type = "l26"
  }
}

resource "proxmox_virtual_environment_vm" "services_vm" {
  node_name     = "srv2"
  vm_id         = 123
  name          = "services-vm"
  scsi_hardware = "virtio-scsi-single"
  on_boot       = true

  agent {
    enabled = true
    timeout = "15m"
    type    = "virtio"
  }

  cpu {
    cores   = 2
    sockets = 1
    type    = "x86-64-v2-AES"
  }

  memory {
    dedicated = 6144
  }

  operating_system {
    type = "l26"
  }
}

resource "proxmox_virtual_environment_vm" "stock_analysis_vm" {
  node_name     = "srv2"
  vm_id         = 127
  name          = "stock-analysis-vm"
  started       = false
  on_boot       = false
  scsi_hardware = "virtio-scsi-single"

  agent {
    enabled = true
    timeout = "15m"
    type    = "virtio"
  }

  cpu {
    cores   = 6
    sockets = 1
    type    = "x86-64-v2-AES"
  }

  memory {
    dedicated = 12288
  }

  operating_system {
    type = "l26"
  }
}

resource "proxmox_virtual_environment_vm" "stock_analysis_worker_vm" {
  node_name     = "srv2"
  vm_id         = 128
  name          = "stock-analysis-worker-vm"
  started       = false
  on_boot       = false
  scsi_hardware = "virtio-scsi-single"

  agent {
    enabled = true
    timeout = "15m"
    type    = "virtio"
  }

  cpu {
    cores   = 6
    sockets = 1
    type    = "x86-64-v2-AES"
  }

  memory {
    dedicated = 49152
  }

  operating_system {
    type = "l26"
  }
}

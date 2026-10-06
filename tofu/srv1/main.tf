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

resource "proxmox_virtual_environment_container" "jellyfin_ct" {
  node_name    = "srv1"
  vm_id        = 123
  unprivileged = false

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
    dedicated = 4096
    swap      = 512
  }

  disk {
    datastore_id = "local-lvm"
    size         = 16
  }

  features {
    nesting = true
  }

  initialization {
    hostname = "jellyfin-ct"

    ip_config {
      ipv4 {
        address = "10.0.10.123/24"
        gateway = "10.0.10.1"
      }
    }
  }

  mount_point {
    path   = "/media"
    volume = "/mnt/hdd-3tb/jellyfin/data"
    backup = false
  }

  network_interface {
    name     = "eth0"
    bridge   = "vmbr0"
    vlan_id  = 10
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

resource "proxmox_virtual_environment_vm" "torrent_vm" {
  node_name     = "srv1"
  vm_id         = 114
  name          = "torrent-vm"
  started       = false
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
    dedicated = 8192
  }

  network_device {
    bridge   = "vmbr0"
    vlan_id  = 10
    firewall = true
  }

  operating_system {
    type = "l26"
  }
}

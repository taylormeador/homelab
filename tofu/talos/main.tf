terraform {
  required_providers {
    proxmox = {
      source  = "bpg/proxmox"
      version = "~> 0.116"
    }
    talos = {
      source  = "siderolabs/talos"
      version = "~> 0.12"
    }
  }
}

# --- Proxmox providers (one per non-clustered host) ---

provider "proxmox" {
  alias     = "srv1"
  endpoint  = "https://srv1.lab.taylor-meador.com:8006"
  api_token = var.srv1_api_token
  insecure  = true
}

provider "proxmox" {
  alias     = "srv2"
  endpoint  = "https://srv2.lab.taylor-meador.com:8006"
  api_token = var.srv2_api_token
  insecure  = true
}

provider "proxmox" {
  alias     = "srv3"
  endpoint  = "https://srv3.lab.taylor-meador.com:8006"
  api_token = var.srv3_api_token
  insecure  = true
}

variable "srv1_api_token" {
  type      = string
  sensitive = true
}

variable "srv2_api_token" {
  type      = string
  sensitive = true
}

variable "srv3_api_token" {
  type      = string
  sensitive = true
}

# --- Talos config ---

locals {
  cluster_name     = "dexter"
  cluster_endpoint = "https://10.0.30.10:6443"
  talos_version    = "v1.14.2"
  schematic_id     = "53513e54bb39202f35694412577a6bc53d484744d35a126e5d42ef34785c0d83"
  talos_iso        = "hdd-6tb:iso/metal-amd64.iso"
  gateway          = "10.0.30.1"
  vip              = "10.0.30.10"

  control_nodes = {
    "talos-cp-1" = { ip = "10.0.30.11", host = "srv1", vm_id = 311 }
    "talos-cp-2" = { ip = "10.0.30.12", host = "srv2", vm_id = 312 }
    "talos-cp-3" = { ip = "10.0.30.13", host = "srv3", vm_id = 313 }
  }

  worker_nodes = {
    "talos-w-1" = { ip = "10.0.30.21", host = "srv1", vm_id = 321 }
    "talos-w-2" = { ip = "10.0.30.22", host = "srv2", vm_id = 322 }
  }

  control_node_ips = [for n in local.control_nodes : n.ip]
  worker_node_ips  = [for n in local.worker_nodes : n.ip]
  all_node_ips     = concat(local.control_node_ips, local.worker_node_ips)

  # Discover DHCP IPs from the QEMU guest agent. The agent reports many
  # interfaces (lo, bond0, dummy0, etc.) before the real NIC — pick the
  # first one with a routable IPv4.
  cni_prefixes = ["flannel", "cni", "cali", "veth", "docker", "kube-"]

  vm_discovered_ip = {
    for name, vm in {
      "talos-cp-1" = proxmox_virtual_environment_vm.cp_1
      "talos-cp-2" = proxmox_virtual_environment_vm.cp_2
      "talos-cp-3" = proxmox_virtual_environment_vm.cp_3
      "talos-w-1"  = proxmox_virtual_environment_vm.w_1
      "talos-w-2"  = proxmox_virtual_environment_vm.w_2
    } :
    name => [
      for i, iface in vm.network_interface_names : vm.ipv4_addresses[i][0]
      if length(vm.ipv4_addresses[i]) > 0
        && !startswith(vm.ipv4_addresses[i][0], "127.")
        && !startswith(vm.ipv4_addresses[i][0], "169.254.")
        && length([for p in local.cni_prefixes : true if startswith(iface, p)]) == 0
    ][0]
  }

  install_image = "factory.talos.dev/metal-installer/${local.schematic_id}:${local.talos_version}"

  common_config_patches = [
    yamlencode({
      apiVersion = "v1alpha1"
      kind       = "HostnameConfig"
      auto       = { "$patch" = "delete" }
    }),
    yamlencode({
      machine = {
        features = {
          diskQuotaSupport = true
        }
      }
    }),
    yamlencode({
      apiVersion = "v1alpha1"
      kind       = "UnattendedInstallConfig"
      installer = {
        image = local.install_image
      }
      provisioning = {
        diskSelector = {
          match = "disk.dev_path == \"/dev/sda\""
        }
      }
    }),
  ]
}

# --- Talos secrets and machine configs ---

resource "talos_machine_secrets" "this" {
  talos_version = local.talos_version
}

data "talos_machine_configuration" "controlplane" {
  cluster_name     = local.cluster_name
  machine_type     = "controlplane"
  cluster_endpoint = local.cluster_endpoint
  machine_secrets  = talos_machine_secrets.this.machine_secrets
  talos_version    = local.talos_version
}

data "talos_machine_configuration" "worker" {
  cluster_name     = local.cluster_name
  machine_type     = "worker"
  cluster_endpoint = local.cluster_endpoint
  machine_secrets  = talos_machine_secrets.this.machine_secrets
  talos_version    = local.talos_version
}

data "talos_client_configuration" "this" {
  cluster_name         = local.cluster_name
  client_configuration = talos_machine_secrets.this.client_configuration
  endpoints            = local.control_node_ips
  nodes                = local.all_node_ips
}

# --- VMs (separate blocks per provider since hosts aren't clustered) ---

resource "proxmox_virtual_environment_vm" "cp_1" {
  provider      = proxmox.srv1
  node_name     = "srv1"
  vm_id         = 311
  name          = "talos-cp-1"
  on_boot       = true
  scsi_hardware = "virtio-scsi-single"

  agent {
    enabled = true
  }

  cpu {
    cores   = 2
    sockets = 1
    type    = "host"
  }

  memory {
    dedicated = 2048
  }

  disk {
    datastore_id = "local-lvm"
    interface    = "scsi0"
    size         = 15
    file_format  = "raw"
  }

  cdrom {
    file_id   = local.talos_iso
    interface = "ide2"
  }

  network_device {
    bridge  = "vmbr0"
    vlan_id = 30
  }

  operating_system {
    type = "l26"
  }
}

resource "proxmox_virtual_environment_vm" "cp_2" {
  provider      = proxmox.srv2
  node_name     = "srv2"
  vm_id         = 312
  name          = "talos-cp-2"
  on_boot       = true
  scsi_hardware = "virtio-scsi-single"

  agent {
    enabled = true
  }

  cpu {
    cores   = 2
    sockets = 1
    type    = "host"
  }

  memory {
    dedicated = 2048
  }

  disk {
    datastore_id = "local-lvm"
    interface    = "scsi0"
    size         = 15
    file_format  = "raw"
  }

  cdrom {
    file_id   = local.talos_iso
    interface = "ide2"
  }

  network_device {
    bridge  = "vmbr0"
    vlan_id = 30
  }

  operating_system {
    type = "l26"
  }
}

resource "proxmox_virtual_environment_vm" "cp_3" {
  provider      = proxmox.srv3
  node_name     = "srv3"
  vm_id         = 313
  name          = "talos-cp-3"
  on_boot       = true
  scsi_hardware = "virtio-scsi-single"

  agent {
    enabled = true
  }

  cpu {
    cores   = 2
    sockets = 1
    type    = "host"
  }

  memory {
    dedicated = 2048
  }

  disk {
    datastore_id = "local-lvm"
    interface    = "scsi0"
    size         = 15
    file_format  = "raw"
  }

  cdrom {
    file_id   = local.talos_iso
    interface = "ide2"
  }

  network_device {
    bridge  = "vmbr0"
    vlan_id = 30
  }

  operating_system {
    type = "l26"
  }
}

resource "proxmox_virtual_environment_vm" "w_1" {
  provider      = proxmox.srv1
  node_name     = "srv1"
  vm_id         = 321
  name          = "talos-w-1"
  on_boot       = true
  scsi_hardware = "virtio-scsi-single"

  agent {
    enabled = true
  }

  cpu {
    cores   = 8
    sockets = 1
    type    = "host"
  }

  memory {
    dedicated = 8192
  }

  disk {
    datastore_id = "local-lvm"
    interface    = "scsi0"
    size         = 15
    file_format  = "raw"
  }

  disk {
    datastore_id = "local-lvm"
    interface    = "scsi1"
    size         = 75
    file_format  = "raw"
  }

  cdrom {
    file_id   = local.talos_iso
    interface = "ide2"
  }

  network_device {
    bridge  = "vmbr0"
    vlan_id = 30
  }

  operating_system {
    type = "l26"
  }
}

resource "proxmox_virtual_environment_vm" "w_2" {
  provider      = proxmox.srv2
  node_name     = "srv2"
  vm_id         = 322
  name          = "talos-w-2"
  on_boot       = true
  scsi_hardware = "virtio-scsi-single"

  agent {
    enabled = true
  }

  cpu {
    cores   = 8
    sockets = 1
    type    = "host"
  }

  memory {
    dedicated = 8192
  }

  disk {
    datastore_id = "local-lvm"
    interface    = "scsi0"
    size         = 15
    file_format  = "raw"
  }

  disk {
    datastore_id = "local-lvm"
    interface    = "scsi1"
    size         = 75
    file_format  = "raw"
  }

  cdrom {
    file_id   = local.talos_iso
    interface = "ide2"
  }

  network_device {
    bridge  = "vmbr0"
    vlan_id = 30
  }

  operating_system {
    type = "l26"
  }
}

# --- Apply Talos configs ---
# Each node gets the base config plus per-node patches for static IP,
# hostname, and VIP (control plane only).

resource "talos_machine_configuration_apply" "controlplane" {
  for_each = local.control_nodes

  client_configuration        = talos_machine_secrets.this.client_configuration
  machine_configuration_input = data.talos_machine_configuration.controlplane.machine_configuration
  node                        = local.vm_discovered_ip[each.key]
  endpoint                    = local.vm_discovered_ip[each.key]

  config_patches = concat(local.common_config_patches, [
    yamlencode({
      apiVersion = "v1alpha1"
      kind       = "LinkConfig"
      name       = "ens18"
      addresses  = [{ address = "${each.value.ip}/24" }]
      routes     = [{ gateway = local.gateway }]
    }),
    yamlencode({
      apiVersion = "v1alpha1"
      kind       = "HostnameConfig"
      hostname   = each.key
    }),
    yamlencode({
      apiVersion = "v1alpha1"
      kind       = "Layer2VIPConfig"
      name       = local.vip
      link       = "ens18"
    }),
  ])

  depends_on = [
    proxmox_virtual_environment_vm.cp_1,
    proxmox_virtual_environment_vm.cp_2,
    proxmox_virtual_environment_vm.cp_3,
  ]
}

resource "talos_machine_bootstrap" "this" {
  depends_on = [
    talos_machine_configuration_apply.controlplane,
  ]

  node                 = local.control_nodes["talos-cp-1"].ip
  endpoint             = local.control_nodes["talos-cp-1"].ip
  client_configuration = talos_machine_secrets.this.client_configuration
}

resource "talos_machine_configuration_apply" "worker" {
  for_each = local.worker_nodes

  depends_on = [
    talos_machine_bootstrap.this,
  ]

  client_configuration        = talos_machine_secrets.this.client_configuration
  machine_configuration_input = data.talos_machine_configuration.worker.machine_configuration
  node                        = local.vm_discovered_ip[each.key]
  endpoint                    = local.vm_discovered_ip[each.key]

  config_patches = concat(local.common_config_patches, [
    yamlencode({
      apiVersion = "v1alpha1"
      kind       = "LinkConfig"
      name       = "ens18"
      addresses  = [{ address = "${each.value.ip}/24" }]
      routes     = [{ gateway = local.gateway }]
    }),
    yamlencode({
      apiVersion = "v1alpha1"
      kind       = "HostnameConfig"
      hostname   = each.key
    }),
  ])
}

resource "talos_cluster_kubeconfig" "this" {
  depends_on = [
    talos_machine_bootstrap.this,
  ]

  client_configuration = talos_machine_secrets.this.client_configuration
  node                 = local.control_nodes["talos-cp-1"].ip
}

# --- Outputs ---

output "talosconfig" {
  value     = data.talos_client_configuration.this.talos_config
  sensitive = true
}

output "kubeconfig" {
  value     = talos_cluster_kubeconfig.this.kubeconfig_raw
  sensitive = true
}

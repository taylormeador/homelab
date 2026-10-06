# Network and VLANs

## Physical Topology

Two Proxmox hypervisors (srv1, srv2) connected to Netgear GS308E managed switches. OPNsense router handles inter-VLAN routing and is the default gateway for all VLANs.

```
Internet
   |
OPNsense (10.0.10.1) ── VLAN trunk
   |
GS308E switches ── tagged uplinks between switches
   |         |
  srv1      srv2
```

## VLANs

| VLAN | Purpose             | Subnet        | Gateway    |
|------|---------------------|---------------|------------|
| 1    | Switch management   | (default)     |            |
| 10   | Wired infrastructure| 10.0.10.0/24  | 10.0.10.1  |
| 20   | WiFi                | (separate)    | 10.0.10.1  |

Future candidates: K8s/portal network, DMZ for public-facing services.

## VLAN-Aware Bridge Architecture

Both hypervisors use a single VLAN-aware bridge (`vmbr0`) instead of per-VLAN bridges. This is the Proxmox-recommended approach.

### How it works

```
vmbr0 (VLAN-aware bridge)
├── nic0 / eno1          physical NIC, carries all tagged VLAN traffic
├── vmbr0.10             host management IP (how the host itself joins VLAN 10)
├── guest vlan_id=10     bridge tags guest traffic with VLAN 10
├── guest vlan_id=20     bridge tags guest traffic with VLAN 20
└── etc.
```

**Guests** are assigned to a VLAN in Proxmox (GUI or tofu `vlan_id`). The bridge adds/removes VLAN tags automatically. Guests never see the tags — they just send and receive normal frames.

**The host** needs a VLAN subinterface (`vmbr0.10`) to participate in a VLAN itself. Without it, the host's management traffic goes out untagged and gets dropped by the switch.

### `/etc/network/interfaces` template

```
auto vmbr0
iface vmbr0 inet manual
        bridge-ports nic0       # or eno1 — the physical NIC
        bridge-stp off
        bridge-fd 0
        bridge-vlan-aware yes
        bridge-vids 10 20       # all VLANs the bridge should carry

auto vmbr0.10
iface vmbr0.10 inet static
        address 10.0.10.X/24   # host's management IP
        gateway 10.0.10.1
```

Key points:
- `vmbr0` has **no IP** (`inet manual`) — the IP lives on `vmbr0.10`
- `bridge-vlan-aware yes` enables VLAN filtering on the bridge
- `bridge-vids` lists VLANs the bridge carries; add new VLANs here
- `vmbr0.10` is a Linux VLAN subinterface, not a bridge — create it as "Linux VLAN" in the Proxmox GUI


## Switch Configuration (GS308E)

The GS308E is web-GUI-only — no SNMP, no API, no Ansible management. Configuration must be done manually through its web interface.


### Port configuration

| Port role         | VLAN 10 | VLAN 20 | PVID | Notes                           |
|-------------------|---------|---------|------|---------------------------------|
| Hypervisor        | Tagged  | Tagged  | —    | All traffic is tagged           |
| OPNsense trunk    | Tagged  | Tagged  | —    | Carries all VLANs               |
| Inter-switch link | Tagged  | Tagged  | —    | Must trunk all VLANs            |
| Wired device      | Untagged| Exclude | 10   | Device doesn't know about VLANs |

Hypervisor ports carry only tagged traffic because the VLAN-aware bridge and `vmbr0.10` subinterface both tag their frames. There is no untagged traffic on these ports.

## OpenTofu

Guests specify their VLAN through the bridge and optional VLAN tag:

**Containers:**
```hcl
network_interface {
  name     = "eth0"
  bridge   = "vmbr0"
  vlan_id  = 10
  firewall = true
}
```

**VMs:**
```hcl
network_device {
  bridge   = "vmbr0"
  vlan_id  = 10
  firewall = true
}
```

## Adding a New VLAN

1. **OPNsense**: create a VLAN interface, assign a subnet, add firewall rules for inter-VLAN traffic
2. **Switches**: add the VLAN as tagged on trunk/hypervisor ports, untagged on access ports
3. **Hypervisors**: add the VID to `bridge-vids` in `/etc/network/interfaces`, apply config
4. **Guests**: set `vlan_id` in tofu config or Proxmox GUI

## Migrating srv2

srv2 currently uses the older per-VLAN bridge approach (`eno1.10` → `vmbr10`, `eno1.20` → `vmbr20`). This works but creates extra bridges and doesn't let you manage VLANs per-guest in the GUI as cleanly.

Migration plan:
1. Create `vmbr0` as a VLAN-aware bridge on `eno1` with `bridge-vids 10 20`
2. Create `vmbr0.10` with srv2's management IP (`10.0.10.102/24`)
3. Update each guest to use `vmbr0` with the appropriate `vlan_id` instead of `vmbr10`/`vmbr20`
4. Update switch port from untagged to tagged
5. Remove old bridges (`vmbr10`, `vmbr20`) and VLAN subinterfaces (`eno1.10`, `eno1.20`)
6. Update tofu configs to use `vmbr0` + `vlan_id` instead of `vmbrN`

This is a disruptive change — every guest loses connectivity briefly during the switchover. Plan for a maintenance window.

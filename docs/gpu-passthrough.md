# GPU Passthrough

RX 5600 XT (Navi 10) passed through to Nobara VM (ID 115) on srv1 for gaming.

## Prerequisites

- **IOMMU**: enabled via `intel_iommu=on iommu=pt` in GRUB (already set on srv1)
- **VT-d and ACS**: enabled in BIOS (P520c: F1 to enter BIOS setup)
- **Secure Boot**: disabled (required for vendor-reset unsigned kernel module)

## GPU Details

```
67:00.0 VGA  - AMD Navi 10 [RX 5600 XT]  [1002:731f]  IOMMU group 4
67:00.1 Audio - AMD Navi 10 HDMI Audio    [1002:ab38]  IOMMU group 5
```

Both functions are in separate IOMMU groups with no other devices — clean isolation, no ACS override needed.

## Host Configuration (srv1)

### VFIO binding

`/etc/modprobe.d/vfio.conf`:
```
options vfio-pci ids=1002:731f,1002:ab38
```

### Blacklist host GPU driver

`/etc/modprobe.d/blacklist-amdgpu.conf`:
```
blacklist amdgpu
```

### Kernel modules

Added to `/etc/modules`:
```
vfio
vfio_iommu_type1
vfio_pci
vendor-reset
```

After changes: `update-initramfs -u -k all` and reboot.

### Verify VFIO claimed the GPU

```bash
lspci -k -s 67:00.0
# Should show: Kernel driver in use: vfio-pci
```

## AMD Reset Bug

Navi GPUs (RX 5000/5600/5700 series) can't reset properly after VM shutdown, preventing the VM from starting again without a host reboot. The `vendor-reset` DKMS module fixes this.

### Install vendor-reset

```bash
apt install pve-headers-$(uname -r) dkms git
git clone https://github.com/gnif/vendor-reset.git /opt/vendor-reset
cd /opt/vendor-reset
dkms install .
echo "vendor-reset" >> /etc/modules
```

Requires Secure Boot disabled (unsigned module). If `modprobe vendor-reset` fails with "Key was rejected by service", check `mokutil --sb-state`.

## Proxmox VM Configuration

Add PCI device in the Proxmox GUI (Hardware > Add > PCI Device):
- **Raw Device**: `0000:67:00.0`
- **All Functions**: checked (grabs VGA + audio)
- **PCI-Express**: checked
- **Primary GPU**: checked (VM outputs to the physical GPU, VNC no longer works)

## Important Notes

- With Primary GPU enabled, the VM display goes to the physical monitor connected to the RX 5600 XT. VNC/console in Proxmox GUI will be blank.
- SSH into the VM for admin: `ssh nobara-vm.lab.taylor-meador.com`
- The BIOS screen still displays on the GPU during POST (before Linux loads VFIO), so you can enter BIOS setup even with the GPU passed through.
- srv1 has no display output once Linux boots — manage via SSH or the Proxmox web UI on another machine.

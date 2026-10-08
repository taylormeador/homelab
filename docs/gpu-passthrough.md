# GPU Passthrough

RX 5600 XT (Navi 10) on srv1, shared between gaming VMs. Only one VM can use the GPU at a time.

## Gaming VMs

| VM | ID | IP | OS | Notes |
|----|----|----|----|----|
| nobara-vm | 115 | 10.0.10.115 | Nobara (Fedora) | Traditional desktop, BedrockOnLinux |
| bazzite-vm | 116 | 10.0.10.116 | Bazzite (Fedora Atomic) | HTPC/console mode, controller-friendly, immutable OS |

Bazzite is the primary gaming VM. Uses Distrobox for installing software (e.g. BedrockOnLinux) since the base OS is immutable. Apps are exported from Distrobox and added as non-Steam games for Gaming Mode.

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
- SSH into the VM for admin: `ssh bazzite-vm.lab.taylor-meador.com` (or `nobara-vm`)
- The BIOS screen still displays on the GPU during POST (before Linux loads VFIO), so you can enter BIOS setup even with the GPU passed through.
- srv1 has no display output once Linux boots — manage via SSH or the Proxmox web UI on another machine.
- Bazzite VMs require OVMF (UEFI) + EFI disk — Fedora-based ISOs don't support legacy BIOS boot.
- Bazzite is excluded from the Alloy logging playbook (`!bazzite-vm`) since it's immutable and can't run apt. Use Proxmox's built-in monitoring instead.

## USB Passthrough

Pass through input devices by vendor/device ID (Hardware > Add > USB Device):

| Device | Vendor:Device |
|--------|--------------|
| Keychron K4 Pro | 3434:0240 |
| Logitech USB Receiver | 046d:c53f |
| 8BitDo Ultimate 2 (2.4GHz dongle) | 2dc8:310b |
| 8BitDo Ultimate 2 (wired) | 2dc8:6013 |

Using vendor/device ID instead of port means they work across unplug/replug cycles.

### 8BitDo Controller Passthrough Fix

The 8BitDo Ultimate 2 dongle won't pass through by default because the host's `xpad` and `usbhid` drivers claim it before QEMU can. Other USB devices bind to `usbfs` (QEMU's userspace driver) automatically, but gamepads get grabbed by the kernel.

**Fix:** Create a udev rule on the Proxmox host to prevent any driver from claiming the device:

`/etc/udev/rules.d/99-8bitdo-passthrough.rules`:
```
SUBSYSTEM=="usb", ATTR{idVendor}=="2dc8", ATTR{idProduct}=="310b", ENV{DEVTYPE}=="usb_device", OPTIONS+="ignore_device"
```

Reload rules:
```bash
udevadm control --reload-rules
```

If using wired mode (product ID 6013), add a second rule for that ID.

**Notes:**
- Controller mode switch must be set to 2.4G for the dongle to work
- The 2.4GHz dongle pairs directly to the controller — no OS-level pairing needed
- Steam detects the controller automatically once the OS sees it as a gamepad

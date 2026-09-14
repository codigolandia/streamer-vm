# streamer-vm

*[English](README.md) | [Português](README.pt-br.md)*

`streamer-vm` is a Pure Go CLI tool designed to manage QEMU+KVM virtual machines with VirGL GPU acceleration (OpenGL pass-through), SPICE display, and native integration for OBS Studio recording and live streaming on Linux.

## Features

- **VirGL GPU Acceleration**: Native hardware OpenGL acceleration via the host GPU (e.g., AMD Radeon / Intel / NVIDIA through DRM scanout).
- **Zero External Dependencies**: Standalone single binary compiled in Pure Go using only the standard library.
- **Internationalization (i18n)**: Native English (`en`) and Portuguese (`pt`) language support with automatic locale detection or via the `--lang` flag.
- **Qcow2 Overlay Isolation**: Keeps the base disk image intact and uses writable COW overlays for easy and instantaneous `reset`.
- **SPICE Protocol**: Low-latency display, Pipewire audio passthrough, bidirectional clipboard sharing (`vdagent`), and USB device redirection (XHCI).
- **Tailored for Linux**: Optimized exclusively for Linux desktop systems with Wayland/X11 and QEMU/KVM.

## Prerequisites

On Linux host:
- `QEMU` (with KVM and VirGL support: `qemu-system-x86_64`)
- `qemu-img`
- UEFI `OVMF` firmware (`OVMF_CODE_4M.fd` and `OVMF_VARS_4M.fd`)
- Optional: `spice-client-gtk` (`spicy` or `virt-viewer`) for direct desktop window VM display

## Installation

### Quick Install / Update

Install or update to the latest release into `~/.local/bin` or `~/bin` (whichever is first in `$PATH`):

```bash
curl -fsSL https://github.com/codigolandia/streamer-vm/releases/latest/download/install.sh | bash
```

To roll back to the previously installed version (restoring `streamer-vm.backup`):

```bash
curl -fsSL https://github.com/codigolandia/streamer-vm/releases/latest/download/install.sh | bash -s -- --rollback
```

### Manual Download or Build from Source

Download the pre-compiled binary for Linux (amd64 / arm64) from [Releases](https://github.com/codigolandia/streamer-vm/releases) or build from source:

```bash
go build -o streamer-vm .
```

## Quick Start

```bash
# 1. Initialize environment directories and verify host prerequisites
streamer-vm init

# 2. Create a new VM with an installation ISO (initial setup mode, writing directly to base disk)
streamer-vm create ubuntu-live -cpus 4 -memory 8 -disk 50 -iso ~/Downloads/ubuntu-24.04.iso

# 3. Start VM and automatically launch SPICE client GUI (spicy)
streamer-vm start ubuntu-live --gui

# 4. View SPICE connection URL (if using spicy or virt-viewer separately)
streamer-vm spice-url ubuntu-live

# 5. After OS installation is complete, gracefully shut down the VM (or power off from guest OS)
streamer-vm stop ubuntu-live

# 6. Commit the installed OS as golden base image, detach ISO, and activate COW overlay
streamer-vm commit ubuntu-live --remove-iso

# 7. Start the VM again (now running on a COW overlay on top of the base image)
streamer-vm start ubuntu-live

# 8. Reset overlay disk back to clean state from last commit (discard all experimental changes)
streamer-vm reset ubuntu-live

# Tips:
# - Use 'streamer-vm update <name> --iso <path>' or '--remove-iso' to manage ISO media anytime.
# - Use '--lang=en' or '--lang=pt' (or set STREAMER_LANG) to switch interface language.
```

## License

MIT License. See [LICENSE](LICENSE) for details.

# streamer-vm — Design Document

**Status:** Design approved
**Date:** 2026-07-29
**Host:** Debian 13 (Trixie), QEMU 10.0.11, AMD RX 6800 XT, Wayland, OBS Studio 32.2.1 (Flatpak)

---

## 1. Visão Geral

Script em Go que gerencia VMs QEMU+KVM para gravar live streams. A VM funciona como uma janela no desktop do host, e o OBS Studio no host é usado para gravar e fazer streaming. A VM usa aceleração gráfica VirGL via GPU AMD RX 6800 XT do host.

### Princípios

- Máxima performance gráfica (VirGL + GPU real)
- OBS no host para gravação/streaming (sem instalar OBS na VM)
- Zero dependências externas no binário final (Go com cobra)
- Disco persistente com overlay qcow2
- Protocolo Spice para display (latência baixa, GL, clipboard bidirecional)

---

## 2. Arquitetura

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Host (Debian 13 / Wayland)                            │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │              QEMU + KVM                                          │   │
│  │                                                                  │   │
│  │  CPU: host-passthrough (4 vCPUs)                                 │   │
│  │  RAM: 8 GB                                                       │   │
│  │  GPU: virtio-gpu-gl-device (virgl=on) → rendernode=/dev/dri/card0│   │
│  │       (VirGL → libvirglrenderer → radeonsi → RX 6800 XT hardware)│   │
│  │  Display: spice-app,gl=on (porta 5900)                           │   │
│  │  Storage: virtio-scsi + qcow2 overlay                            │   │
│  │  Network: virtio-net + user mode (slirp)                         │   │
│  │  Audio: pipewire audiodev → intel-hda                            │   │
│  │  UEFI: OVMF_CODE_4M.fd + OVMF_VARS_4M.fd                         │   │
│  │  Clipboard: vdagent (virtserialport)                             │   │
│  └──────────────┬───────────────────────────────────────────────────┘   │
│                 │ Spice protocol (TCP 127.0.0.1:5900)                    │
│  ┌──────────────▼───────────────────────────────────────────────────┐   │
│  │  spice-client-gtk (janela GTK no desktop Wayland)                 │   │
│  │  ← VM aparece como janela nativa do sistema                       │   │
│  └──────────────┬───────────────────────────────────────────────────┘   │
│                 │                                                        │
│  ┌──────────────▼───────────────────────────────────────────────────┐   │
│  │  OBS Studio (Flatpak)                                             │   │
│  │  Source: "Window Capture (Wayland)" via xdg-desktop-portal        │   │
│  │  → Gravação local (x264/x265) + streaming (RTMP/SRT)              │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Aceleração Gráfica — VirGL

```
VM: virtio-gpu-gl-device (virgl=on)
  │
  ▼
Mesa 3D no guest traduz OpenGL → VirGL
  │
  ▼
QEMU host: spice-app com GL habilitado
  │
  ▼
spice-server: egl-headless com rendernode=/dev/dri/card0
  │
  ▼
libvirglrenderer no host → DRM/KMS → radeonsi driver → RX 6800 XT (hardware)
  │
  ▼
spice-client-gtk no host recebe o framebuffer e exibe na janela
```

**Prerequisites no guest:**
- Kernel Linux ≥ 5.9 (virtio-gpu driver nativo)
- Mesa ≥ 20.0 (suporte virgl)
- Display server com suporte EGL/GLX (Xorg ou Wayland)
- Debian 13 / Ubuntu 24.04 funcionam out-of-the-box

---

## 4. Protocolo de Exibição — Spice vs VNC vs SPICE GL

### Por que Spice

| Critério | Spice | VNC |
|---|---|---|
| Latência | Baixa (protocolo otimizado) | Alta |
| Qualidade | lossless | Compressão com perdas |
| Clipboard | Bidirecional (vdagent) | Não |
| USB redirection | Sim | Não |
| Multi-monitor | Nativo | Limitado |
| Áudio | Codec de alta qualidade | Sem áudio nativo |
| VirGL/GL | Suporte via GL scanout | Não |

### Por que NÃO cliente SPICE em Go puro

A biblioteca `Shells-com/spice` (v0.0.6, MIT, 22 stars) implementa o protocolo SPICE em Go puro, mas:

1. **GL scanout NÃO implementado** — constantes do protocolo existem (`SPICE_MSG_DISPLAY_GL_SCANOUT_UNIX = 320`), mas não há handler para processar dma-buf textures.
2. **TCP sem fd-passing** — a biblioteca conecta via `net.Dial` (TCP). O protocolo GL do SPICE requer Unix socket com ancillary data (file descriptor passing) para transmitir dma-bufs. Sem fd-passing, não há como importar texturas GL.
3. **spice-client-gtk é a única via prática** para VirGL no host.

### Workaround QEMU 2025 (não aplicável)

Patch merged no QEMU blita textura GL para buffer linear para clientes remotos, mas requer cliente compatível — `Shells-com/spice` não implementa isso.

**Decisão:** Usar `spice-client-gtk` como cliente nativo. Script Go controla QEMU; usuário abre janela manualmente.

---

## 5. Captura OBS — Pipeline

```
spice-client-gtk (janela GTK no Wayland)
  │
  ▼
xdg-desktop-portal (rodando no host, serviço ativo)
  │  → Session portal + Screen cast portal
  ▼
OBS Studio (Flatpak, com suporte Wayland via xdg-desktop-portal)
  │  → Fonte "Window Capture (Wayland)"
  ▼
Selecionar janela "spice-client-gtk"
  │
  ├── Gravação → arquivo local (x264/x265, CRF 18)
  └── Streaming → RTMP/SRT para Twitch, YouTube, etc.
```

**Passos para o usuário:**
1. Abrir OBS Studio
2. Sources → `+` → "Window Capture (Wayland)"
3. Selecionar janela "spice-client-gtk"
4. Ajustar crop/position se necessário
5. Configurar cena, gravação, streaming normalmente

**Áudio da VM:** Sai pelo pipewire do host. OBS pode capturar via:
- "PipeWire Capture" → selecionar sink da VM
- Ou capturar a janela completa (inclui áudio via portal)

---

## 6. Estrutura de Diretórios

```
$STREAMER_HOME/                          # default: ~/.local/share/streamer-vm/
├── configs/
│   └── <name>/
│       ├── vm.json              # metadados da VM (cpus, ram, disk, spice_port)
│       ├── ovf-vars.fd          # cópia writeable do OVMF_VARS_4M.fd
│       └── spice-certs/         # certificados TLS do Spice (opcional)
├── disks/
│   ├── <name>.qcow2             # disco base (snapshot overlay)
│   └── <name>-overlay.qcow2    # overlay writeable (cow)
├── iso/
│   └── <os>.iso                 # ISO do SO convidado
├── state/
│   ├── <name>.pid               # PID do processo QEMU
│   └── <name>.spice-port        # porta Spice alocada
└── logs/
    └── <name>.log               # log do QEMU
```

---

## 7. Comando QEMU

```bash
qemu-system-x86_64 \
  -enable-kvm \
  -machine q35,accel=kvm \
  -cpu host,kvm=on,vendor=GenuineIntel \
  -smp ${CPUS},sockets=1,cores=${CPUS},threads=1 \
  -m ${MEMORY}G \
  \
  -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE_4M.fd \
  -drive if=pflash,format=raw,file=$CFG_DIR/ovf-vars.fd \
  \
  -device virtio-gpu-gl-device,virgl=on,gl=on \
  -display spice-app,gl=on \
  -spice port=${SPICE_PORT},disable-ticketing=on,addr=127.0.0.1 \
  \
  -device virtio-scsi-pci \
  -device scsi-hd,drive=disk \
  -drive id=disk,file=$DISK_DIR/${NAME}-overlay.qcow2,format=qcow2,cache=none,aio=native \
  \
  -device virtio-net-pci,netdev=net \
  -netdev user,id=net \
  \
  -audiodev pipewire,id=audio \
  -device intel-hda \
  -device hda-output,audiodev=audio \
  \
  -chardev spicevmc,id=vdagent,name=vdagent \
  -device virtserialport,chardev=vdagent,name=com.redhat.spice.0 \
  \
  -no-reboot \
  -daemonize \
  -logfile $LOG_DIR/${NAME}.log
```

**Pontos chave:**
- `-display spice-app,gl=on` — janela Spice com aceleração GL
- `-spice port=...,disable-ticketing=on` — sem TLS para local (mais rápido)
- `vdagent` — clipboard bidirecional entre host e guest
- `spice-app` — abre automaticamente o cliente Spice no host (se disponível)
- `cache=none,aio=native` — performance máxima de disco

---

## 8. Interface CLI do Script (Go)

```bash
streamer-vm init                    # Cria estrutura de dirs, verifica prereqs
streamer-vm create <name> [flags]   # Cria disco base + config (sem overlay inicial)
streamer-vm start <name> [flags]    # Inicia QEMU, espera Spice pronto
streamer-vm stop <name>             # Shutdown gracioso (ACPI) + cleanup
streamer-vm update <name> [flags]   # Atualiza config (anexa ou remove ISO)
streamer-vm commit <name> [flags]   # Consolida base image e ativa/renova overlay
streamer-vm reset <name>            # Recria overlay limpo (volta ao último commit)
streamer-vm status [name]           # Estado de VMs e modo de disco
streamer-vm list                    # Lista todas
streamer-vm delete <name>           # Remove VM
streamer-vm spice-url <name>        # Imprime spice://host:port para copiar
```

**Flags de `create`:**
```
  -cpus N        vCPUs (default: 4)
  -memory N      RAM em GB (default: 8)
  -disk N        Disco em GB (default: 50)
  -iso <path>    ISO do SO convidado
  -spice-port N  Porta Spice (default: auto 5900+)
  -name <name>   Nome da VM
```

---

## 9. Fluxo de Execução

### `init`
1. Cria `$STREAMER_HOME/{configs,disks,iso,state,logs}`
2. Verifica prereqs: KVM, `/dev/dri/card0`, QEMU, OVMF, pipewire
3. Copia `OVMF_VARS_4M.fd` para template base

### `create <name>`
1. Lê flags, gera `vm.json`
2. Aloca porta Spice livre (poll 5900-5999)
3. Cria disco base: `qemu-img create -f qcow2 disks/<name>.qcow2 <size>G` (sem overlay inicial)
4. Copia `OVMF_VARS.fd` para `configs/<name>/ovf-vars.fd`
5. Salva porta em `state/<name>.spice-port`

### `start <name>`
1. Lê `vm.json`
2. Constrói args do QEMU (usa overlay se já existir, senão usa o disco base diretamente)
3. Lança QEMU em daemon mode
4. Salva PID em `state/<name>.pid`
5. Poll até Spice socket estar escutando
6. Imprime: Spice URL + instrução para abrir no cliente Spice

### `stop <name>`
1. Envia ACPI shutdown via QEMU monitor
2. Se não responder em 30s → `kill -SIGTERM` no PID
3. Aguarda processo terminar (poll com timeout)
4. Limpa PID files

### `update <name>`
1. Permite anexar ISO (`-iso <path>`) ou remover ISO (`-remove-iso`)
2. Salva a nova configuração no `vm.json`

### `commit <name>`
1. Para a VM se estiver rodando
2. Se overlay já existir: executa `qemu-img commit disks/<name>-overlay.qcow2` e recria overlay limpo
3. Se overlay não existir (primeiro commit): cria o primeiro `disks/<name>-overlay.qcow2` sobre o disco base
4. Faz backup de `configs/<name>/ovf-vars.fd` para `ovf-vars.base.fd`
5. Se `-remove-iso` for passado, remove a ISO da configuração

### `reset <name>`
1. Para a VM se estiver rodando
2. Valida se o overlay existe (retorna erro se a VM ainda não foi commitada)
3. Deleta o overlay atual e recria overlay limpo sobre a imagem base consolidada
4. Restaura `ovf-vars.base.fd` para `ovf-vars.fd` (se existir)

### `status [name]`
1. Se name fornecido: lê PID, verifica processo, exibe hardware, ISO e modo de disco (Overlay vs Direct Base)
2. Se não: lista todas VMs com status resumido

---

## 10. Configuração do Guest (VM Convidada)

O script **não** instala o SO no guest. Fluxo:

1. Usuário baixa ISO (Ubuntu Desktop, Debian com GNOME, etc.)
2. `streamer-vm create ubuntu-live -iso ~/Downloads/ubuntu-24.04.iso`
3. `streamer-vm start ubuntu-live` — boot da ISO via janela spice (gravação direta no disco base)
4. Instala SO na janela spice-client-gtk
5. `streamer-vm stop ubuntu-live` — desliga a VM
6. `streamer-vm commit ubuntu-live --remove-iso` — consolida disco base, gera overlay e remove live ISO
7. `streamer-vm start ubuntu-live` — boot do SO instalado em overlay COW

**Guest recomendado:** Ubuntu Desktop ou Debian com GNOME (virgl funciona out-of-the-box).

**Drivers no guest:**
- `virtio` (storage, network) — nativo no kernel Linux ≥ 5.9
- `virgl` (GPU) — habilitado automaticamente no display server (Mesa)

---

## 11. Go — Estrutura do Projeto

```
streamer-vm/
├── main.go              # CLI entry point (standard lib)
├── cmd/
│   ├── init.go          # "init" command
│   ├── create.go        # "create" command
│   ├── start.go         # "start" command
│   ├── stop.go          # "stop" command
│   ├── update.go        # "update" command (attach/remove ISO)
│   ├── commit.go        # "commit" command (base image consolidation)
│   ├── reset.go         # "reset" command
│   ├── status.go        # "status" command
│   ├── list.go          # "list" command
│   ├── delete.go        # "delete" command
│   └── spice_url.go     # "spice-url" command
├── vm/
│   ├── config.go        # vm.json parsing/serialization
│   ├── qemu.go          # QEMU process management
│   ├── disk.go          # qemu-img operations
│   └── spice.go         # Spice port allocation, health check
├── guest/
│   └── prereqs.go       # KVM, GPU, OVMF checks
└── internal/
    └── logger.go        # logging utility
```

**Dependências externas:**
- `github.com/spf13/cobra` — CLI framework
- `github.com/spf13/pflag` — flag parsing (compatível com flag standard)

---

## 12. vm.json — Formato de Configuração

```json
{
  "name": "ubuntu-live",
  "cpus": 4,
  "memory_gb": 8,
  "disk_gb": 50,
  "spice_port": 5900,
  "iso_path": "",
  "created_at": "2026-07-29T11:00:00Z",
  "state": "stopped"
}
```

---

## 13. Considerações de Qualidade de Vídeo (OBS)

| Parâmetro | Valor | Razão |
|---|---|---|
| Codec | libx264 ou libx265 | Amplamente compatível |
| Preset | fast ou medium | Equilíbrio CPU/qualidade |
| CRF | 18 | Alta qualidade (lossless perceptual) |
| FPS | 30 ou 60 | Configurável no OBS |
| Formato | .mp4 ou .mkv | .mkv mais seguro se gravada abruptamente |
| Audio | aac 192k | Padrão para streaming |

---

## 14. Segurança

| Aspecto | Decisão |
|---|---|
| Spice TLS | Desabilitado (`disable-ticketing=on`) — VM roda local |
| Network | User mode (slirp) — isolado, sem acesso à rede local |
| Disco | Overlay cow sobre base read-only — reset fácil com `reset` |
| Permissões | `/dev/kvm` (root:kvm), `/dev/dri` (root:render) — usuário nos grupos ✓ |

---

## 15. Riscos e Mitigações

| Risco | Mitigação |
|---|---|
| `spice-app` não abre cliente automaticamente | Script imprime URL `spice://` para usuário abrir manualmente |
| spice-client-gtk não instalado | `init` verifica e instrui: `apt install spice-client-gtk` |
| Virgl não funciona no guest | Guest precisa Mesa ≥ 20.0 + kernel ≥ 5.9 (Debian 13 tem tudo) |
| OBS Flatpak não captura janela Wayland | xdg-desktop-portal-gtk já rodando; se falhar, documentar workaround |
| QEMU crash durante streaming | Overlay cow protege disco base; usuário recria com `reset` |
| Porta Spice conflitante | Auto-aloca porta livre (poll 5900-5999) |
| Disco cheio | Check de espaço antes de `create`; alertas |

---

## 16. Fluxo Completo de Uso

```bash
# 1. Setup inicial
streamer-vm init

# 2. Criar VM Ubuntu Desktop (modo setup inicial, gravando direto no disco base)
streamer-vm create ubuntu-live -cpus 6 -memory 12 -disk 80 -iso ~/Downloads/ubuntu-24.04.iso

# 3. Iniciar — abre janela spice para instalar o SO
streamer-vm start ubuntu-live

# 4. Instalar Ubuntu na janela spice-client-gtk

# 5. Após instalação, parar e consolidar como base image
streamer-vm stop ubuntu-live
streamer-vm commit ubuntu-live --remove-iso

# 6. Iniciar novamente (agora em overlay COW sobre a base instalada)
streamer-vm start ubuntu-live

# 7. No OBS: Window Capture (Wayland) → janela spice-client-gtk
#    → Gravar ou streamar normalmente

# 8. Reset do disco (volta ao estado pós-instalação ou do último commit)
streamer-vm reset ubuntu-live
```

---

## 17. Próximos Passos

1. Implementar estrutura Go com cobra
2. Comandos `init` e `create`
3. Comando `start` com gestão de processos QEMU
4. Comando `stop` com ACPI shutdown
5. Comandos `status`, `list`, `delete`, `reset`
6. Testes end-to-end com VM Ubuntu + OBS

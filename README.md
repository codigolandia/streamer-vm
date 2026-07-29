# streamer-vm

`streamer-vm` é uma ferramenta CLI em Pure Go projetada para gerenciar máquinas virtuais QEMU+KVM com aceleração gráfica VirGL (OpenGL pass-through), exibição SPICE e integração nativa para gravação e transmissão via OBS Studio no Linux.

## Recursos

- **VirGL GPU Acceleration**: Aceleração OpenGL nativa via GPU do host (ex: AMD Radeon / Intel / NVIDIA via DRM scanout).
- **Zero Dependências Externas**: Binário compilado em Pure Go usando apenas a biblioteca padrão.
- **Isolamento via Qcow2 Overlay**: Mantém o disco base intacto e utiliza overlays COW graváveis para fácil reset.
- **Protocolo SPICE**: Suporte a áudio Pipewire, clipboard bidirecional (vdagent) e baixa latência.
- **Suporte a Linux**: Otimizado e voltado exclusivamente para sistemas Linux com Wayland/X11 e QEMU/KVM.

## Pré-requisitos

No host Linux:
- `QEMU` (com suporte a KVM e VirGL)
- `qemu-img`
- Firmware UEFI `OVMF` (`OVMF_CODE_4M.fd` e `OVMF_VARS_4M.fd`)
- Opcional: `spice-client-gtk` (`spicy` ou `virt-viewer`) para exibição direta da VM em janela desktop

## Instalação

Baixe o binário pré-compilado para Linux (amd64 / arm64) das [Releases](https://github.com/codigolandia/streamer-vm/releases) ou compile a partir do código fonte:

```bash
go build -o streamer-vm .
```

## Uso Rápido

```bash
# 1. Inicializar diretórios e verificar pré-requisitos do sistema
streamer-vm init

# 2. Criar uma nova VM
streamer-vm create ubuntu-live -cpus 4 -memory 8 -disk 50 -iso ~/Downloads/ubuntu-24.04.iso

# 3. Iniciar a VM
streamer-vm start ubuntu-live

# 4. Exibir URL de conexão SPICE
streamer-vm spice-url ubuntu-live

# 5. Consultar status
streamer-vm status ubuntu-live

# 6. Desligar a VM graciosamente (ACPI)
streamer-vm stop ubuntu-live

# 7. Resetar disco overlay para o estado limpo inicial
streamer-vm reset ubuntu-live
```

## Licença

MIT License. Veja [LICENSE](LICENSE) para mais detalhes.

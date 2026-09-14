# streamer-vm

*[English](README.md) | [Português](README.pt-br.md)*

`streamer-vm` é uma ferramenta CLI em Pure Go projetada para gerenciar máquinas virtuais QEMU+KVM com aceleração gráfica VirGL (OpenGL pass-through), exibição SPICE e integração nativa para gravação e transmissão via OBS Studio no Linux.

## Recursos

- **Aceleração Gráfica VirGL (GPU)**: Aceleração OpenGL nativa via GPU do host (ex: AMD Radeon / Intel / NVIDIA via DRM scanout).
- **Zero Dependências Externas**: Binário compilado em Pure Go usando apenas a biblioteca padrão.
- **Suporte a Internacionalização (i18n)**: Suporte nativo aos idiomas Inglês (`en`) e Português (`pt`), com detecção automática do sistema ou via flag `--lang`.
- **Isolamento via Qcow2 Overlay**: Mantém o disco base intacto e utiliza overlays COW graváveis para fácil restauração (`reset`).
- **Protocolo SPICE**: Suporte a áudio Pipewire, clipboard bidirecional (vdagent), redirecionamento de dispositivos USB (XHCI) e baixa latência.
- **Suporte a Linux**: Otimizado e voltado exclusivamente para sistemas Linux com Wayland/X11 e QEMU/KVM.

## Pré-requisitos

No host Linux:
- `QEMU` (com suporte a KVM e VirGL: `qemu-system-x86_64`)
- `qemu-img`
- Firmware UEFI `OVMF` (`OVMF_CODE_4M.fd` e `OVMF_VARS_4M.fd`)
- Opcional: `spice-client-gtk` (`spicy` ou `virt-viewer`) para exibição direta da VM em janela desktop

## Instalação

### Instalação Rápida / Atualização

Instale ou atualize para a versão mais recente em `~/.local/bin` ou `~/bin` (o que estiver primeiro no seu `$PATH`):

```bash
curl -fsSL https://github.com/codigolandia/streamer-vm/releases/latest/download/install.sh | bash
```

Para fazer rollback para a versão anterior (restaurando `streamer-vm.backup`):

```bash
curl -fsSL https://github.com/codigolandia/streamer-vm/releases/latest/download/install.sh | bash -s -- --rollback
```

### Download Manual ou Compilação pelo Código-Fonte

Baixe o binário pré-compilado para Linux (amd64 / arm64) das [Releases](https://github.com/codigolandia/streamer-vm/releases) ou compile a partir do código fonte:

```bash
go build -o streamer-vm .
```

## Uso Rápido

```bash
# 1. Inicializar diretórios e verificar pré-requisitos do sistema
streamer-vm init

# 2. Criar uma nova VM com imagem ISO (modo setup inicial, direto no disco base)
streamer-vm create ubuntu-live -cpus 4 -memory 8 -disk 50 -iso ~/Downloads/ubuntu-24.04.iso

# 3. Iniciar a VM e abrir automaticamente a janela do cliente SPICE (spicy)
streamer-vm start ubuntu-live --gui

# 4. Exibir URL de conexão SPICE (caso use spicy ou virt-viewer separadamente)
streamer-vm spice-url ubuntu-live

# 5. Após concluir a instalação do SO, desligar a VM (ou desligar pelo SO)
streamer-vm stop ubuntu-live

# 6. Gravar o estado instalado como base image e remover a ISO do instalador
streamer-vm commit ubuntu-live --remove-iso

# 7. Iniciar a VM novamente (agora rodando em overlay COW sobre a base instalada)
streamer-vm start ubuntu-live

# 8. Resetar disco overlay para o estado limpo do último commit (descarta alterações)
streamer-vm reset ubuntu-live

# Dicas:
# - Use 'streamer-vm update <name> --iso <path>' ou '--remove-iso' para alterar mídias ISO.
# - Use '--lang=pt' ou '--lang=en' (ou defina STREAMER_LANG) para alternar o idioma da interface.
```

## Licença

MIT License. Veja [LICENSE](LICENSE) para mais detalhes.

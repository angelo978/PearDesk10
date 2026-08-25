# PearDesk-Tor (Orion)

Versione separata di PearDesk con **Tor Orion come unico trasporto**.

## Trasporto

- Host: avvia il processo `tor`, crea un servizio Onion v3 e salva l'indirizzo
  completo generato da Tor.
- Client: accetta esclusivamente un indirizzo completo `*.onion` e si collega
  esclusivamente al proxy SOCKS5 locale di Tor (`127.0.0.1:9050`).
- Non sono presenti P2P, Hyperswarm, WAN diretto, relay, Cloudflared o fallback.
- L'autenticazione password viene eseguita sul canale Tor prima del video.

## Identità

L'ID non è abbreviato e non viene inventato localmente. La prima avvio senza
un ID salvato avvia Tor Orion; quando Tor genera il suo hostname v3, questo
hostname completo viene scritto nel campo ID e nella configurazione.

Gli ID vecchi basati su IP, porte o codici brevi vengono rifiutati.

## Canali Tor

Il protocollo definisce canali separati per:

- monitor/video
- mouse
- tastiera
- trasferimento file
- clipboard testo
- clipboard file

Ogni canale utilizza un tipo di frame dedicato e non può passare a un altro
trasporto. Il testo clipboard è bidirezionale; il trasferimento file usa
chunking e ricostruzione.

## Autostart

Non viene installato alcun servizio systemd, Windows Service o LaunchDaemon.

- Linux: `~/.config/autostart/peardesk.desktop`
- Windows: `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`

## Requisiti Linux

```bash
sudo apt install tor libx11-dev libxrandr-dev libxcursor-dev libxi-dev \
  libxinerama-dev libgl1-mesa-dev libgles2-mesa-dev \
  libfontconfig1-dev libfreetype6-dev libxtst-dev \
  libavcodec-dev libavutil-dev libswscale-dev libx264-dev \
  pkg-config gcc wget curl ca-certificates
```

Build AppImage:

```bash
sudo bash build-appimage.sh
```

Il demone `tor` deve essere disponibile nel `PATH`. La versione Orion avvia
un proprio processo Tor per l'host; il client richiede un proxy SOCKS5 Tor
locale sulla porta `9050`.

## Verifica sorgenti

```bash
GOTOOLCHAIN=local go build -tags headless ./pkg/...
```

La GUI e i nomi dei controlli restano quelli della versione PearDesk2.
PearDesk10

PearDesk10 è un desktop remoto leggero che instrada tutto il traffico esclusivamente su Tor Orion.
Non usa WAN diretta, relay, Hyperswarm, WebRTC, Cloudflared o altri trasporti: solo Onion v3.

Pensato per chi vuole un controllo remoto semplice, isolato e senza dipendenze cloud.
✔ Cosa FA davvero (oggi)

    Connessione solo Tor Onion v3

    Autenticazione password sul canale Tor

    Streaming video del desktop

    Controllo remoto di mouse e tastiera

    Clipboard testo bidirezionale

    Trasferimento file con chunking

    Avvio automatico opzionale (Linux/Windows)

    Build AppImage e Windows già pronti

    Nessun servizio systemd o Windows Service

    Nessun fallback WAN

    Nessuna dipendenza cloud

❗ Cosa NON fa ancora (senza raccontare balle)

    Non ha 6 onion endpoint separati

    Non ha 6 connessioni Tor fisiche indipendenti

    Non ha riconnessione automatica ogni 60 secondi

    Non ha canali fisici separati (solo logici)

    Non ha recovery automatico del circuito Tor

    Non è multi‑utente o multi‑sessione

Oggi PearDesk10 usa un’unica connessione Tor con multiplexing dei frame.
🔐 Trasporto Tor
Host

    Avvia Tor Orion locale

    Genera un servizio Onion v3

    Scrive l’hostname completo nella configurazione

    Accetta solo connessioni Tor

Client

    Richiede esclusivamente un indirizzo .onion

    Usa il proxy SOCKS5 locale: 127.0.0.1:9050

    Non tenta altri trasporti

🆔 Identità

    L’ID è l’hostname Onion v3 generato da Tor

    Non vengono accettati ID abbreviati, IP o codici brevi

    La prima esecuzione senza ID avvia Tor e genera l’identità

📡 Canali logici (stato attuale)

I frame trasmessi sulla connessione Tor includono:

    video

    mouse

    tastiera

    clipboard testo

    
Ogni frame ha un tipo dedicato, ma viaggia sullo stesso socket Tor.
🚀 Autostart

    Linux: ~/.config/autostart/peardesk.desktop

    Windows: HKCU\Software\Microsoft\Windows\CurrentVersion\Run

Nessun servizio systemd o Windows Service.
🐧 Requisiti Linux
Codice

sudo apt install tor libx11-dev libxrandr-dev libxcursor-dev libxi-dev \
  libxinerama-dev libgl1-mesa-dev libgles2-mesa-dev \
  libfontconfig1-dev libfreetype6-dev libxtst-dev \
  libavcodec-dev libavutil-dev libswscale-dev libx264-dev \
  pkg-config gcc wget curl ca-certificates

🏗 Build AppImage
Codice


Richiede Go e rsrc per incorporare l’icona.
🔍 Verifica sorgenti
Codice

GOTOOLCHAIN=local go build -tags headless ./pkg/...

La GUI utilizza Fyne.

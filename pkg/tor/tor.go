package torboot

import (
    "archive/zip"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

// URL ufficiale Tor Expert Bundle (solo binari, niente browser)
const torZipURL = "https://www.torproject.org/dist/torbrowser/tor-win32-0.4.8.10.zip"

// Percorso dove installiamo Tor automaticamente
func torInstallPath() string {
    appdata := os.Getenv("APPDATA")
    return filepath.Join(appdata, "PearDesk10", "tor")
}

func torExePath() string {
    return filepath.Join(torInstallPath(), "tor.exe")
}

// Controlla se Tor è già in esecuzione
func isTorRunning() bool {
    out, _ := exec.Command("tasklist").Output()
    return strings.Contains(string(out), "tor.exe")
}

// Scarica Tor Expert Bundle
func downloadTorZip(dest string) error {
    fmt.Println("Scarico Tor Expert Bundle...")

    resp, err := http.Get(torZipURL)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    f, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer f.Close()

    _, err = io.Copy(f, resp.Body)
    return err
}

// Estrae solo tor.exe dal bundle
func extractTor(zipPath string, destDir string) error {
    fmt.Println("Estraggo tor.exe...")

    r, err := zip.OpenReader(zipPath)
    if err != nil {
        return err
    }
    defer r.Close()

    os.MkdirAll(destDir, 0755)

    for _, file := range r.File {
        // cerchiamo solo tor.exe
        if filepath.Base(file.Name) == "tor.exe" {
            rc, err := file.Open()
            if err != nil {
                return err
            }
            defer rc.Close()

            outPath := filepath.Join(destDir, "tor.exe")
            outFile, err := os.Create(outPath)
            if err != nil {
                return err
            }
            defer outFile.Close()

            _, err = io.Copy(outFile, rc)
            return err
        }
    }

    return errors.New("tor.exe non trovato nello zip")
}

// Avvia Tor
func startTor() error {
    tor := torExePath()
    if _, err := os.Stat(tor); os.IsNotExist(err) {
        return errors.New("tor.exe non trovato")
    }

    fmt.Println("Avvio Tor...")
    cmd := exec.Command(tor)
    cmd.Start()

    // aspetta che apra il SOCKS5
    time.Sleep(2 * time.Second)
    return nil
}

// Funzione principale da chiamare all'avvio di PearDesk10
func EnsureTor() error {

    // 1) Tor già in esecuzione?
    if isTorRunning() {
        fmt.Println("Tor è già in esecuzione.")
        return nil
    }

    // 2) Tor installato localmente?
    if _, err := os.Stat(torExePath()); err == nil {
        return startTor()
    }

    // 3) Scarica Tor
    os.MkdirAll(torInstallPath(), 0755)
    zipPath := filepath.Join(torInstallPath(), "tor.zip")

    if err := downloadTorZip(zipPath); err != nil {
        return err
    }

    // 4) Estrai tor.exe
    if err := extractTor(zipPath, torInstallPath()); err != nil {
        return err
    }

    // 5) Avvia Tor
    return startTor()
}

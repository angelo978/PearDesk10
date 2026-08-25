package main

import (
    _ "embed"
    "log"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"

    "github.com/peardesk/peardesk/pkg/config"
    "github.com/peardesk/peardesk/pkg/ui"
    "github.com/peardesk/peardesk/pkg/id"
)

//go:embed icon.png
var iconData []byte

var Version = "2.0.0"

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("config load error: %v", err)
    }

    // ⭐ CARICA ID FISSO SE ESISTE
    if savedID, err := id.Load(); err == nil {
        cfg.HostID = savedID
    }

    a := app.NewWithID("com.peardesk.app")

    // ICONA EMBEDDED (Linux + Windows)
    if res := loadIcon(); res != nil {
        a.SetIcon(res)
    }

    // ⭐ PASSA A UI LA FUNZIONE PER SALVARE L’ID QUANDO PREMI “RIGENERA”
    ui.SetIDSaveFunc(func(newID string) {
        id.Save(newID)
        cfg.HostID = newID
    })

    mw := ui.NewMainWindow(a, cfg)
    mw.Show()
}

func loadIcon() fyne.Resource {
    if len(iconData) > 0 {
        return &fyne.StaticResource{
            StaticName:    "icon.png",
            StaticContent: iconData,
        }
    }
    return nil
}

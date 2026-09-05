package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/benzjeremy/spotify-screensaver/server"
	"github.com/benzjeremy/spotify-screensaver/spotify"
	"github.com/benzjeremy/spotify-screensaver/store"
)

const Version = "v1.1"

//go:embed assets/*
var embeddedAssets embed.FS

func main() {
	portFlag := flag.Int("port", 43210, "Port für internen HTTP-Server (nur 127.0.0.1)")
	fullscreenFlag := flag.Bool("fullscreen", false, "Screensaver direkt im Vollbildmodus starten")
	browserFlag := flag.Bool("browser", false, "Standard-Browser anstelle des WebKitGTK-Fensters verwenden")
	versionFlag := flag.Bool("version", false, "Zeigt die Programmversion an")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Spotify Screensaver %s by Jeremy Benz\n", Version)
		return
	}

	log.Printf("🌌 Spotify Screensaver %s startet...\n", Version)

	// 1. Initialisiere sicheren, verschlüsselten Secrets-Speicher (AES-256-GCM, PBKDF2 100k)
	secStore, err := store.NewSecureStore()
	if err != nil {
		log.Fatalf("[Sicherheit] Fehler beim Initialisieren des SecureStore: %v\n", err)
	}

	// 2. Initialisiere Spotify Controller
	ctrl := spotify.NewController(secStore)

	// 3. Bereite eingebettete Assets vor
	assetsSubFS, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		log.Fatalf("[Assets] Fehler beim Laden der eingebetteten Assets: %v\n", err)
	}

	// 4. Starte gesicherten lokalen HTTP Server (nur 127.0.0.1, Anti-CSRF, Anti-DNS-Rebinding)
	srv := server.NewServer(*portFlag, ctrl, secStore, assetsSubFS)
	appURL, err := srv.Start()
	if err != nil {
		log.Fatalf("[Server] Fehler beim Starten des Webservers: %v\n", err)
	}
	log.Printf("[Server] Interner Webserver aktiv unter: %s\n", appURL)

	// 5. Signal-Handling für sauberes Beenden
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\n[Shutdown] Beende Spotify Screensaver sauber...")
		srv.Stop()
		os.Exit(0)
	}()

	// 6. Starte Benutzeroberfläche
	if *browserFlag {
		log.Printf("[GUI] Starte im Browser-Modus: %s\n", appURL)
		LaunchGUI("Spotify Screensaver", appURL, 1200, 800, false)
	} else {
		LaunchGUI("Spotify Screensaver", appURL, 1200, 800, *fullscreenFlag)
	}
}

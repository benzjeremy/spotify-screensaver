//go:build windows

package main

import (
	"log"
	"os/exec"
)

func LaunchGUI(title, url string, width, height int, fullscreen bool) {
	log.Println("[GUI] Starte Spotify Screensaver im App-Modus unter Windows...")
	commands := [][]string{
		{"msedge.exe", "--app=" + url},
		{"chrome.exe", "--app=" + url},
		{"cmd.exe", "/c", "start", url},
	}

	for _, cmdArgs := range commands {
		if path, err := exec.LookPath(cmdArgs[0]); err == nil {
			cmd := exec.Command(path, cmdArgs[1:]...)
			if err := cmd.Start(); err == nil {
				log.Printf("[Browser] Geöffnet mit %s (%s)\n", cmdArgs[0], url)
				return
			}
		}
	}

	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

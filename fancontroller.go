package main

import (
	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
)

const (
	APP_VERSION = "0.1"

	DEBUG_INFO = false
)

func main() {
	var appConfig AppConfig
	readAppConfig(&appConfig)

	serial := NewSerial(&appConfig)
	appGUI := NewAppGUI(serial, &appConfig)

	ui.Main(appGUI.SetupUI)
}

package main

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

const (
	MAX_TEMP = 150.0
	MAX_RPM  = 3000.0

	APP_CONFIG = "./fancontroller.toml"
)

type AppConfig struct {
	Ports              []string
	MaxRPM             int
	MaxTemp            int
	AutoStartInSystray bool
}

func readAppConfig(appConfig *AppConfig) {
	appConfig.MaxRPM = MAX_RPM
	appConfig.MaxTemp = MAX_TEMP
	if _, err := os.Stat(APP_CONFIG); !os.IsNotExist(err) {
		if _, err := toml.DecodeFile(APP_CONFIG, appConfig); err != nil {
			log.Printf("err=%v", err)
		}
	}
}

func writeAppConfig(appConfig *AppConfig) {
	var f *os.File
	var err error
	if f, err = os.Create(APP_CONFIG); err != nil {
		log.Printf("err=%v", err)
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	if err = enc.Encode(appConfig); err != nil {
		log.Printf("err=%v", err)
	}
}

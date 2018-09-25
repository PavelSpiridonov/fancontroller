package main

import (
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

type Serial struct {
	port *serial.Port

	status *Status
	config *Config

	appGUI    *AppGUI
	appConfig *AppConfig

	lockPort     sync.Mutex
	stopReadPort bool
}

func NewSerial(appConfig *AppConfig) *Serial {
	return &Serial{
		appConfig: appConfig,
	}
}

func (ser *Serial) GetConfig() Config {
	if ser.config != nil {
		return *ser.config
	}
	return Config{}
}

func (ser *Serial) GetStatus() Status {
	if ser.status != nil {
		return *ser.status
	}
	return Status{}
}

func (ser *Serial) StopRead() {
	ser.stopReadPort = true

	ser.config = nil
	ser.status = nil

	ser.lockPort.Lock()
	if ser.port != nil {
		ser.port.Close()
	}
	ser.port = nil
	ser.lockPort.Unlock()
}

func (ser *Serial) ApplyConfig(config *Config) bool {
	configStr := configToStr(config) + "\r\n"
	if DEBUG_INFO {
		log.Printf("configStr=%s", configStr)
	}
	ser.lockPort.Lock()
	_, err := ser.port.Write([]byte(configStr))
	ser.lockPort.Unlock()
	if err != nil {
		if DEBUG_INFO {
			log.Printf("err=%v", err)
		}
		ser.appGUI.ShowError(err, true)
		return false
	}
	return true
}

func (ser *Serial) readPort() {
	ser.stopReadPort = false
	for {
		if ser.stopReadPort {
			return
		}

		buf := make([]byte, 256)
		ser.lockPort.Lock()
		n, err := ser.port.Read(buf)
		ser.lockPort.Unlock()
		if err != nil && err != io.EOF {
			if DEBUG_INFO {
				log.Printf("err=%v", err)
			}
			ser.appGUI.ShowError(err, true)
			ser.appGUI.CloseMainWindow(true)
			return
		} else if n > 0 {
			if DEBUG_INFO {
				log.Printf("buf=%q", buf[:n])
			}
			v := parseData(buf[:n])
			s, ok := v.(*Status)
			if ok {
				ser.status = s
				ser.appGUI.UpdateStatusPage()
			}
			c, ok := v.(*Config)
			if ok {
				ser.config = c
				ser.appGUI.UpdateConfigPages()
			}
			_, ok = v.(SuccessApply)
			if ok {
				ser.appGUI.UpdateActionButtons(false)

				go ser.queryConfig()
				ser.appGUI.ShowMessage("Config successfully applied")
			}
			error, ok := v.(ErrorMessage)
			if ok && len(strings.Trim(error.Message, " ")) > 0 {
				ser.appGUI.UpdateActionButtons(true)
				ser.appGUI.ShowError(errors.New(error.Message), true)
			}
		}
	}
}

func (ser *Serial) checkPort() error {
	beginTime := time.Now()
	for {
		delta := time.Now().Sub(beginTime)
		if delta > time.Second*3 {
			break
		}

		buf := make([]byte, 256)
		ser.lockPort.Lock()
		n, err := ser.port.Read(buf)
		ser.lockPort.Unlock()
		if err != nil && err != io.EOF {
			return err
		} else if n > 0 {
			if DEBUG_INFO {
				log.Printf("buf=%q", buf[:n])
			}
			v := parseData(buf[:n])
			_, ok := v.(*Status)
			if ok {
				return nil
			}
		}
	}
	return errors.New("Couldn't got fan controller status")
}

func (ser *Serial) queryConfig() bool {
	cmd := "FCQ\r\n"
	if DEBUG_INFO {
		log.Printf("queryCmd=%s", cmd)
	}
	ser.lockPort.Lock()
	_, err := ser.port.Write([]byte(cmd))
	ser.lockPort.Unlock()
	if err != nil {
		log.Printf("err=%v", err)

		ser.appGUI.ShowError(err, true)
		return false
	}
	return true
}

func (ser *Serial) ConnectToController(portName string) bool {
	c := &serial.Config{Name: portName, Baud: 9600, ReadTimeout: time.Millisecond * 100}
	var err error
	ser.port, err = serial.OpenPort(c)
	if err != nil {
		ser.appGUI.ShowError(err, false)
		return false
	}
	err = ser.checkPort()
	if err != nil {
		ser.port.Close()
		ser.port = nil

		ser.appGUI.ShowError(err, false)
		return false
	}

	ser.appGUI.UpdateStatusPage()
	ser.appGUI.UpdateConfigPages()
	ser.appGUI.UpdateConfig(portName)

	go ser.queryConfig()
	go ser.readPort()

	return true
}

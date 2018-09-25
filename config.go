package main

import (
	"fmt"
	"log"
	"strings"
)

type Temperatures struct {
	SensorA int8
	SensorB int8
	SensorC int8
	SensorD int8
}

type Outputs struct {
	Fan1 int8
	Fan2 int8
	Fan3 int8
	Fan4 int8
}

type RPMS struct {
	Fan1A int16
	Fan1B int16
	Fan2A int16
	Fan2B int16
	Fan3A int16
	Fan3B int16
	Fan4A int16
	Fan4B int16
}

type Status struct {
	Temperatures Temperatures
	Outputs      Outputs
	RPMS         RPMS
}

const (
	SENSOR_NOT_CONNECTED = iota
	SENSOR_TYPE_C
	SENSOR_TYPE_F
)

const (
	SENSOR_A = iota
	SENSOR_B
	SENSOR_C
	SENSOR_D
	SENSOR_A_D
	SENSOR_B_D
	SENSOR_C_D
	MANUAL_CONTROL
)

const (
	FAN_NOT_CONNECTED = iota
	FAN_2_WIRE
	FAN_3_WIRE_X1_TACHO
	FAN_3_WIRE_X2_TACHO
	FAN_3_WIRE_X4_TACHO
	FAN_4_WIRE
)

type SensorTypes struct {
	SensorTypeA int8
	SensorTypeB int8
	SensorTypeC int8
	SensorTypeD int8
}

type Sensors struct {
	SensorTypeA int8
	SensorTypeB int8
	SensorTypeC int8
	SensorTypeD int8
}

type FanConfig struct {
	MinimumPower       int8
	SensorControlling  int8
	MinimumTemperature int16
	MaximumTemperature int16
	AllowStopped       bool
	FanTypeA           int8
	FanTypeB           int8
}

type Config struct {
	SensorTypes SensorTypes
	Fan1Config  FanConfig
	Fan2Config  FanConfig
	Fan3Config  FanConfig
	Fan4Config  FanConfig
}

type SuccessApply struct {
}

type ErrorMessage struct {
	Message string
}

func checkCommand(cmd, cmdName string) bool {
	if cmd == cmdName || cmd == (string(0x0)+cmdName) {
		return true
	}
	return false
}

func checkCommandPrefix(cmd, cmdName string) bool {
	if strings.HasPrefix(cmd, cmdName) || strings.HasPrefix(cmd, (string(0x0)+cmdName)) {
		return true
	}
	return false
}

func parseData(d []byte) interface{} {
	sd := string(d)
	ss := strings.Split(sd, "\r\n")
	for _, ss := range ss {
		rs := ""
		if ss != string(0x0) {
			rs += ss
		}
		sp := strings.Split(rs, ",")
		if len(sp) > 0 {
			if checkCommand(sp[0], "FCD") && len(sp) == 17 {
				status := &Status{
					Temperatures: Temperatures{
						SensorA: StrToInt8(sp[1]),
						SensorB: StrToInt8(sp[2]),
						SensorC: StrToInt8(sp[3]),
						SensorD: StrToInt8(sp[4]),
					},
					Outputs: Outputs{
						Fan1: StrToInt8(sp[5]),
						Fan2: StrToInt8(sp[6]),
						Fan3: StrToInt8(sp[7]),
						Fan4: StrToInt8(sp[8]),
					},
					RPMS: RPMS{
						Fan1A: StrToInt16(sp[9]),
						Fan1B: StrToInt16(sp[10]),
						Fan2A: StrToInt16(sp[11]),
						Fan2B: StrToInt16(sp[12]),
						Fan3A: StrToInt16(sp[13]),
						Fan3B: StrToInt16(sp[14]),
						Fan4A: StrToInt16(sp[15]),
						Fan4B: StrToInt16(sp[16]),
					},
				}
				if DEBUG_INFO {
					log.Printf("status=%s", ToJSON(status))
				}
				return status
			} else if checkCommand(sp[0], "FCR") && len(sp) == 33 {
				config := &Config{
					SensorTypes: SensorTypes{
						SensorTypeA: StrToInt8(sp[1]),
						SensorTypeB: StrToInt8(sp[2]),
						SensorTypeC: StrToInt8(sp[3]),
						SensorTypeD: StrToInt8(sp[4]),
					},
					Fan1Config: FanConfig{
						MinimumPower:       StrToInt8(sp[5]),
						SensorControlling:  StrToInt8(sp[6]),
						MinimumTemperature: StrToInt16(sp[7]),
						MaximumTemperature: StrToInt16(sp[8]),
						AllowStopped:       Int8ToBool(StrToInt8(sp[9])),
						FanTypeA:           StrToInt8(sp[10]),
						FanTypeB:           StrToInt8(sp[11]),
					},
					Fan2Config: FanConfig{
						MinimumPower:       StrToInt8(sp[12]),
						SensorControlling:  StrToInt8(sp[13]),
						MinimumTemperature: StrToInt16(sp[14]),
						MaximumTemperature: StrToInt16(sp[15]),
						AllowStopped:       Int8ToBool(StrToInt8(sp[16])),
						FanTypeA:           StrToInt8(sp[17]),
						FanTypeB:           StrToInt8(sp[18]),
					},
					Fan3Config: FanConfig{
						MinimumPower:       StrToInt8(sp[19]),
						SensorControlling:  StrToInt8(sp[20]),
						MinimumTemperature: StrToInt16(sp[21]),
						MaximumTemperature: StrToInt16(sp[22]),
						AllowStopped:       Int8ToBool(StrToInt8(sp[23])),
						FanTypeA:           StrToInt8(sp[24]),
						FanTypeB:           StrToInt8(sp[25]),
					},
					Fan4Config: FanConfig{
						MinimumPower:       StrToInt8(sp[26]),
						SensorControlling:  StrToInt8(sp[27]),
						MinimumTemperature: StrToInt16(sp[28]),
						MaximumTemperature: StrToInt16(sp[29]),
						AllowStopped:       Int8ToBool(StrToInt8(sp[30])),
						FanTypeA:           StrToInt8(sp[31]),
						FanTypeB:           StrToInt8(sp[32]),
					},
				}
				if DEBUG_INFO {
					log.Printf("config=%s", ToJSON(config))
				}
				return config
			} else if checkCommand(sp[0], "FCA") && len(sp) == 1 {
				return SuccessApply{}
			} else if checkCommandPrefix(sp[0], "ERR") && len(sp) == 1 {
				errMsg := strings.TrimPrefix(sp[0], "ERR:")
				if DEBUG_INFO {
					log.Printf("errMsg=%q", errMsg)
				}

				return ErrorMessage{Message: errMsg}
			}
		}
	}
	return nil
}

func configToStr(config *Config) string {
	return fmt.Sprintf("FCS,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d",
		config.SensorTypes.SensorTypeA, config.SensorTypes.SensorTypeB, config.SensorTypes.SensorTypeC, config.SensorTypes.SensorTypeD,
		config.Fan1Config.MinimumPower, config.Fan1Config.SensorControlling, config.Fan1Config.MinimumTemperature, config.Fan1Config.MaximumTemperature,
		BoolToInt(config.Fan1Config.AllowStopped), config.Fan1Config.FanTypeA, config.Fan1Config.FanTypeB,
		config.Fan2Config.MinimumPower, config.Fan2Config.SensorControlling, config.Fan2Config.MinimumTemperature, config.Fan2Config.MaximumTemperature,
		BoolToInt(config.Fan2Config.AllowStopped), config.Fan2Config.FanTypeA, config.Fan2Config.FanTypeB,
		config.Fan3Config.MinimumPower, config.Fan3Config.SensorControlling, config.Fan3Config.MinimumTemperature, config.Fan3Config.MaximumTemperature,
		BoolToInt(config.Fan3Config.AllowStopped), config.Fan3Config.FanTypeA, config.Fan3Config.FanTypeB,
		config.Fan4Config.MinimumPower, config.Fan4Config.SensorControlling, config.Fan4Config.MinimumTemperature, config.Fan4Config.MaximumTemperature,
		BoolToInt(config.Fan4Config.AllowStopped), config.Fan4Config.FanTypeA, config.Fan4Config.FanTypeB,
	)
}

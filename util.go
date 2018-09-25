package main

import (
	"encoding/json"
	"strconv"
)

func ToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func ToPrettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return ""
	}
	return string(b)
}

func Int8ToBool(i int8) bool {
	if i > 0 {
		return true
	}
	return false
}

func StrToInt8(s string) int8 {
	v, err := strconv.Atoi(s)
	if err == nil {
		return int8(v)
	}
	return 0
}

func StrToInt16(s string) int16 {
	v, err := strconv.Atoi(s)
	if err == nil {
		return int16(v)
	}
	return 0
}

func BoolToInt(b bool) int {
	if b == false {
		return 0
	} else {
		return 1
	}
}

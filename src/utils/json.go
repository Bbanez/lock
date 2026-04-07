package utils

import "encoding/json"

func SerializeJsonPretty[T any](st T) string {
	bytes, err := json.MarshalIndent(st, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func SerializeJson[T any](st T) string {
	bytes, err := json.Marshal(st)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

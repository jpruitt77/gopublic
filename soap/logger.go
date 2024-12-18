package soap

import (
	"encoding/json"
	"log"
)

func l(m ...interface{}) {
	if Verbose {
		log.Println(m...)
	}
}

func LogJSON(v interface{}) {
	if Verbose {
		json, err := json.MarshalIndent(v, "", " ")
		if err != nil {
			log.Println("Could not log json...")
			return
		}
		log.Println(string(json))
	}
}

func jsonDump(v interface{}) string {
	if !Verbose {
		return "not dumping"
	}
	jsonBytes, err := json.MarshalIndent(v, "", "	")
	if err != nil {
		return "error in json dump :: " + err.Error()
	}
	return string(jsonBytes)
}

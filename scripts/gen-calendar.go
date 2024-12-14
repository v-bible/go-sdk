package main

import (
	"encoding/json"
	"fmt"
	"os"

	utils "github.com/v-bible/go-sdk/pkg/utils"
)

func main() {
	if _, err := os.Stat("./dist"); os.IsNotExist(err) {
		err := os.Mkdir("./dist", 0o755)
		if err != nil {
			panic(err)
		}
	}

	for i := 2020; i < 2026; i++ {
		data, err := utils.GenerateCalendar(i, nil)
		if err != nil {
			panic(err)
		}

		jsonString, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(fmt.Sprintf("./dist/calendar-%d.json", i), jsonString, 0o644)
		if err != nil {
			panic(err)
		}
	}
}

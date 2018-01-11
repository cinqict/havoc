package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	indexName = "index.json"
)

type sample struct {
	Name     string `json:"name"`
	FileName string `json:"filename"`
}

type samples []sample

var Samples samples

func getSampleFileName(n string) string {
	for _, s := range Samples {
		if s.Name == n {
			return baseBucket + s.FileName
		}
	}
	fmt.Printf("sample not found: %s", n)
	os.Exit(1)
	return "nope"
}

func getSamplesFromS3() {
	resp, err := http.Get(baseBucket + indexName)
	if err != nil {
		log.Fatal(err)
	} else {
		defer resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(&Samples); err != nil {
		fmt.Println(err)
		fmt.Println("unable to load sample index")
		os.Exit(1)
	}

	fmt.Println("samples loaded")
}

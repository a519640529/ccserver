package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

type Config struct {
	Name      string
	Passwd    string
	Addr      string
	RemoteDir string
	LocalDir  string
}

func main() {
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.ParseFiles("./upload.temp")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile("./upload.txt", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}

	if err = tmpl.Execute(f, cfg); err != nil {
		log.Fatal(err)
	}
}

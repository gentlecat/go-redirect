package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

var (
	outputDir      = flag.String("out", "./out", "Output directory")
	configFilePath = flag.String("cfg", "./config.json", "Configuration file")

	fileTemplate = template.Must(template.ParseFiles("template.html"))
)

var config struct {
	Domain   string   `json:"domain"`
	Packages []string `json:"packages"`
}

func main() {
	flag.Parse()

	opDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	log.Printf("Generating the files (%s)...", opDir)
	defer log.Printf("Done in %v!", time.Since(start))

	configFile, err := os.Open(*configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(*outputDir, 0777); err != nil {
		log.Fatal(err)
	}

	for _, p := range config.Packages {
		generateFile(p)
	}

}

func generateFile(p string) {
	filePath := path.Join(*outputDir, p+".html")
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = fileTemplate.Execute(file, struct {
		Domain, Package string
	}{
		config.Domain,
		p,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Generated %s", filePath)
}

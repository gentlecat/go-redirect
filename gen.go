package main

import (
	"context"
	"flag"
	"html/template"
	"log"
	"os"
	"path"
	"time"

	"github.com/google/go-github/v33/github"
)

var (
	outputDir = flag.String("out", "./out", "Output directory")

	fileTemplate = template.Must(template.ParseFiles("template.html"))
)

func main() {
	flag.Parse()

	start := time.Now()
	defer log.Printf("Done in %v!", time.Since(start))

	domainName := os.Getenv("DOMAIN_NAME")
	githubUsername := os.Getenv("GITHUB_ACTOR")
	if domainName == "" {
		log.Fatal("Domain name must be specified in DOMAIN_NAME env variable")
	}
	if githubUsername == "" {
		log.Fatal("GitHub username must be specified in GITHUB_USERNAME env variable")
	}
	log.Printf("Got configuration [domainName=%s, githubUsername=%s]", domainName, githubUsername)

	log.Printf("Generating the files (at %s)...", *outputDir)

	if err := os.MkdirAll(*outputDir, 0777); err != nil {
		log.Fatal(err)
	}

	for _, p := range getRepositories(githubUsername) {
		if p.Language != nil && *p.Language == "Go" {
			generateFile(domainName, *p.Name)
		}
	}
}

func generateFile(domainName, packageName string) {
	filePath := path.Join(*outputDir, packageName+".html")
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = fileTemplate.Execute(file, struct {
		Domain, Package string
	}{
		domainName,
		packageName,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Generated %s", filePath)
}

func getRepositories(username string) []*github.Repository {
	log.Printf("Getting the list of repositories for user %s from GitHub", username)
	repos, _, err := github.NewClient(nil).Repositories.List(context.Background(), username, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %v repositories", len(repos))
	return repos
}

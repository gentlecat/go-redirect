package main

import (
	"context"
	"flag"
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	git "github.com/go-git/go-git/v5"
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
			log.Printf("Found Go repository \"%s\". Generating paths...", *p.Name)
			for _, d := range getRepositoryPaths(p) {
				generateFile(domainName, d)
			}
		} else {
			log.Printf("Primary language of repository \"%s\" is not Go. Ignoring.", *p.Name)
		}
	}
}

func generateFile(domainName, outputPath string) {
	filePath := path.Join(*outputDir, outputPath+".html")
	log.Printf("Generating %s", filePath)
	if err := os.MkdirAll(path.Dir(filePath), 0777); err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = fileTemplate.Execute(file, struct {
		Domain, Package string
	}{
		domainName,
		outputPath,
	})
	if err != nil {
		log.Fatal(err)
	}
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

func getRepositoryPaths(repo *github.Repository) []string {
	tmpDir := path.Join(*outputDir, "tmp")
	tmpRepoPath := path.Join(tmpDir, *repo.Name)
	_, err := git.PlainClone(tmpRepoPath, false, &git.CloneOptions{
		URL: *repo.CloneURL,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpRepoPath)

	return listDirs(tmpRepoPath, tmpDir)
}

func listDirs(curPath, tmpDir string) []string {
	var dirs []string

	err := filepath.Walk(curPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if strings.Contains(p, ".git") {
			return nil
		}
		resultPath := strings.TrimPrefix(p, tmpDir)
		_, i := utf8.DecodeRuneInString(resultPath)
		dirs = append(dirs, strings.TrimPrefix(resultPath, resultPath[i:]))
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	return dirs
}

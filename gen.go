package main

import (
	"context"
	"flag"
	"fmt"
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
	defer fmt.Printf("Done in %v!\n", time.Since(start))

	domainName := os.Getenv("DOMAIN_NAME")
	githubUsername := os.Getenv("GITHUB_ACTOR")
	if domainName == "" {
		log.Fatal("Domain name must be specified in DOMAIN_NAME env variable")
	}
	if githubUsername == "" {
		log.Fatal("GitHub username must be specified in GITHUB_USERNAME env variable")
	}
	fmt.Printf("Got configuration [domainName=%s, githubUsername=%s]\n", domainName, githubUsername)

	fmt.Printf("Generating the files (at %s)...\n", *outputDir)

	if err := os.MkdirAll(*outputDir, 0777); err != nil {
		log.Fatal(err)
	}

	for _, p := range getRepositories(githubUsername) {
		if p.Language != nil && *p.Language == "Go" {
			fmt.Printf("> Found Go repository \"%s\". Generating paths...\n", *p.Name)
			for _, repoPath := range getRepositoryPaths(p) {
				generateFile(domainName, *p.Name, repoPath)
			}
		} else {
			fmt.Printf("> Primary language of repository \"%s\" is not Go. Ignoring.\n", *p.Name)
		}
	}
}

func generateFile(domainName, packageName, outputPath string) {
	filePath := path.Join(*outputDir, outputPath+".html")
	fmt.Printf("  + %s", outputPath+".html\n")
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
		packageName,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func getRepositories(username string) []*github.Repository {
	fmt.Printf("Getting the list of repositories for user %s from GitHub... ", username)
	repos, _, err := github.NewClient(nil).Repositories.List(context.Background(), username, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %v repositories.\n", len(repos))
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

// listDirs returns a list of paths with all sub-directories within a given directory.
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

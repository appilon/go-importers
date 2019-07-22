package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/appilon/go-importers/github"
	"github.com/appilon/go-importers/godoc"
	"github.com/appilon/go-importers/util"
)

var packages = [...]string{
	"helper/acctest",
	"helper/customdiff",
	"helper/encryption",
	"helper/hashcode",
	"helper/logging",
	"helper/mutexkv",
	"helper/pathorcontents",
	"helper/resource",
	"helper/schema",
	"helper/structure",
	"helper/validation",
	"httpclient",
	"plugin",
	"terraform",
}

type repo struct {
	Stars    int                 `json:"stars"`
	Packages map[string][]string `json:"packages"`
}

func main() {
	client := github.NewClient(context.Background(), util.MustEnv("GITHUB_PERSONAL_TOKEN"))
	r := make(map[string]*repo)

	log.Printf("Loading ignorelist from GitHub...")
	ignore, err := loadIgnoreSet(client)
	if err != nil {
		log.Fatalf("Error loading set of projects to ignore: %s", err)
	}
	log.Printf("Ignorelist of %d entries loaded.", len(ignore))

	for _, pkg := range packages {
		pkg = "github.com/hashicorp/terraform/" + pkg

		log.Printf("Discovering importers of %q on godoc ... ", pkg)
		importers, err := godoc.ListImporters(pkg, ignore, true)
		if err != nil {
			log.Fatalf("Error fetching importers of %s: %s", pkg, err)
		}
		log.Printf("%d found.", len(importers))

		for _, imp := range importers {
			log.Printf("Processing %q ...", imp)
			// non github repos will have the full package path
			// it will be unclear to us where the project namespace begins
			// and where the package tree begins
			proj := github.RepoRoot(imp)
			if _, ok := r[proj]; !ok {
				var stars int
				if strings.HasPrefix(imp, "github.com") {
					var err error
					owner, repo := github.OwnerRepo(imp)
					stars, err = client.GetStars(owner, repo)
					if err != nil {
						log.Println(err)
					}
				}
				r[proj] = &repo{
					Stars: stars,
					Packages: map[string][]string{
						imp: {pkg},
					},
				}
			} else {
				r[proj].Packages[imp] = append(r[proj].Packages[imp], pkg)
			}
		}
	}

	if err := json.NewEncoder(os.Stdout).Encode(r); err != nil {
		log.Fatalf("Error writing report: %s", err)
	}
}

func loadIgnoreSet(client *github.Client) (map[string]bool, error) {
	log.Printf("Listing repositories under %q ...", "terraform-providers")
	ignoreForksOf, err := client.ListRepositories("terraform-providers")
	if err != nil {
		return nil, err
	}
	log.Printf("%d repositories found.", len(ignoreForksOf))

	ignoreForksOf = append(ignoreForksOf, "github.com/hashicorp/terraform", "github.com/hashicorp/otto")

	var ignoredForks []string
	for i, upstream := range ignoreForksOf {
		log.Printf("Listing forks of %q (%d/%d)...", upstream, i+1, len(ignoreForksOf))
		owner, repo := github.OwnerRepo(upstream)
		forks, err := client.ListForks(owner, repo)
		if err != nil {
			return nil, err
		}
		log.Printf("%d forks of %q found.", len(forks), upstream)
		ignoredForks = append(ignoredForks, forks...)
	}

	return util.StringListToSet(append(ignoredForks, ignoreForksOf...)), nil
}

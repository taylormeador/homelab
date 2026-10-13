package server

import (
	"log"
	"os"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/client"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
)

func CloneRepo(url string, directory string, token string) {
	err := os.RemoveAll(directory)
	if err != nil {
		log.Fatalf("Error deleting directory: %v\n", err)
	}

	r, err := git.PlainClone(directory, &git.CloneOptions{
		URL:      url,
		Progress: nil,
		ClientOptions: []client.Option{
			client.WithHTTPAuth(&http.BasicAuth{
				Username: "abc123",
				Password: token,
			}),
		},
	})
	if err != nil {
		log.Fatalf("error cloning repo: %s", err.Error())
	}
	defer func() { _ = r.Close() }()

	ref, err := r.Head()
	if err != nil {
		log.Fatalf("error getting head: %s", err.Error())
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		log.Fatalf("error while getting commit: %s", err.Error())
	}

	log.Println(commit)
}

package server

import (
	"log"
	"os"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/client"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
)

func CloneRepo(url string, directory string, token string) error {
	err := os.RemoveAll(directory)
	if err != nil {
		return err
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
		return err
	}
	defer func() { _ = r.Close() }()

	ref, err := r.Head()
	if err != nil {
		return err
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	log.Printf("cloned repo and checked out commit %s: %s", commit.Hash, commit.Message)

	return nil
}

package repo

import (
	"log"
	"os"
	"sync"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/client"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Repo struct {
	mu        sync.Mutex
	url       string
	directory string
	token     string
	repo      *git.Repository
}

func NewRepo(url string, directory string, token string) (*Repo, error) {
	r := &Repo{url: url, directory: directory, token: token}
	if err := r.RefreshLocal(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Repo) authOptions() []client.Option {
	return []client.Option{
		client.WithHTTPAuth(&http.BasicAuth{
			Username: "abc123",
			Password: r.token,
		}),
	}
}

func (r *Repo) RefreshLocal() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := os.RemoveAll(r.directory); err != nil {
		return status.Errorf(codes.Internal, "failed to clean repo directory: %v", err)
	}

	repo, err := git.PlainClone(r.directory, &git.CloneOptions{
		URL:           r.url,
		Progress:      nil,
		ClientOptions: r.authOptions(),
	})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to clone repo: %v", err)
	}
	r.repo = repo

	ref, err := repo.Head()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get HEAD: %v", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get commit: %v", err)
	}

	log.Printf("cloned repo and checked out commit %s: %s", commit.Hash, commit.Message)

	return nil
}

func (r *Repo) publishChanges(commitMessage string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get worktree: %v", err)
	}

	if _, err := w.Add("."); err != nil {
		return status.Errorf(codes.Internal, "failed to stage changes: %v", err)
	}

	_, err = w.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "dexterd",
			Email: "dexterd@taylor-meador.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to commit: %v", err)
	}

	if err := r.repo.Push(&git.PushOptions{
		ClientOptions: r.authOptions(),
	}); err != nil {
		return status.Errorf(codes.Internal, "failed to push: %v", err)
	}

	log.Printf("pushed: %s", commitMessage)
	return nil
}

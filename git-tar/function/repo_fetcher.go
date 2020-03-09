package function

import (
	"fmt"
	"log"
	"os/exec"
)

type RepoFetcher interface {
	Clone(url, path string) error
	Checkout(commitID, path string) error
}

type GitRepoFetcher struct {
}

func (c GitRepoFetcher) Clone(url, path string) error {
	git := exec.Command("git", "clone", url)
	git.Dir = path
	log.Printf("Cloning %s to %s", url, path)

	err := git.Start()
	if err != nil {
		return fmt.Errorf("Cannot start git: %t", err)
	}

	return git.Wait()
}

func (c GitRepoFetcher) Checkout(commitID, path string) error {
	git := exec.Command("git", "checkout", commitID)
	git.Dir = path
	log.Printf("Checking out SHA: %s to %s", commitID, path)

	err := git.Start()
	if err != nil {
		return fmt.Errorf("Cannot start git checkout: %t", err)
	}

	err = git.Wait()
	return err
}

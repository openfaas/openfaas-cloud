package function

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// Handle a serverless request
func Handle(req []byte) []byte {

	pushEvent := PushEvent{}
	err := json.Unmarshal(req, &pushEvent)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}

	path, err := clone(pushEvent)
	if err != nil {
		log.Println("Clone ", err.Error())
		os.Exit(-2)
	}

	path, err = shrinkwrap(pushEvent, path)
	if err != nil {
		log.Println("Shrinkwrap ", err.Error())
		os.Exit(-2)
	}

	path, err = makeTar(pushEvent, path)
	if err != nil {
		log.Println("Tar ", err.Error())
		os.Exit(-2)
	}

	return []byte("Tar at: " + path)
}

func shrinkwrap(pushEvent PushEvent, filePath string) (string, error) {
	buildCmd := exec.Command("faas-cli", "build", "-f", "stack.yml", "--shrinkwrap")
	buildCmd.Dir = filePath
	err := buildCmd.Start()
	if err != nil {
		return "", fmt.Errorf("Cannot start faas-cli build: %t", err)
	}
	err = buildCmd.Wait()

	return "", err
}

func makeTar(pushEvent PushEvent, filePath string) (string, error) {
	contextTar, err := os.Create(path.Join(filePath, "context.tar"))
	// contextTar, err:= os.Open(path.Join(filepath, "context.tar"))
	if err != nil {
		return "", err
	}

	tarWriter := tar.NewWriter(contextTar)
	defer tarWriter.Close()
	// base := filepath.Join( filePath+ "/build/")
	err = filepath.Walk(filePath, func(path string, f os.FileInfo, pathErr error) error {
		if pathErr != nil {
			return pathErr
		}

		if f.Name() == "context.tar" {
			return nil
		}

		targetFile, err1 := os.Open(path)
		log.Println(path)

		if err1 != nil {
			return err1
		}

		header, headerErr := tar.FileInfoHeader(f, f.Name())
		if headerErr != nil {
			return headerErr
		}
		header.Name = strings.TrimPrefix(path, base)
		if err1 = tarWriter.WriteHeader(header); err != nil {
			return err1
		}

		if f.Mode().IsDir() {
			return nil
		}

		_, err1 = io.Copy(tarWriter, targetFile)
		return err1
	})
	if err != nil {
		return "", err
	}

	return path.Join(filePath, "context.tar"), err
}

func clone(pushEvent PushEvent) (string, error) {
	workDir := os.TempDir()
	destPath := path.Join(workDir, pushEvent.Repository.Name)
	if _, err := os.Stat(destPath); err == nil {
		truncateErr := os.RemoveAll(destPath)
		if truncateErr != nil {
			return "", truncateErr
		}
	}

	git := exec.Command("git", "clone", pushEvent.Repository.CloneURL)
	git.Dir = workDir
	err := git.Start()
	if err != nil {
		return "", fmt.Errorf("Cannot start git: %t", err)
	}
	err = git.Wait()

	git = exec.Command("git", "checkout", pushEvent.AfterCommitID)
	git.Dir = destPath
	err = git.Start()
	if err != nil {
		return "", fmt.Errorf("Cannot start git checkout: %t", err)
	}
	err = git.Wait()

	return destPath, err
}

type PushEvent struct {
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
	}
	AfterCommitID string `json:"after"`
}

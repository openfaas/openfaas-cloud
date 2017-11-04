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

	"github.com/openfaas/faas-cli/stack"
)

// Handle a serverless request
func Handle(req []byte) []byte {

	pushEvent := PushEvent{}
	err := json.Unmarshal(req, &pushEvent)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}

	clonePath, err := clone(pushEvent)
	if err != nil {
		log.Println("Clone ", err.Error())
		os.Exit(-2)
	}

	stack, err := parseYAML(pushEvent, clonePath)
	if err != nil {
		log.Println("parseYAML ", err.Error())
		os.Exit(-2)
	}

	var shrinkWrapPath string
	shrinkWrapPath, err = shrinkwrap(pushEvent, clonePath)
	if err != nil {
		log.Println("Shrinkwrap ", err.Error())
		os.Exit(-2)
	}

	var tars []string
	tars, err = makeTar(pushEvent, shrinkWrapPath, stack)
	if err != nil {
		log.Println("Tar ", err.Error())
		os.Exit(-2)
	}

	return []byte(fmt.Sprintf("Tar at: %s", tars))
}

func parseYAML(pushEvent PushEvent, filePath string) (*stack.Services, error) {
	parsed, err := stack.ParseYAMLFile(path.Join(filePath, "stack.yml"), "", "")
	return parsed, err
}

func shrinkwrap(pushEvent PushEvent, filePath string) (string, error) {
	buildCmd := exec.Command("faas-cli", "build", "-f", "stack.yml", "--shrinkwrap")
	buildCmd.Dir = filePath
	err := buildCmd.Start()
	if err != nil {
		return "", fmt.Errorf("Cannot start faas-cli build: %t", err)
	}
	err = buildCmd.Wait()

	return filePath, err
}

func makeTar(pushEvent PushEvent, filePath string, services *stack.Services) ([]string, error) {
	tars := []string{}
	fmt.Printf("Tar up %s\n", filePath)
	for k, v := range services.Functions {
		fmt.Println("Start work on: ", v.Handler, k)
		contextTar, err := os.Create(path.Join(filePath, fmt.Sprintf("%s.tar", k)))
		if err != nil {
			return []string{}, err
		}

		tarWriter := tar.NewWriter(contextTar)
		defer tarWriter.Close()

		base := filepath.Join(filePath, filepath.Join("build", k))
		fmt.Println("Base: ", base, filePath, k)
		err = filepath.Walk(base, func(path string, f os.FileInfo, pathErr error) error {
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
			log.Println("trim ", path, base)
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
			return []string{}, err
		}
		tars = append(tars, path.Join(filePath, "context.tar"))
	}

	return tars, nil
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

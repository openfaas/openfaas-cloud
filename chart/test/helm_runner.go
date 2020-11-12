package test

import (
	execute "github.com/alexellis/go-execute/pkg/v1"
	"log"
)

func helmRunner(parts ...string) (execute.ExecResult, error) {
	return helmRunnerToLocation("./tmp", parts...)

}
func helmRunnerToLocation(location string, parts ...string) (execute.ExecResult, error) {
	firstParts := []string{
		"template",
		"--output-dir", location,
		"../openfaas-cloud",
		"-f", "values.yml",
	}
	fullParts := append(firstParts, parts...)
	task := execute.ExecTask{
		Command: "helm",
		Args:    fullParts,
	}

	res, err := task.Execute()

	if err != nil {
		log.Panic(err)
	}

	if res.ExitCode != 0 {
		log.Panicf("Command exit code: %q, Command StdErr: %q", res.ExitCode, res.Stderr)
	}

	return res, err
}

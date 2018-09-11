package function

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

// Handle a serverless request
func Handle(req []byte) string {
	tmpl, err := template.ParseFiles("index.html")

	if err != nil {
		log.Panic("Error loading template file: ", err)
	}

	var tpl bytes.Buffer

	template := PipelineTemplate{}

	var logData string
	if query, exists := os.LookupEnv("Http_Query"); exists {
		if qs, err := url.ParseQuery(query); err == nil {
			sha := qs.Get("commitSHA")
			repoPath := qs.Get("repoPath")
			function := qs.Get("function")

			path := os.Getenv("gateway_url") + fmt.Sprintf("/function/pipeline-log?commitSHA=%s&repoPath=%s&function=%s", sha, repoPath, function)
			res, err := http.Get(path)

			if err != nil {
				log.Printf("error %s when fetching URL %s", err.Error(), path)
				os.Exit(1)
			}
			defer res.Body.Close()

			bytesOut, _ := ioutil.ReadAll(res.Body)
			logData = string(bytesOut)
			template.Log = logData
			template.CommitSHA = sha
			template.Function = function
			template.RepoPath = repoPath
		}
	}

	err = tmpl.Execute(&tpl, template)

	if err != nil {
		log.Panic("Error executing template: ", err)
	}

	return tpl.String()
}

type PipelineTemplate struct {
	Log       string
	Function  string
	RepoPath  string
	CommitSHA string
}

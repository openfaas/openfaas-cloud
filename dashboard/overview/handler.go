package function

import (
	"bytes"
	"html/template"
	"log"
	"net/url"
	"os"
)

type UserData struct {
	User           string
	SelectedRepo   string
	PublicURL      string
	PrettyURL      string
	QueryPrettyURL string
	BuildBranch    string
}

// Handle a serverless request
func Handle(req []byte) string {
	tmpl, err := template.ParseFiles("index.html")

	if err != nil {
		log.Panic("Error loading template file: ", err)
	}

	userData1 := UserData{
		PublicURL:      os.Getenv("public_url"),
		PrettyURL:      os.Getenv("pretty_url"),
		QueryPrettyURL: os.Getenv("query_pretty_url"),
	}

	vals, _ := url.ParseQuery(os.Getenv("Http_Query"))

	user := vals.Get("user")
	repo := vals.Get("repo")

	userData1.User = user
	userData1.SelectedRepo = repo
	userData1.BuildBranch = buildBranch()

	log.Println("User: ", user, " Repo: ", repo)

	var tpl bytes.Buffer

	err = tmpl.Execute(&tpl, userData1)

	if err != nil {
		log.Panic("Error executing template: ", err)
	}

	return tpl.String()
}

func buildBranch() string {
	branch := os.Getenv("build_branch")
	if branch == "" {
		return "master"
	}
	return branch
}

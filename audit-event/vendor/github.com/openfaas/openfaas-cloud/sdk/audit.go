package sdk

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func PostAudit(auditEvent AuditEvent) {
	c := http.Client{}
	bytesOut, _ := json.Marshal(&auditEvent)
	reader := bytes.NewBuffer(bytesOut)

	req, _ := http.NewRequest(http.MethodPost, os.Getenv("audit_url"), reader)

	res, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
}

type AuditEvent struct {
	Source  string
	Message string
	Owner   string
	Repo    string
}

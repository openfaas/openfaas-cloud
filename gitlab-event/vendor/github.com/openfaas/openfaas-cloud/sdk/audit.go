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
	auditURL := os.Getenv("audit_url")

	if len(auditURL) == 0 {
		log.Println("PostAudit invalid auditURL, empty string")
		return
	}

	req, _ := http.NewRequest(http.MethodPost, auditURL, reader)

	res, err := c.Do(req)
	if err != nil {
		log.Println("PostAudit", err)
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

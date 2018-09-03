package sdk

type Audit interface {
	Post(AuditEvent) error
}

type NilLogger struct {
}

func (l NilLogger) Post(auditEvent AuditEvent) error {
	return nil
}

type AuditLogger struct {
}

func (l AuditLogger) Post(auditEvent AuditEvent) error {
	PostAudit(auditEvent)
	return nil
}

type StatusReporter interface {
	Report(status *Status)
}

type NilStatusReporter struct {
}

func (r NilStatusReporter) Report(status *Status) {
	return
}

type GithubStatusReporter struct {
	HmacKey string
	Gateway string
}

func (r GithubStatusReporter) Report(status *Status) {
	ReportStatus(status, r.Gateway, r.HmacKey)
}

type Secret interface {
	Read(string) (string, error)
}

type NilSecretReader struct {
}

func (r NilSecretReader) Read(key string) (string, error) {
	return "", nil
}

type SecretReader struct {
}

func (r SecretReader) Read(key string) (string, error) {
	return ReadSecret(key)
}

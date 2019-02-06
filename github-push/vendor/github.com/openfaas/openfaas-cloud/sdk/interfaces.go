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

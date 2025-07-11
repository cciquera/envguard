package cmd

type ScanResult struct {
	Source     string `json:"source"`     // e.g., terraform
	Resource   string `json:"resource"`   // e.g., aws_instance.web
	ChangeType string `json:"changeType"` // e.g., update
	Severity   string `json:"severity"`   // info, warning, critical
	Message    string `json:"message"`
}

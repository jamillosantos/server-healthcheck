package svchealthcheck

const (
	HealthPath = "/healthz"
	ReadyPath  = "/readyz"
)

type CheckResponse struct {
	StatusCode int                           `json:"-"`
	Status     string                        `json:"status"`
	Checks     map[string]CheckResponseEntry `json:"checks"`
}

type CheckResponseEntry struct {
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration"`
}

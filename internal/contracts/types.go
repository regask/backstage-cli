package contracts

import "encoding/json"

// OverlayResponse is the /deploy-management/service/overlays payload. baseEnv
// is the service's base config.env text; per-env overlay texts key on env name.
type OverlayResponse struct {
	Service       string            `json:"service"`
	BaseEnv       string            `json:"baseEnv"`
	EnvOverlays   map[string]string `json:"envOverlays"`
	SecretBase    string            `json:"secretBase"`
	SecretOverlay map[string]string `json:"secretOverlays"`
	Vaults        map[string]string `json:"vaults"` // env -> Azure Key Vault name (may be empty)
}

// MatrixRow is one service's deploy state per environment.
type MatrixRow struct {
	Service      string            `json:"service"`
	Environments map[string]string `json:"environments"` // env -> deployed version/sha
}

type MatrixResponse struct {
	Rows      []MatrixRow `json:"rows"`
	UpdatedAt string      `json:"updatedAt"`
}

type TicketLookupResponse struct {
	Results map[string][]string `json:"results"` // ticket -> envs it is deployed to
}

// ApprovalRequest is the /approvals/requests/:id payload (subset used by the
// CLI). Mirrors packages/backend/src/modules/approvals/types.ts. resultUrl is
// the release link (published GitHub release for release kinds, PR URL for
// gitops kinds); payload carries kind-specific data incl. the originating
// scaffolder taskId.
type ApprovalRequest struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	Requester string          `json:"requester"`
	Status    string          `json:"status"`
	Title     string          `json:"title"`
	Summary   string          `json:"summary"`
	ResultURL string          `json:"resultUrl,omitempty"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
	Approver  string          `json:"approver,omitempty"`
	Error     string          `json:"error,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// TaskID extracts the originating scaffolder taskId from the payload, if any,
// for the "runs" backlink. Returns "" when absent.
func (a ApprovalRequest) TaskID() string {
	if len(a.Payload) == 0 {
		return ""
	}
	var p struct {
		TaskID string `json:"taskId"`
	}
	_ = json.Unmarshal(a.Payload, &p)
	return p.TaskID
}

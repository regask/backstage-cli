package contracts

import "encoding/json"

// EnvDeploy is one environment's deploy state for a service.
type EnvDeploy struct {
	Tag          string `json:"tag,omitempty"`
	ArgocdApp    string `json:"argocdApp,omitempty"`
	ArgocdURL    string `json:"argocdUrl,omitempty"`
	SyncStatus   string `json:"syncStatus,omitempty"`
	HealthStatus string `json:"healthStatus,omitempty"`
	CommitURL    string `json:"commitUrl,omitempty"`
	PublishedAt  string `json:"publishedAt,omitempty"`
}

// MatrixRow is one service's deploy state across environments. The
// /deploy-management/matrix endpoint returns a bare array of these.
type MatrixRow struct {
	ServiceRef  string               `json:"serviceRef"`
	ServiceName string               `json:"serviceName"`
	Owner       string               `json:"owner,omitempty"`
	Type        string               `json:"type,omitempty"`
	Envs        map[string]EnvDeploy `json:"envs"`
	// drift is present in the payload but the CLI doesn't render it yet.
}

// OverlayEnv is a service's per-environment overlay text.
type OverlayEnv struct {
	SecretsText       string `json:"secretsText,omitempty"`
	EnvText           string `json:"envText,omitempty"`
	KustomizationText string `json:"kustomizationText,omitempty"`
}

// OverlayBundle is the /deploy-management/service/overlays payload: base
// config plus per-env overlays.
type OverlayBundle struct {
	ServiceRef      string                `json:"serviceRef"`
	BaseSecretsText string                `json:"baseSecretsText,omitempty"`
	BaseEnvText     string                `json:"baseEnvText,omitempty"`
	Overlays        map[string]OverlayEnv `json:"overlays"`
}

// TicketCommit is a single commit associated with a ticket, deployed to the
// listed envs.
type TicketCommit struct {
	Sha      string   `json:"sha"`
	ShortSha string   `json:"shortSha"`
	Message  string   `json:"message"`
	URL      string   `json:"url,omitempty"`
	Ticket   string   `json:"ticket"`
	Envs     []string `json:"envs"`
	Date     string   `json:"date,omitempty"`
}

// TicketServiceResult is one service's match for a ticket lookup.
type TicketServiceResult struct {
	ServiceRef   string         `json:"serviceRef"`
	Slug         string         `json:"slug"`
	Count        int            `json:"count"`
	DeployedEnvs []string       `json:"deployedEnvs"`
	Commits      []TicketCommit `json:"commits"`
}

// TicketLookupResult is the /deploy-management/ticket-lookup payload.
type TicketLookupResult struct {
	Services          []TicketServiceResult `json:"services"`
	NotFound          []string              `json:"notFound"`
	SearchRateLimited bool                  `json:"searchRateLimited"`
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

package tui

import "github.com/regask/backstage-cli/internal/contracts"

type servicesLoadedMsg struct{ Rows []contracts.MatrixRow }
type approvalsLoadedMsg struct{ Items []contracts.ApprovalRequest }
type switchViewMsg struct{ View string }
type bannerMsg struct {
	Text  string
	IsErr bool
}
type errMsg struct{ Err error }

func (e errMsg) Error() string { return e.Err.Error() }

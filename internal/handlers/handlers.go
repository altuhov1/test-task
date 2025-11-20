package handlers

import (
	"test-task/internal/services"
)

type Handler struct {
	TeamManag        services.TeamManager
	UserManag        services.UserManager
	PullRequestManag services.PullRequestManager
}

func NewHandler(
	TeamManag services.TeamManager,
	UserManag services.UserManager,
	PullRequestManag services.PullRequestManager,
) (*Handler, error) {

	return &Handler{
		TeamManag:        TeamManag,
		UserManag:        UserManag,
		PullRequestManag: PullRequestManag,
	}, nil
}

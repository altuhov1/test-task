package services

import (
	"context"
	"test-task/internal/models"
)

type TeamManager interface {
	CreateTeam(ctx context.Context, team models.Team) (*models.Team, error)
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
}

type UserManager interface {
}

type PullRequestManager interface {
}

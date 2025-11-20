package storage

import (
	"context"
	"test-task/internal/models"
)

type PullReqStorage interface {
}

type TeamStorage interface {
	CreateTeam(ctx context.Context, team models.Team) error
	GetTeamInfo(ctx context.Context, teamName string) (*models.Team, error)
}

type UserStorage interface {
}

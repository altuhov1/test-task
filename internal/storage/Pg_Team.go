package storage

import (
	"context"
	"fmt"
	"test-task/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewTeamPostgresStorage(pool *pgxpool.Pool) *TeamPostgresStorage {

	return &TeamPostgresStorage{pool: pool}
}

func (s *TeamPostgresStorage) CreateTeam(ctx context.Context, team models.Team) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)", team.TeamName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		return models.ErrTeamExists
	}

	for _, member := range team.Members {
		if err := s.createUser(ctx, tx, member); err != nil {
			return fmt.Errorf("failed to create user %s: %w", member.UserID, err)
		}
	}

	userIDs := make([]string, len(team.Members))
	for i, member := range team.Members {
		userIDs[i] = member.UserID
	}

	_, err = tx.Exec(ctx, "INSERT INTO teams (name, user_ids) VALUES ($1, $2)", team.TeamName, userIDs)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *TeamPostgresStorage) createUser(ctx context.Context, tx pgx.Tx, user models.TeamMember) error {
	query := `
		INSERT INTO users (user_id, is_active) 
		VALUES ($1, $2)
		ON CONFLICT (user_id) 
		DO UPDATE SET is_active = EXCLUDED.is_active
	`
	_, err := tx.Exec(ctx, query, user.UserID, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create/update user: %w", err)
	}
	return nil
}

func (s *TeamPostgresStorage) GetTeamInfo(ctx context.Context, teamName string) (*models.Team, error) {
	query := `
        SELECT t.name, u.user_id, u.is_active
        FROM teams t
        CROSS JOIN UNNEST(t.user_ids) AS user_id
        JOIN users u USING (user_id)
        WHERE t.name = $1
        ORDER BY u.user_id
    `

	rows, err := s.pool.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query team: %w", err)
	}
	defer rows.Close()

	var team models.Team
	var members []models.TeamMember

	for rows.Next() {
		var member models.TeamMember
		var teamName string

		err := rows.Scan(&teamName, &member.UserID, &member.IsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}

		team.TeamName = teamName
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team members: %w", err)
	}

	if len(members) == 0 {
		return nil, models.ErrNotFound
	}

	team.Members = members
	return &team, nil
}

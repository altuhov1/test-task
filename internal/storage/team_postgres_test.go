package storage

import (
	"context"
	"test-task/internal/models"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			user_id TEXT PRIMARY KEY,
			is_active BOOLEAN NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS teams (
			name TEXT PRIMARY KEY,
			user_ids TEXT[] NOT NULL
		);
	`)
	require.NoError(t, err)

	return pool
}

func TestTeamPostgresStorage_CreateTeam_Success(t *testing.T) {
	pool := setupTestDB(t)
	storage := NewTeamPostgresStorage(pool)
	ctx := context.Background()

	team := models.Team{
		TeamName: "backend",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	err := storage.CreateTeam(ctx, team)
	require.NoError(t, err)

	createdTeam, err := storage.GetTeamInfo(ctx, "backend")
	require.NoError(t, err)
	assert.Equal(t, "backend", createdTeam.TeamName)
	assert.Len(t, createdTeam.Members, 2)
}

func TestTeamPostgresStorage_CreateTeam_AlreadyExists(t *testing.T) {
	pool := setupTestDB(t)
	storage := NewTeamPostgresStorage(pool)
	ctx := context.Background()

	team := models.Team{
		TeamName: "payments",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	err := storage.CreateTeam(ctx, team)
	require.NoError(t, err)

	err = storage.CreateTeam(ctx, team)
	assert.ErrorIs(t, err, models.ErrTeamExists)
}

func TestTeamPostgresStorage_GetTeamInfo_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	storage := NewTeamPostgresStorage(pool)
	ctx := context.Background()

	// Пытаемся получить несуществующую команду
	team, err := storage.GetTeamInfo(ctx, "nonexistent")
	assert.ErrorIs(t, err, models.ErrNotFound)
	assert.Nil(t, team)
}

func TestTeamPostgresStorage_CreateTeam_UpdatesUserActivity(t *testing.T) {
	pool := setupTestDB(t)
	storage := NewTeamPostgresStorage(pool)
	ctx := context.Background()

	team1 := models.Team{
		TeamName: "team1",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}
	err := storage.CreateTeam(ctx, team1)
	require.NoError(t, err)

	team2 := models.Team{
		TeamName: "team2",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: false},
		},
	}
	err = storage.CreateTeam(ctx, team2)
	require.NoError(t, err)

	team, err := storage.GetTeamInfo(ctx, "team2")
	require.NoError(t, err)
	assert.False(t, team.Members[0].IsActive)
}

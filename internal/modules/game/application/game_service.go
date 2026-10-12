package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	db "xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrGameNotFound   = errors.New("game not found")
	ErrRegionNotFound = errors.New("region not found")
	ErrServerNotFound = errors.New("server not found")

	ErrGameSlugTaken   = errors.New("game slug already exists")
	ErrRegionSlugTaken = errors.New("region slug already exists")
	ErrServerSlugTaken = errors.New("server slug already exists")

	ErrInvalidName = errors.New("invalid name")
	ErrInvalidSlug = errors.New("invalid slug")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type GameService struct {
	queries *db.Queries
}

func NewGameService(pool *pgxpool.Pool) *GameService {
	return &GameService{
		queries: db.New(pool),
	}
}

type RegisterGameInput struct {
	Name        string
	Slug        string
	Description string
	Publisher   string
}
type RegisterGameRegionInput struct {
	Name   string
	Slug   string
}
type RegisterGameRegionServerInput struct {
	Name     string
	Slug     string
}

func (g *GameService) RegisterGame(ctx context.Context, input RegisterGameInput) (db.Game, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.Description = strings.TrimSpace(input.Description)
	input.Publisher = strings.TrimSpace(input.Publisher)

	if len(input.Name) < 2 || len(input.Name) > 255 {
		return db.Game{}, ErrInvalidName
	}
	if len(input.Slug) < 2 || len(input.Slug) > 255 || !slugPattern.MatchString(input.Slug) {
		return db.Game{}, ErrInvalidSlug
	}

	game, err := g.queries.CreateGame(
		ctx,
		db.CreateGameParams{
			Name: input.Name,
			Slug: input.Slug,
			Description: pgtype.Text{
				String: input.Description,
				Valid:  true,
			},
			Publisher: pgtype.Text{
				String: input.Publisher,
				Valid:  true,
			},
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "games_slug_key" {
			return db.Game{}, ErrGameSlugTaken
		}
		return db.Game{}, fmt.Errorf("create game : %w", err)
	}
	return game, nil
}

func (g *GameService) RegisterRegionGame(ctx context.Context, gameID pgtype.UUID, input RegisterGameRegionInput) (db.GameRegion, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	if len(input.Name) < 2 || len(input.Name) > 255 {
		return db.GameRegion{}, ErrInvalidName
	}
	if len(input.Slug) < 2 || len(input.Slug) > 255 || !slugPattern.MatchString(input.Slug) {
		return db.GameRegion{}, ErrInvalidSlug
	}
	game, err := g.queries.GetGameById(ctx,gameID)
	if err != nil {
		return db.GameRegion{}, fmt.Errorf("get game by id: %w", err)
	}
	
	gameRegion, err := g.queries.CreateGameRegion(
		ctx, 
		db.CreateGameRegionParams{
			GameID: game.ID,
			Name:   input.Name,
			Slug:   input.Slug,
		},
	)
	if err != nil {
		return db.GameRegion{}, fmt.Errorf("create game region: %w", err)
	}
	return gameRegion, nil
}
func (g *GameService) GetGameByID(ctx context.Context, gameID pgtype.UUID) (db.Game, error) {
	game, err := g.queries.GetGameById(ctx, gameID)
	if err != nil {
		return db.Game{}, fmt.Errorf("get game by id: %w", err)
	}
	return game, nil
}

func (g *GameService) ListGames(ctx context.Context) ([]db.Game, error) {
	return g.queries.ListGames(ctx)
}
func (g *GameService) ListRegionByGame(ctx context.Context, gameID pgtype.UUID) ([]db.GameRegion, error) {
	return g.queries.ListAllRegionsByGame(ctx, gameID)
}

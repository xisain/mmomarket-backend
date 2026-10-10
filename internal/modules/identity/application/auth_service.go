package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"xisain/mmomarket-backend/internal/modules/identity/infrastructure"
	"xisain/mmomarket-backend/internal/platform/database/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvalidCredentials = errors.New("invalid Credential")
	ErrEmailTaken         = errors.New("email taken")
	ErrUnauthorized       = errors.New("unauthorized")
)

type AuthService struct {
	queries *db.Queries
}

func NewAuthService(q *db.Queries) *AuthService {
	return &AuthService{queries: q}
}
func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (db.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	displayName = strings.TrimSpace(displayName)
	_, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil {
		return db.User{}, ErrEmailTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, err
	}
	hash, err := infrastructure.HashPassword(password)
	if err != nil {
		return db.User{}, err
	}
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		DisplayName:  displayName,
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_email_unique" {
			return db.User{}, ErrEmailTaken
		}

		return db.User{}, err
	}

	return user, nil
}
func (s *AuthService) Login(ctx context.Context, email, password string) (db.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, "", err
		}
		return db.User{}, "", ErrInvalidCredentials
	}
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return db.User{}, "", err
	}
	rawToken := base64.RawURLEncoding.EncodeToString(token)
	tokenHash := HashSessionToken(rawToken)

	_, err = s.queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:           user.ID,
		SessionTokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(7 * 24 * time.Hour),
			Valid: true,
		},
	})
	if err != nil {
		return db.User{}, "", err
	}
	return user, rawToken, nil

}

func HashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])

}

func (s *AuthService) GetCurrentUser(
	ctx context.Context,
	token string,
) (db.User, error) {
	if token == "" {
		return db.User{}, ErrUnauthorized
	}
	session, err := s.queries.GetSessionByTokenHash(
		ctx,
		HashSessionToken(token),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrUnauthorized
		}
		return db.User{}, err
	}

	user, err := s.queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrUnauthorized
		}
		return db.User{}, err
	}

	if user.Status != "active" {
		return db.User{}, ErrUnauthorized
	}

	return user, nil
}

func (s *AuthService) Logout(
	ctx context.Context,
	token string,
) error {
	if token == "" {
		return nil
	}

	return s.queries.RevokeSession(
		ctx,
		HashSessionToken(token),
	)
}

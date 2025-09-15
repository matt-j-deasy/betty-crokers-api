package services

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/matt-j-deasy/betty-crokers-api/config"
	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
	"github.com/matt-j-deasy/betty-crokers-api/utils"
)

type AuthService struct {
	users   *repositories.UserRepository
	secret  []byte
	issuer  string
	exp     time.Duration
	nowFunc func() time.Time
}

func NewAuthService(repos *repositories.RepositoriesCollection, cfg config.Environment) *AuthService {
	return &AuthService{
		users:   repos.UserRepo,
		secret:  []byte(cfg.JWTSecret),
		issuer:  cfg.JWTIssuer,
		exp:     time.Duration(cfg.JWTExpMinutes) * time.Minute,
		nowFunc: time.Now,
	}
}

func (s *AuthService) Register(email, password string) (*models.User, error) {
	slog.Debug("Registering new user", slog.String("email", email))
	if _, err := s.users.FindByEmail(email); err == nil {
		slog.Debug("Email already in use", slog.String("email", email))
		return nil, fmt.Errorf("email already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // cost ~10
	if err != nil {
		slog.Debug("Password hashing failed", slog.String("email", email), slog.String("err", err.Error()))
		return nil, err
	}

	u := &models.User{
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.users.Create(u); err != nil {
		slog.Debug("User creation failed", slog.String("email", email), slog.String("err", err.Error()))
		return nil, err
	}
	return u, nil
}

func (s *AuthService) Login(email, password string) (string, time.Time, *models.User, error) {
	u, err := s.users.FindByEmail(email)
	if err != nil {
		if utils.IsNotFound(err) {
			return "", time.Time{}, nil, errors.New("invalid credentials")
		}
		return "", time.Time{}, nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", time.Time{}, nil, errors.New("invalid credentials")
	}

	exp := s.nowFunc().Add(s.exp)
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", u.ID),
		"email": u.Email,
		"iss":   s.issuer,
		"iat":   s.nowFunc().Unix(),
		"exp":   exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		slog.Error("sign token failed", "err", err)
		return "", time.Time{}, nil, err
	}
	return signed, exp, u, nil
}

package auth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/theggv/kf2-stats-backend/pkg/common/config"
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type AuthService struct {
	db *sql.DB

	userService     *users.UserService
	steamApiService *steamapi.SteamApiUserService
}

func NewAuthService(db *sql.DB) *AuthService {
	service := AuthService{
		db: db,
	}

	return &service
}

func (s *AuthService) Inject(
	userService *users.UserService,
	steamApiService *steamapi.SteamApiUserService,
) {
	s.userService = userService
	s.steamApiService = steamApiService
}

func (s *AuthService) Login(
	req steamapi.ValidateOpenIdRequest,
) (*Token, error) {
	steamData, err := s.steamApiService.ValidateOpenId(req)
	if err != nil {
		return nil, err
	}

	userId, err := s.userService.FindCreateFind(users.CreateUserRequest{
		AuthId:   steamData.SteamId,
		AuthType: models.Steam,
		Name:     steamData.Name,
	})
	if err != nil {
		return nil, err
	}

	tokens, err := s.generateTokens(&models.TokenPayload{
		UserId:     userId,
		Name:       steamData.Name,
		SteamId:    steamData.SteamId,
		Avatar:     steamData.Avatar,
		ProfileUrl: steamData.ProfileUrl,
	})
	if err != nil {
		return nil, err
	}

	{
		tx, err := s.db.Begin()
		if err != nil {
			return nil, err
		}

		defer tx.Rollback()

		// save token
		_, err = tx.Exec(
			`INSERT INTO users_token (user_id, token) VALUES (?, ?)`,
			userId, tokens.RefreshToken,
		)
		if err != nil {
			return nil, err
		}

		// update steam data
		_, err = tx.Exec(`
			INSERT INTO users_steam_data 
				(user_id, steam_id, name, avatar, profile_url) 
			VALUES (?, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE
				name = ?, avatar = ?, profile_url = ?, updated_at = CURRENT_TIMESTAMP
			`,
			userId, steamData.SteamId, steamData.Name, steamData.Avatar, steamData.ProfileUrl,
			steamData.Name, steamData.Avatar, steamData.ProfileUrl,
		)
		if err != nil {
			return nil, err
		}

		tx.Commit()
	}

	return tokens, nil
}

func (s *AuthService) Refresh(refreshToken string) (*Token, error) {
	userId, err := s.findToken(refreshToken)
	if err != nil {
		return nil, err
	}

	payload, err := util.ValidateToken(refreshToken, config.Instance.JwtRefreshSecretKey)
	if err != nil {
		return nil, err
	}

	payloadUserId, ok := payload.(float64)
	if !ok || userId != int(payloadUserId) {
		return nil, errors.New("invalid token")
	}

	tokenPayload, err := s.getSteamDataFromDB(userId)
	if err != nil {
		return nil, err
	}

	// try to update steam info
	summary, err := s.steamApiService.GetUserSummary([]string{tokenPayload.SteamId})
	if err == nil && len(summary) == 1 {
		tokenPayload.Name = summary[0].Name
		tokenPayload.Avatar = summary[0].Avatar
		tokenPayload.ProfileUrl = summary[0].ProfileUrl

		s.db.Exec(`
			UPDATE users_steam_data 
			SET name = ?, avatar = ?, profile_url = ? WHERE user_id = ?`,
			tokenPayload.Name, tokenPayload.Avatar, tokenPayload.ProfileUrl, tokenPayload.UserId,
		)
	}

	tokens, err := s.generateTokens(tokenPayload)
	if err != nil {
		return nil, err
	}

	err = s.updateToken(refreshToken, tokens.RefreshToken)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	return s.deleteToken(refreshToken)
}

func (s *AuthService) GetSteamData(id int) (*models.TokenPayload, error) {
	return s.getSteamDataFromDB(id)
}

func (s *AuthService) getSteamDataFromDB(userId int) (*models.TokenPayload, error) {
	stmt := `SELECT steam_id, name, avatar, profile_url FROM users_steam_data WHERE user_id = ?`
	row := s.db.QueryRow(stmt, userId)

	payload := models.TokenPayload{UserId: userId}
	err := row.Scan(&payload.SteamId, &payload.Name, &payload.Avatar, &payload.ProfileUrl)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

func (s *AuthService) findToken(refreshToken string) (int, error) {
	stmt := `SELECT user_id FROM users_token WHERE token = ?`
	row := s.db.QueryRow(stmt, refreshToken)

	var userId int
	err := row.Scan(&userId)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (s *AuthService) updateToken(oldToken string, newToken string) error {
	stmt := `UPDATE users_token SET token = ? WHERE token = ?`
	_, err := s.db.Exec(stmt, newToken, oldToken)
	return err
}

func (s *AuthService) deleteToken(refreshToken string) error {
	stmt := `DELETE FROM users_token WHERE token = ?`
	_, err := s.db.Exec(stmt, refreshToken)
	return err
}

func (s *AuthService) generateTokens(payload *models.TokenPayload) (*Token, error) {
	config := config.Instance

	accessTokenString, err := signToken(
		payload, config.JwtAccessSecretKey, config.JwtAccessExpiresIn,
	)
	if err != nil {
		return nil, err
	}

	refreshTokenString, err := signToken(
		payload.UserId, config.JwtRefreshSecretKey, config.JwtRefreshExpiresIn,
	)
	if err != nil {
		return nil, err
	}

	return &Token{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

func signToken(payload any, key string, expiresIn string) (string, error) {
	duration, err := time.ParseDuration(expiresIn)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"payload": payload,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(duration).Unix(),
	})

	return token.SignedString([]byte(key))
}

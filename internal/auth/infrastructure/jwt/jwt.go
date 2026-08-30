package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const defaultTokenTTL = 24 * time.Hour

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

type Service struct {
	secret []byte
	ttl    time.Duration
}

func New(secret string, ttl ...time.Duration) *Service {
	tokenTTL := defaultTokenTTL
	if len(ttl) > 0 && ttl[0] != 0 {
		tokenTTL = ttl[0]
	}
	return &Service{secret: []byte(secret), ttl: tokenTTL}
}

func (s *Service) Generate(userID, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Iat:    now.Unix(),
		Exp:    now.Add(s.ttl).Unix(),
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	unsigned := encodeSegment(headerJSON) + "." + encodeSegment(claimsJSON)
	signature := s.sign(unsigned)
	return unsigned + "." + signature, nil
}

func (s *Service) Validate(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSignature := s.sign(unsigned)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return Claims{}, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, ErrInvalidToken
	}

	if claims.UserID == "" || claims.Email == "" || claims.Role == "" {
		return Claims{}, ErrInvalidToken
	}
	if time.Now().Unix() > claims.Exp {
		return Claims{}, ErrExpiredToken
	}

	return claims, nil
}

func (s *Service) sign(unsigned string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(unsigned))
	return encodeSegment(mac.Sum(nil))
}

func encodeSegment(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

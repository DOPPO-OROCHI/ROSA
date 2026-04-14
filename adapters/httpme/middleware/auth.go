package middleware

import (
	"TheWar/internal/domain/player"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ФУЛЛ СПИЗЖЕННАЯ У ИИ СХЕМА АВТОРИЗАЦИИ
var (
	ErrInitDataMissing      = errors.New("missing init data")
	ErrInitDataBadFormat    = errors.New("bad init data")
	ErrInitDataBadSignature = errors.New("invalid signature")
	ErrInitDataExpired      = errors.New("auth expired")
	ErrInitDataBadUser      = errors.New("bad user")
)

type ctxKey string

const AuthKey ctxKey = "auth_user"

type TokenStore struct {
	db  *gorm.DB
	ttl time.Duration
}
type tgInitUser struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	LanguageCode string `json:"language_code"`
}
type session struct {
	UserID    uint
	TGID      int
	ExpiredAt time.Time
}
type AuthUser struct {
	UserID uint
	TGID   int
}

func AuthMiddleware(store *TokenStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/auth/telegram" || r.URL.Path == "/auth/dev" || r.URL.Path == "/healthz" { //<-исправить после тестов!
				next.ServeHTTP(w, r)
				return
			}
			c, err := r.Cookie("session")
			if err != nil || c.Value == "" {
				log.Printf("cookie err=%v value=%v", err, c)
				WriteErr(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			sess, ok := store.Validate(c.Value)
			if !ok {
				WriteErr(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), AuthKey, AuthUser{
				UserID: sess.UserID,
				TGID:   sess.TGID,
			})
			log.Printf("auth ok user = %d, path = %s", sess.UserID, r.URL.Path)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ValidateTelegramInitData(initDataRaw, botToken string, maxAge time.Duration) (tgInitUser, error) {
	if initDataRaw == "" {
		return tgInitUser{}, ErrInitDataMissing
	}
	values, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return tgInitUser{}, ErrInitDataBadFormat
	}
	hashHex := values.Get("hash")
	if hashHex == "" {
		return tgInitUser{}, ErrInitDataBadSignature
	}
	values.Del("hash")
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+values.Get(k))
	}
	dataCheckString := strings.Join(parts, "\n")
	secretMac := hmac.New(sha256.New, []byte("WebAppData"))
	secretMac.Write([]byte(botToken))
	secret := secretMac.Sum(nil)
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(dataCheckString))
	expected := m.Sum(nil)
	got, err := hex.DecodeString(hashHex)
	if err != nil {
		return tgInitUser{}, ErrInitDataBadSignature
	}
	if !hmac.Equal(expected, got) {
		return tgInitUser{}, ErrInitDataBadSignature
	}
	authDataStr := values.Get("auth_date")
	if authDataStr == "" {
		return tgInitUser{}, ErrInitDataBadFormat
	}
	authDate, err := strconv.ParseInt(authDataStr, 10, 64)
	if err != nil || authDate <= 0 {
		return tgInitUser{}, ErrInitDataBadFormat
	}
	if time.Since(time.Unix(authDate, 0)) > maxAge {
		return tgInitUser{}, ErrInitDataExpired
	}
	userJSON := values.Get("user")
	if userJSON == "" {
		return tgInitUser{}, ErrInitDataBadUser
	}
	var u tgInitUser
	if err := json.Unmarshal([]byte(userJSON), &u); err != nil {
		return tgInitUser{}, ErrInitDataBadUser
	}
	return u, nil
}

func (s *TokenStore) Validate(token string) (session, bool) {
	if s == nil || s.db == nil || token == "" {
		return session{}, false
	}
	now := time.Now()
	var row player.AuthSession
	if err := s.db.Where("token_hash = ? AND expires_at > ?", tokenHash(token), now).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return session{}, false
		}
		log.Printf("session validate db err: %v", err)
		return session{}, false
	}
	return session{
		UserID:    row.UserID,
		TGID:      int(row.TGID),
		ExpiredAt: row.ExpiresAt,
	}, true
}

func (s *TokenStore) Issue(userID uint, tgID int) (token string, exp time.Time, err error) {
	if s == nil || s.db == nil {
		return "", time.Time{}, fmt.Errorf("token store db is nil")
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	exp = time.Now().Add(s.ttl)
	row := player.AuthSession{
		TokenHash: tokenHash(token),
		UserID:    userID,
		TGID:      int64(tgID),
		ExpiresAt: exp,
	}
	if err := s.db.Create(&row).Error; err != nil {
		return "", time.Time{}, err
	}

	_ = s.db.Where("expires_at <= ?", time.Now()).Delete(&player.AuthSession{}).Error

	return token, exp, nil
}

func NewTokenStore(db *gorm.DB, ttl time.Duration) *TokenStore {
	return &TokenStore{
		db:  db,
		ttl: ttl,
	}
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

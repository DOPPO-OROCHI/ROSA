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
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

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
	mu   sync.RWMutex
	data map[string]session
	ttl  time.Duration
}
type tgInitUser struct {
	ID int `json:"id"`
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

func AuthMiddleware(db *gorm.DB, store *TokenStore) func(http.Handler) http.Handler {
	botToken := os.Getenv("BOT_API")
	if botToken == "" {
		panic("BOT_API empty")
	}
	initDataMaxAge := 5 * time.Minute
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ah := r.Header.Get("Authorization"); ah != "" {
				const prefix = "Bearer "
				if strings.HasPrefix(ah, prefix) {
					tok := strings.TrimSpace(strings.TrimPrefix(ah, prefix))
					if tok != "" {
						if sess, ok := store.Validate(tok); ok {
							ctx := context.WithValue(r.Context(), AuthKey, AuthUser{
								UserID: sess.UserID,
								TGID:   sess.TGID,
							})
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}
				}
			}
			initData := r.Header.Get("X-Telegram-Init-Data")
			tgID, err := validateTelegramInitData(initData, botToken, initDataMaxAge)
			if err != nil {
				WriteErr(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			var dbUser player.TelegramUser
			if err := db.Where("tg_id = ?", tgID).First(&dbUser).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					WriteErr(w, http.StatusUnauthorized, "unauthorized")
					return
				}
				WriteErr(w, http.StatusInternalServerError, "something went wrong")
				return
			}
			token, _, terr := store.Issue(dbUser.ID, tgID)
			if terr == nil {
				// Клиент один раз забирает и дальше ходит по Bearer.
				w.Header().Set("X-Session-Token", token)
			}
			ctx := context.WithValue(r.Context(), AuthKey, AuthUser{
				UserID: dbUser.ID,
				TGID:   tgID,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func validateTelegramInitData(initDataRaw, botToken string, maxAge time.Duration) (int, error) {
	if initDataRaw == "" {
		return 0, ErrInitDataMissing
	}

	values, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return 0, ErrInitDataBadFormat
	}

	hashHex := values.Get("hash")
	if hashHex == "" {
		return 0, ErrInitDataBadSignature
	}
	values.Del("hash")

	// 1) data_check_string: sorted key=value joined with "\n"
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

	// 2) secret = HMAC_SHA256(key="WebAppData", message=botToken)
	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(botToken))
	secret := secretMAC.Sum(nil)

	// 3) expected = HMAC_SHA256(key=secret, message=dataCheckString)
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(dataCheckString))
	expected := m.Sum(nil)

	// 4) compare bytes (hash is hex)
	got, err := hex.DecodeString(hashHex)
	if err != nil {
		return 0, ErrInitDataBadSignature
	}
	if !hmac.Equal(expected, got) {
		return 0, ErrInitDataBadSignature
	}

	// TTL / anti-replay window
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return 0, ErrInitDataBadFormat
	}
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil || authDate <= 0 {
		return 0, ErrInitDataBadFormat
	}
	if time.Since(time.Unix(authDate, 0)) > maxAge {
		return 0, ErrInitDataExpired
	}

	userJSON := values.Get("user")
	if userJSON == "" {
		return 0, ErrInitDataBadUser
	}
	var u tgInitUser
	if err := json.Unmarshal([]byte(userJSON), &u); err != nil || u.ID == 0 {
		return 0, ErrInitDataBadUser
	}
	return u.ID, nil
}

func (s *TokenStore) Validate(token string) (session, bool) {
	s.mu.RLock()
	sess, ok := s.data[token]
	s.mu.RUnlock()
	if !ok {
		return session{}, false
	}
	if time.Now().After(sess.ExpiredAt) {
		s.mu.Lock()
		delete(s.data, token)
		s.mu.Unlock()
		return session{}, false
	}
	return sess, true
}

func (s *TokenStore) Issue(userID uint, tgID int) (token string, exp time.Time, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	exp = time.Now().Add(s.ttl)
	s.mu.Lock()
	s.data[token] = session{UserID: userID, TGID: tgID, ExpiredAt: exp}
	return token, exp, nil
}

func NewTokenStore(ttl time.Duration) *TokenStore {
	return &TokenStore{
		data: make(map[string]session),
		ttl:  ttl,
	}
}

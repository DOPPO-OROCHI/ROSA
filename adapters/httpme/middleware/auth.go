package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
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

func ValidateTelegramInitData(initDataRaw, botToken string, maxAge time.Duration) (int, error) {
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
		return 0, ErrInitDataBadSignature
	}
	if !hmac.Equal(expected, got) {
		return 0, ErrInitDataBadSignature
	}
	authDataStr := values.Get("auth_date")
	if authDataStr == "" {
		return 0, ErrInitDataBadFormat
	}
	authDate, err := strconv.ParseInt(authDataStr, 10, 64)
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
	if err := json.Unmarshal([]byte(userJSON), &u); err != nil {
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
	defer s.mu.Unlock()
	s.data[token] = session{UserID: userID, TGID: tgID, ExpiredAt: exp}
	return token, exp, nil
}

func NewTokenStore(ttl time.Duration) *TokenStore {
	return &TokenStore{
		data: make(map[string]session),
		ttl:  ttl,
	}
}

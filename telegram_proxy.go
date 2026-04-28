package main

import (
	"TheWar/internal/infra/db"
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

/*А вот это пиздец новая стезя прям для меня. Это подключение к прокси епта... И тут ща будет конспект...
 */
func newTelegramHTTPClientFromEnv() (*http.Client, error) {
	IP := db.GoDotEnvVariable("PROXY_IP")
	if IP == "" {
		return nil, errors.New("nil ip")
	}
	PORT := db.GoDotEnvVariable("PROXY_PORT")
	if PORT == "" {
		return nil, errors.New("nil port")
	}
	LOGIN := db.GoDotEnvVariable("PROXY_LOGIN")
	if LOGIN == "" {
		return nil, errors.New("nil login")
	}
	PASSWORD := db.GoDotEnvVariable("PROXY_PSW")
	if PASSWORD == "" {
		return nil, errors.New("nil password")
	}
	addr := net.JoinHostPort(IP, PORT)
	auth := &proxy.Auth{
		User:     LOGIN,
		Password: PASSWORD,
	}
	dialer, err := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		},
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 65 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
	}
	return client, nil
}

package main

import (
	"TheWar/adapters"
	"TheWar/adapters/httpme/handlers"
	"TheWar/adapters/httpme/middleware"
	"TheWar/adapters/telegram"
	"TheWar/internal/applycation"
	"TheWar/internal/bootstrap"
	"TheWar/internal/domain/game"
	"TheWar/internal/infra/cache"
	"TheWar/internal/infra/db"
	"TheWar/internal/infra/repository"
	"TheWar/internal/transport"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.Init(); err != nil {
		log.Fatalf("bootstrap init failed: %v", err)
	}
	var rc cache.ResolverCache
	if err := rc.Reload(db.DB); err != nil {
		log.Fatalf("resolver cache reload failed: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	store := middleware.NewTokenStore(24 * time.Hour)
	hub := transport.NewHub()
	app := adapters.App{
		CreateMatch: handlers.NewCreateMatchHandler(handlers.CreateMatchHandlerDeps{DB: db.DB}),
		GetMatch:    handlers.NewGetMatchHandler(handlers.GetMatchHandlerDeps{DB: db.DB}),
		MatchesList: handlers.NewMathesListHandler(handlers.MathesListHandlerDeps{DB: db.DB}),
		ApplyAction: handlers.NewApplyActionHandler(handlers.ApplyActionHandlerDeps{DB: db.DB, Resolvers: mustBeResolvers(&rc)}),
		GetMe:       handlers.NewGetMeHandler(handlers.GetMeHandler{DB: db.DB}),
		GetDeck:     handlers.NewGetDeckHandler(handlers.DeckHandlerDeps{DB: db.DB}),
		SaveDeck:    handlers.NewSaveDeckHandler(handlers.DeckHandlerDeps{DB: db.DB}),
		CardsList:   handlers.NewCardsListHandler(dtoLikeDepsForCards(db.DB)),
		HeroesList:  handlers.NewHeroesListHandler(handlers.HeroListHandler{DB: db.DB}),
		SelectHero:  handlers.NewSelectedHeroHandler(db.DB),
		StreamMatch: handlers.NewStreamMatchHandler(handlers.StreamMatchDeps{Hub: hub}),
	}
	mux := adapters.NewMux(app)
	httpHandler := middleware.AuthMiddleware(store)(mux)

	srv := &http.Server{
		Addr:         ":1234",
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	errCh := make(chan error, 2)
	go func() {
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				now := time.Now().Unix()
				ids, err := repository.ListExpiredMatches(db.DB, now, 50)
				if err != nil {
					continue
				}
				for _, id := range ids {
					st, changed, err := applycation.ApplyTimeOutToMatchTX(db.DB, id)
					if err != nil || !changed || st == nil {
						continue
					}
					handlers.PublishMatchToSSE(hub, st)
				}
			}
		}
	}()
	go func() {
		if err := runTelegramBot(ctx); err != nil {
			errCh <- err
		}
	}()
	select {
	case <-ctx.Done():
	case err := <-errCh:
		log.Printf("fatal error: %v", err)
		stop()
	}
	shutDownCtx, cansel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cansel()
	_ = srv.Shutdown(shutDownCtx)
	log.Printf("shutdown domplete")
}

func runTelegramBot(ctx context.Context) error {
	api := db.GoDotEnvVariable("BOT_API")
	bot, err := tgbotapi.NewBotAPI(api)
	if err != nil {
		return err
	}
	bot.Debug = true
	log.Printf("ура подключились к боту: %v", &bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			log.Printf("telegram bot shutting down")
			return nil
		case update, ok := <-updates:
			if !ok {
				return errors.New("telegram update ch closed")
			}
			if update.Message == nil {
				continue
			}
			if err := telegram.AddNewUser(db.DB, update); err != nil {
				log.Printf("add user error: %v", err)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "ДОБРО ПОЖАЛОВАТЬ В ЛУЧШУЮ КАРТОЧНУЮ ИГРУ")
			_, _ = bot.Send(msg)
		}
	}
}

func mustBeResolvers(rc *cache.ResolverCache) game.Resolvers {
	res, ok := rc.Get()
	if !ok {
		panic("resolver cach not init")
	}
	return res
}

func dtoLikeDepsForCards(dbConn *gorm.DB) handlers.CardListHandlerDeps {
	return handlers.CardListHandlerDeps{DB: dbConn}
}

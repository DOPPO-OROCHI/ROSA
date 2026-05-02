package main

import (
	"TheWar/adapters"
	"TheWar/adapters/httpme/handlers"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/applycation"
	"TheWar/internal/bootstrap"
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/queue"
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
	store := middleware.NewTokenStore(db.DB, 72*time.Hour)
	hub := transport.NewHub()
	matchQueue := queue.NewQueue()
	app := adapters.App{
		CreateMatch:  handlers.NewCreateMatchHandler(handlers.CreateMatchHandlerDeps{DB: db.DB}),
		GetMatch:     handlers.NewGetMatchHandler(handlers.GetMatchHandlerDeps{DB: db.DB}),
		MatchesList:  handlers.NewMathesListHandler(handlers.MathesListHandlerDeps{DB: db.DB}),
		ApplyAction:  handlers.NewApplyActionHandler(handlers.ApplyActionHandlerDeps{DB: db.DB, Resolvers: mustBeResolvers(&rc), Hub: hub}),
		GetMe:        handlers.NewGetMeHandler(handlers.GetMeHandler{DB: db.DB}),
		GetDeck:      handlers.NewGetDeckHandler(handlers.DeckHandlerDeps{DB: db.DB}),
		SaveDeck:     handlers.NewSaveDeckHandler(handlers.DeckHandlerDeps{DB: db.DB}),
		CardsList:    handlers.NewCardsListHandler(dtoLikeDepsForCards(db.DB)),
		HeroesList:   handlers.NewHeroesListHandler(handlers.HeroListHandler{DB: db.DB}),
		SelectHero:   handlers.NewSelectedHeroHandler(db.DB),
		StreamMatch:  handlers.NewStreamMatchHandler(handlers.StreamMatchDeps{Hub: hub, Store: store}),
		AuthTelegram: handlers.NewAuthTelegramHandler(handlers.AuthTelegramDeps{DB: db.DB, Store: store}),
		AuthDev:      handlers.NewAuthDevHandler(handlers.AuthDevDeps{DB: db.DB, Store: store}),
		JoinQueue:    handlers.JoinQueue(handlers.NewJoinHandler(handlers.JoinQueueHandlerDeps{DB: db.DB, Queue: matchQueue})),
		LeaveQueue:   handlers.LeaveQueue(handlers.NewLeaveQueueHandler(handlers.LeaveQueueHandlerDeps{Queue: matchQueue})),
		QueueStatus:  handlers.QueueStatus(handlers.NewQueueStatusHandler(handlers.QueueStatusHandlerDeps{DB: db.DB, Queue: matchQueue})),
		AcceptQueue:  handlers.AcceptQueue(handlers.NewAcceptQueueHandler(handlers.AcceptQueueHandlerDeps{DB: db.DB, Queue: matchQueue})),
		DeclineQueue: handlers.DeclineQueue(handlers.NewDeclineQueueHandler(handlers.DeclineQueueHandlerDeps{Queue: matchQueue})),
		ReadyMatch:   handlers.NewReadyMatchHandler(handlers.ReadyMatchHandlersDeps{DB: db.DB, Hub: hub}),
	}
	mux := adapters.NewMux(app)
	httpHandler := middleware.AuthMiddleware(store)(mux)
	srv := &http.Server{
		Addr:         ":1234",
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0,
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
				matchQueue.CleanupExpired()
				now := time.Now().Unix()
				ids, err := repository.ListExpiredMatches(db.DB, now, 50)
				if err != nil {
					log.Printf("ecpired scan err: %v", err)
					log.Printf("ecpired ids: %v", ids)
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
		errCh <- srv.ListenAndServe()
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err := runTelegramBot(ctx); err != nil {
				log.Printf("telegram bot err: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()
	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("fatal err : %v", err)
		}
		stop()
	}
	shutDownCtx, cansel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cansel()
	_ = srv.Shutdown(shutDownCtx)
	log.Printf("shutdown domplete")
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

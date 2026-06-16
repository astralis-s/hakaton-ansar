// Command bot — автономный Telegram-бот «Амана»: принимает сообщения заказчиков,
// проводит первичную регистрацию (ФИО + телефон) и связывает их с менеджером
// через встроенный веб-инбокс. Никак не зависит от основного приложения CRM.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/app"
	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/config"
	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/store"
	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/telegram"
	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/webui"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	config.LoadDotenv(".env") // best-effort: локальная разработка

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	st, err := store.Open(cfg.DataFile)
	if err != nil {
		return fmt.Errorf("открыть хранилище %q: %w", cfg.DataFile, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tg := telegram.New(cfg.BotToken)
	me, err := tg.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("проверка токена бота (getMe): %w", err)
	}
	log.Info("telegram бот подключён", "username", me.Username, "id", me.ID)

	bridge := app.NewBridge(st, tg, log)

	// Веб-инбокс менеджера.
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           webui.NewServer(bridge, log, cfg.ManagerPassword).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Info("веб-инбокс менеджера запущен", "addr", cfg.HTTPAddr, "auth", cfg.ManagerPassword != "")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http сервер остановился", "error", err)
			stop()
		}
	}()

	// Long polling Telegram в фоне.
	poller := telegram.NewPoller(tg, log, int(cfg.PollTimeout.Seconds()))
	go poller.Run(ctx, bridge.HandleUpdate)

	<-ctx.Done()
	log.Info("завершение работы…")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	return nil
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-huma-test/db"
	"go-huma-test/handler"
	"go-huma-test/model"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "embed"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema/schema.sql
var schema string

func initDB(dbPath string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	params := []string{
		"PRAGMA busy_timeout = 5000;", // ロックされている場合最大5秒待つ
		"PRAGMA journal_mode = WAL;",  // 読み取りは複数同時に可能だが書き込みは１つだけ。SQLiteをWebAPIで使用する場合はほぼ必須
		"PRAGMA foreign_keys = ON;",   // 外部キー制約を有効化（将来のために）
	}
	for _, p := range params {
		if _, err := sqlDB.Exec(p); err != nil {
			return nil, err
		}
	}

	sqlDB.SetMaxOpenConns(1) // 同時に開ける最大コネクション数
	sqlDB.SetMaxIdleConns(1) // アイドル状態のコネクション数

	if _, err := sqlDB.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return sqlDB, nil
}

func LoggingMiddleware(ctx huma.Context, next func(huma.Context)) {
	fmt.Printf("[%s] %s\n", ctx.Method(), ctx.URL().Path)
	next(ctx)
}

func AuthMiddleware(ctx huma.Context, next func(huma.Context)) {
	// 認証チェック
	token := ctx.Header("Authorization")
	if token == "" {
		huma.WriteErr(huma.NewAPI(huma.Config{}, nil), ctx, http.StatusUnauthorized,
			"Authorization header required")
		return
	}

	next(ctx)
}

func main() {
	sqlDB, err := initDB("./todos.db")
	if err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	queries, err := db.Prepare(context.Background(), sqlDB)
	if err != nil {
		slog.Error("failed to prepare database", "err", err)
		os.Exit(1)
	}
	handler := handler.NewTodoHandler(queries, sqlDB)

	cli := humacli.New(func(h humacli.Hooks, o *model.Options) {
		mux := http.NewServeMux()

		config := huma.DefaultConfig("Todo API", "1.0.0")
		config.Info.Description = "SQLite + sqlc + Humaを使ったシンプルなTodo API"
		config.CreateHooks = []func(huma.Config) huma.Config{}
		api := humago.New(mux, config)

		// ミドルウェア設定
		api.UseMiddleware(LoggingMiddleware)
		api.UseMiddleware(AuthMiddleware)

		huma.Register(api, huma.Operation{
			OperationID: "list-todos",
			Method:      http.MethodGet,
			Path:        "/todos",
			Summary:     "Todo一覧取得",
			Description: "すべてのTodoを取得",
			Tags:        []string{"todos"},
		}, handler.ListTodos)

		huma.Register(api, huma.Operation{
			OperationID: "get-todo",
			Method:      http.MethodGet,
			Path:        "/todos/{id}",
			Summary:     "Todo取得",
			Description: "指定したIDのTodoを取得します。",
			Tags:        []string{"todos"},
		}, handler.GetTodo)

		huma.Register(api, huma.Operation{
			OperationID:   "create-todo",
			Method:        http.MethodPost,
			Path:          "/todos",
			Summary:       "Todo作成",
			Description:   "新しいTodoを作成します。",
			Tags:          []string{"todos"},
			DefaultStatus: http.StatusCreated,
		}, handler.CreateTodo)

		huma.Register(api, huma.Operation{
			OperationID: "update-todo",
			Method:      http.MethodPut,
			Path:        "/todos/{id}",
			Summary:     "Todo更新",
			Description: "指定したIDのTodoを更新します。",
			Tags:        []string{"todos"},
		}, handler.UpdateTodo)

		huma.Register(api, huma.Operation{
			OperationID: "delete-todo",
			Method:      http.MethodDelete,
			Path:        "/todos/{id}",
			Summary:     "Todo削除",
			Description: "指定したIDのTodoを削除します。",
			Tags:        []string{"todos"},
		}, handler.DeleteTodo)

		huma.Register(api, huma.Operation{
			OperationID: "toggle-todo",
			Method:      http.MethodPost,
			Path:        "/todos/{id}/toggle",
			Summary:     "Todo完了状態切り替え",
			Description: "指定したIDのTodoの完了状態を切り替えます。",
			Tags:        []string{"todos"},
		}, handler.ToggleTodo)

		srv := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", o.Host, o.Port),
			Handler: mux,
		}

		h.OnStart(func() {
			addr := fmt.Sprintf("%s:%d", o.Host, o.Port)
			log.Printf("🚀 Todo API Server starting on http://%s", addr)
			log.Printf("📚 API Documentation: http://%s/docs", addr)
			log.Printf("📚 Get OpenAPI File: http://%s/openapi.yaml", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %s\n", err)
			}
		})

		h.OnStop(func() {
			log.Println("Shutting down server...")

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("Server shutdown error: %s\n", err)
				os.Exit(1)
			}

			log.Println("Server stopped gracefully")
		})
	})

	cli.Run()
}

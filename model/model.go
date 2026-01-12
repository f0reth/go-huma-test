// Package model はTodo管理APIのリクエスト/レスポンスモデルを定義する。
// このパッケージはAPIの入出力データ構造を提供し、
// バリデーションルールとドキュメント情報を含む。
package model

// Options はサーバーの起動オプションを表す構造体
type Options struct {
	Port int    `doc:"Port to listen on." short:"p" default:"8888"`
	Host string `doc:"Hostname to listen on." default:"localhost"`
}

// TodoResponse はTodoのレスポンスを表す構造体
type TodoResponse struct {
	ID          int64   `json:"id" example:"1" doc:"TodoのID"`
	Title       string  `json:"title" example:"買い物" doc:"Todoのタイトル"`
	Description *string `json:"description,omitempty" example:"牛乳を買う" doc:"Todoの詳細説明"`
	Completed   bool    `json:"completed" example:"false" doc:"完了状態"`
	CreatedAt   string  `json:"created_at" example:"2024-01-01T00:00:00Z" doc:"作成日時"`
	UpdatedAt   string  `json:"updated_at" example:"2024-01-01T00:00:00Z" doc:"更新日時"`
}

// ListTodosInput はTodoリスト取得のリクエストパラメータを表す構造体
type ListTodosInput struct {
	Completed bool `query:"completed" doc:"完了状態でフィルタリング"`
}

// ListTodosOutput はTodoリスト取得のレスポンスを表す構造体
type ListTodosOutput struct {
	Body struct {
		Todos []TodoResponse `json:"todos" doc:"Todoのリスト"`
	}
}

// GetTodoInput はTodo取得のリクエストパラメータを表す構造体
type GetTodoInput struct {
	ID int64 `path:"id" doc:"TodoのID"`
}

// GetTodoOutput はTodo取得のレスポンスを表す構造体
type GetTodoOutput struct {
	Body TodoResponse
}

// CreateTodoInput はTodo作成のリクエストボディを表す構造体
type CreateTodoInput struct {
	Body struct {
		Title       string  `json:"title" minLength:"1" maxLength:"200" doc:"Todoのタイトル"`
		Description *string `json:"description,omitempty" maxLength:"1000" doc:"Todoの詳細説明"`
	}
}

// CreateTodoOutput はTodo作成のレスポンスを表す構造体
type CreateTodoOutput struct {
	Body TodoResponse
}

// UpdateTodoInput はTodo更新のリクエストパラメータとボディを表す構造体
type UpdateTodoInput struct {
	ID   int64 `path:"id" doc:"TodoのID"`
	Body struct {
		Title       string  `json:"title" minLength:"1" maxLength:"200" doc:"Todoのタイトル"`
		Description *string `json:"description,omitempty" maxLength:"1000" doc:"Todoの詳細説明"`
		Completed   bool    `json:"completed" doc:"完了状態"`
	}
}

// UpdateTodoOutput はTodo更新のレスポンスを表す構造体
type UpdateTodoOutput struct {
	Body TodoResponse
}

// DeleteTodoInput はTodo削除のリクエストパラメータを表す構造体
type DeleteTodoInput struct {
	ID int64 `path:"id" doc:"TodoのID"`
}

// DeleteTodoOutput はTodo削除のレスポンスを表す構造体
type DeleteTodoOutput struct {
	Body struct {
		Message string `json:"message" example:"Todo deleted successfully" doc:"削除結果メッセージ"`
	}
}

// ToggleTodoInput はTodo完了状態トグルのリクエストパラメータを表す構造体
type ToggleTodoInput struct {
	ID int64 `path:"id" doc:"TodoのID"`
}

// ToggleTodoOutput はTodo完了状態トグルのレスポンスを表す構造体
type ToggleTodoOutput struct {
	Body TodoResponse
}

package model

import "time"

type User struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Paper struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"user_id"`
	Title       string    `json:"title"`
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	ParseStatus string    `json:"parse_status"`
	ParseError  string    `json:"parse_error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ParseJob struct {
	ID         uint64     `json:"id"`
	PaperID    uint64     `json:"paper_id"`
	Status     string     `json:"status"`
	Progress   uint8      `json:"progress"`
	RetryCount uint8      `json:"retry_count"`
	MaxRetries uint8      `json:"max_retries"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	ErrorMsg   string     `json:"error_message,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type PaperTranslation struct {
	ID             uint64    `json:"id"`
	PaperID        uint64    `json:"paper_id"`
	TargetLanguage string    `json:"target_language"`
	Status         string    `json:"status"`
	Content        string    `json:"content,omitempty"`
	ErrorMsg       string    `json:"error_message,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

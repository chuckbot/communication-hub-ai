package domain

import (
	"context"
	"time"
)

// Email representa la entidad básica de entrada del sistema
type Email struct {
	ID        string    `json:"id"`
	RawBody   string    `json:"raw_body"`
	Sender    string    `json:"sender"`
	Subject   string    `json:"subject"`
	Timestamp time.Time `json:"timestamp"`
}

// LLMMessage representa la estructura para Short-term Memory (Contexto de chat)
type LLMMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// LLMClient define el contrato para cualquier modelo de lenguaje
// Esto permite cumplir con el requerimiento de "trade-offs entre latencia y calidad"
type LLMClient interface {
    Chat(ctx context.Context, system string, user string) (string, error) 
    // Asegúrate de que aquí diga Chat si vas a usar Chat en la implementación
}

// Repository define cómo persistimos los datos (Supabase/Postgres)
// Cumple con el requerimiento de "Integración con bases de datos y sistemas internos"
type Repository interface {
	SaveEmail(ctx context.Context, email *Email) error
	UpdateStatus(ctx context.Context, emailID string, status string) error
	SaveEvent(ctx context.Context, event map[string]interface{}) error
}

// Queue define el contrato para nuestra Event-Driven Architecture (Redis)
type Queue interface {
	Publish(ctx context.Context, topic string, payload interface{}) error
	Subscribe(ctx context.Context, topic string) (<-chan []byte, error)
}
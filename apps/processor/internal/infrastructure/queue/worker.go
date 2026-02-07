package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/tu-usuario/hub-ai-processor/internal/agent"
	"github.com/tu-usuario/hub-ai-processor/internal/domain"
)

type Worker struct {
	redisClient *redis.Client
	planner     *agent.Planner
	repo        domain.Repository
}

func NewWorker(rc *redis.Client, p *agent.Planner, r domain.Repository) *Worker {
	return &Worker{redisClient: rc, planner: p, repo: r}
}

// Start consume mensajes de Redis y activa el pipeline agéntico
func (w *Worker) Start(ctx context.Context) error {
	pubsub := w.redisClient.Subscribe(ctx, "inbound_emails")
	defer pubsub.Close()

	log.Println("Worker iniciado: Escuchando eventos de Redis...")

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("Error recibiendo mensaje: %v", err)
			continue
		}

		// 1. Unmarshal del email (En inglés)
		var email domain.Email
		if err := json.Unmarshal([]byte(msg.Payload), &email); err != nil {
			log.Printf("Error parseando email: %v", err)
			continue
		}

		// 2. Ejecutar el Planner Agent
		plan, err := w.planner.Plan(ctx, email)
		if err != nil {
			log.Printf("Planner falló para email %s: %v", email.ID, err)
			// Aquí aplicaríamos el Dead Letter Queue (DLQ) para resiliencia
			continue
		}

		// 3. Persistir resultado (HITL o Direct Sync)
		err = w.repo.SaveEmail(ctx, &email) // Guardamos el rastro
		if err != nil {
			log.Printf("Error persistiendo en Supabase: %v", err)
		}
        
        log.Printf("Procesado exitoso: ID %s - Acción: %s - HITL: %v", email.ID, plan.Action, plan.RequiresHITL)
	}
}
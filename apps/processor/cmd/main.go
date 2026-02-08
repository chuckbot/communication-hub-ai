package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/chuckbot/hub-ai-processor/internal/agent"
	"github.com/chuckbot/hub-ai-processor/internal/infrastructure/llm"
	"github.com/chuckbot/hub-ai-processor/internal/infrastructure/queue"
	// "github.com/chuckbot/hub-ai-processor/internal/infrastructure/repository"
)

func main() {
	// 1. Context con Graceful Shutdown (Indispensable para no cortar procesos de IA a medias)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("--- Starting Communication Hub AI Processor (USA Context) ---")

	// 2. Infraestructura: Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// 3. Inyección de Dependencias: LLM (Groq)
	groqKey := os.Getenv("GROQ_API_KEY")
	if groqKey == "" {
		log.Fatal("FATAL: GROQ_API_KEY is not set. AI Agent cannot start.")
	}

	// Instanciamos el cliente de Groq que ahora usará el Prompt optimizado para USA
	llmClient := llm.NewGroqClient(groqKey)

	// El Planner se encarga de la lógica de "English-to-English extraction"
	// Mantenemos el threshold de 0.85 para el flujo Human-in-the-loop
	planner := agent.NewPlanner(llmClient, 0.85)

	// 4. Inicialización del Worker
	// Nota: El repositorio sigue en nil hasta que conectemos Supabase
	worker := queue.NewWorker(rdb, planner, nil)

	// 5. Ejecución del Worker (Concurrency)
	errChan := make(chan error, 1)
	go func() {
		log.Println("Worker active: Listening for inbound school emails from Redis...")
		if err := worker.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// 6. Bloqueo y Shutdown
	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received. Closing gracefully...")
		// Damos 10s para que termine la llamada actual al LLM de Groq
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Println("Processor stopped successfully.")
	case err := <-errChan:
		log.Fatalf("Critical Error in Worker: %v", err)
	}
}
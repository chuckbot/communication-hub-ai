package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tu-usuario/hub-ai-processor/internal/agent"
	"github.com/tu-usuario/hub-ai-processor/internal/infrastructure/queue"
	// "github.com/tu-usuario/hub-ai-processor/internal/infrastructure/llm"
	// "github.com/tu-usuario/hub-ai-processor/internal/infrastructure/repository"
)

func main() {
	// 1. Configuración de Contexto con Cancelación para Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("--- Iniciando Communication Hub AI Processor ---")

	// 2. Inicialización de Infraestructura (Redis para Event-Driven)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// 3. Inyección de Dependencias (Clean Architecture)
	// Aquí inyectarías tus implementaciones reales de LLM y DB
	// llmClient := llm.NewGroqClient(os.Getenv("GROQ_API_KEY"))
	// repo := repository.NewSupabaseRepository(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"))
	
	// Por ahora usamos el Planner con el umbral de HITL definido (0.85)
	planner := agent.NewPlanner(nil, 0.85) 

	// 4. Inicialización del Worker
	worker := queue.NewWorker(rdb, planner, nil) // Inyectar repo cuando esté listo

	// 5. Ejecución del Worker en una Goroutine (Concurrencia de Go)
	errChan := make(chan error, 1)
	go func() {
		log.Println("Worker escuchando eventos de Redis...")
		if err := worker.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// 6. Bloqueo hasta señal de parada o error
	select {
	case <-ctx.Done():
		log.Println("Señal de apagado recibida. Cerrando gracefully...")
		// Tiempo de gracia para terminar procesos pendientes de IA
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Println("Processor apagado de forma segura.")
	case err := <-errChan:
		log.Fatalf("Error crítico en el Worker: %v", err)
	}
}
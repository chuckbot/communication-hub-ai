package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chuckbot/hub-ai-processor/internal/agent"
	"github.com/chuckbot/hub-ai-processor/internal/domain"
	"github.com/chuckbot/hub-ai-processor/internal/infrastructure/llm"
)

func main() {
	// 1. Setup de Dependencias
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("ERROR: Please set the GROQ_API_KEY environment variable")
	}

	client := llm.NewGroqClient(apiKey)
	planner := agent.NewPlanner(client, 0.85)

	// 2. Mock de un correo real de una escuela en USA
	// Escenario: Un correo con fecha relativa ("next Wednesday")
	testEmail := domain.Email{
		ID: "test-usa-001",
		RawBody: `
			Subject: 4th Grade Field Trip - Museum of Science
			Dear Parents, 
			Our field trip is officially scheduled for next Wednesday. 
			Please ensure students arrive at 8:15 AM for boarding. 
			Lunch is provided, but students should bring a water bottle.
		`,
	}

	fmt.Println("--- Sending US School Email to Groq ---")
	
	// 3. Ejecución del Planner
	plan, err := planner.Plan(context.Background(), testEmail)
	if err != nil {
		log.Fatalf("Planner failed: %v", err)
	}

	// 4. Verificación de Resultados
	fmt.Printf("\n[ANALYSIS RESULTS]\n")
	fmt.Printf("Action Categorized: %s\n", plan.Action)
	fmt.Printf("Confidence Level: %.2f\n", plan.Confidence)
	fmt.Printf("Needs Human Review (HITL): %v\n", plan.RequiresHITL)
	fmt.Printf("AI Reasoning: %s\n", plan.Reasoning)
	
	fmt.Printf("\n[EXTRACTED DATA]\n")
	fmt.Printf("Title: %v\n", plan.Data["title"])
	fmt.Printf("Calculated Date (ISO): %v\n", plan.Data["date"])
	fmt.Printf("Summary for Parents: %v\n", plan.Data["description"])
}
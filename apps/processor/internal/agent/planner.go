package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chuckbot/hub-ai-processor/internal/domain"
)

// Constante con las instrucciones maestras para el LLM
const extractionSystemPrompt = `
ROLE: Expert School Administrative Assistant (USA Context).
TASK: Analyze the school email in English and extract calendar events or deadlines.

INSTRUCTIONS:
1. REASONING: Think step-by-step in English. Use the "Current Date" to resolve relative dates like "next Thursday" or "this coming Friday".
2. LANGUAGE: All output must be in English.
3. SYNTHESIS: The "description" should be a concise, professional summary for parents (max 2 sentences).
4. CONFIDENCE: Set confidence < 0.85 if details are missing or the date calculation is uncertain.

JSON SCHEMA:
{
  "reasoning": "Internal thought process in English",
  "action": "CALENDAR_SYNC" | "PAYMENT_REQUIRED" | "GENERAL_INFO",
  "confidence": float,
  "requires_hitl": boolean,
  "extracted_data": {
    "title": "Clear event title",
    "date": "ISO-8601 string",
    "description": "Executive summary in English for the parent"
  }
}
`

type AgentPlan struct {
	Reasoning    string                 `json:"reasoning"`
	Action       string                 `json:"action"`
	Confidence   float64                `json:"confidence"`
	Data         map[string]interface{} `json:"extracted_data"`
	RequiresHITL bool                   `json:"requires_hitl"`
}

type Planner struct {
	llmClient     domain.LLMClient
	hitlThreshold float64
}

func NewPlanner(client domain.LLMClient, threshold float64) *Planner {
	return &Planner{
		llmClient:     client,
		hitlThreshold: threshold,
	}
}

func (p *Planner) Plan(ctx context.Context, email domain.Email) (*AgentPlan, error) {
	// 1. Metadata Enrichment: Le damos al agente el "Hoy" real para sus cálculos
	now := time.Now().Format("Monday, January 02, 2026")

	// 2. Construcción del Prompt con el contexto dinámico
	systemPrompt := fmt.Sprintf("%s\nCONTEXT: Today is %s. Location: Coro, Venezuela.", extractionSystemPrompt, now)
	userPrompt := fmt.Sprintf("EMAIL CONTENT:\n%s", email.RawBody)

	// 3. Ejecución del LLM (Groq)
	response, err := p.llmClient.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	// 4. Parsing con Validación
	var plan AgentPlan
	if err := json.Unmarshal([]byte(response), &plan); err != nil {
		return nil, fmt.Errorf("invalid JSON from LLM: %w", err)
	}

	// 5. Lógica de Negocio HITL (Human-in-the-loop)
	// Si el LLM tiene dudas (confidence bajo) o nosotros forzamos por threshold
	if plan.Confidence < p.hitlThreshold {
		plan.RequiresHITL = true
	}

	// Log de observabilidad senior
	if plan.RequiresHITL {
		fmt.Printf("[HITL] Email %s marcado para revisión manual (Conf: %.2f)\n", email.ID, plan.Confidence)
	}

	return &plan, nil
}
package agent

import (
	"context"
	"encoding/json"
	"fmt"
  "time"
  
	"github.com/chuckbot/hub-ai-processor/internal/domain"
)

// AgentPlan representa la decisión tomada por el Planner
type AgentPlan struct {
	Reasoning   string                 `json:"reasoning"`    // Self-reflection / critic loop
	Action      string                 `json:"action"`       // e.g., "CALENDAR_SYNC", "PAYMENT_REMINDER", "IGNORE"
	Confidence  float64                `json:"confidence"`   // Para el Human-in-the-loop workflow
	Data        map[string]interface{} `json:"extracted_data"`
	RequiresHITL bool                  `json:"requires_hitl"`
}

// Planner es el agente encargado de orquestar la lógica agéntica
type Planner struct {
	llmClient     domain.LLMClient // Interfaz definida en domain
	hitlThreshold float64
}

func NewPlanner(client domain.LLMClient, threshold float64) *Planner {
	return &Planner{
		llmClient:     client,
		hitlThreshold: threshold,
	}
}

// Plan analiza un correo y decide el siguiente paso
func (p *Planner) Plan(ctx context.Context, email domain.Email) (*AgentPlan, error) {
  // 1. Metadata Enrichment: Crucial para la precisión de fechas
  now := time.Now().Format("Monday, Jan 02, 2026")
  
  // 2. Prompt Engineering Avanzado
  // Separamos el rol de las restricciones (Role Separation)
  systemPrompt := fmt.Sprintf(`
    ROLE: Senior School Communication Agent.
    CONTEXT: Today is %s. Current location: Venezuela.
    INPUT: You will receive a school email in English.
    
    TASK:
    - Analyze the English text for events, deadlines, or payments.
    - If the email is in English, perform your internal reasoning in English to avoid semantic loss.
    - Extract dates and normalize them to ISO-8601.
    
    CONSTRAINTS:
    - Output MUST be strictly JSON.
    - reasoning: Explain your thought process in English (Self-reflection).
    - extracted_data: Keys should be stable (title, date, description).
    - If any date is ambiguous (e.g., 'next Tuesday' without a clear reference), set confidence < 0.85.
  `, now)

  // 3. Ejecución del LLM
  response, err := p.llmClient.Complete(ctx, systemPrompt, email.RawBody)
  if err != nil {
    return nil, fmt.Errorf("LLM completion failed: %w", err)
  }

  // 4. Parsing con Validación (Safety & Guardrails)
  var plan AgentPlan
  if err := json.Unmarshal([]byte(response), &plan); err != nil {
    // Aquí podrías implementar un "Self-correction loop" re-enviando el error al LLM
    return nil, fmt.Errorf("invalid JSON from LLM: %w", err)
  }

  // 5. Lógica de Negocio: Autonomía vs Control (HITL)
  // Si el puntaje es bajo o el agente mismo marcó dudas, forzamos intervención humana
  if plan.Confidence < p.hitlThreshold || plan.RequiresHITL {
    plan.RequiresHITL = true
    // Log de observabilidad para debugging en sistemas probabilísticos
    fmt.Printf("[HITL Triggered] Low confidence (%.2f) for email: %s\n", plan.Confidence, email.ID)
  }

  return &plan, nil
}
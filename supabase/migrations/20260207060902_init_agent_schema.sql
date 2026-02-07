-- Tabla principal de comunicaciones procesadas
CREATE TABLE processed_emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    raw_content TEXT NOT NULL,
    sender TEXT,
    subject TEXT,
    extracted_data JSONB, -- Aquí el Agente guarda fechas, tareas, etc.
    confidence_score FLOAT, -- Para decidir si va a HITL
    status TEXT CHECK (status IN ('pending', 'processing', 'completed', 'requires_review', 'failed')),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Tabla para el Master Calendar (Sincronizado por el Agente)
CREATE TABLE school_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_id UUID REFERENCES processed_emails(id),
    title TEXT NOT NULL,
    description TEXT,
    event_date TIMESTAMPTZ NOT NULL,
    is_validated_by_human BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now()
);
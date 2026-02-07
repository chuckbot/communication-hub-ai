# School Communication Hub AI (MVP)

### Tech Stack
- **Backend**: Go (Agentic Processor)
- **Frontend**: Next.js 15 (React Compiler enabled)
- **Infraestructura**: Supabase (Postgres + Auth + RLS)
- **Event-Driven**: Redis
- **AI**: Agentic Architecture (Planner/Executor) with Human-in-the-loop.

### Local Setup (WSL2 / Docker)
1. `supabase start`
2. `docker-compose up -d`
3. `pnpm install && pnpm dev`

### Arquitecture
A **Clean Architecture** project and an event-driven flow to porcess emails.
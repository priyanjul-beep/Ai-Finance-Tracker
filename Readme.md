# 🤖 AI Finance Tracker

A **production-grade, AI-powered personal finance assistant** built with Go (Gin) + Next.js 14.

Track expenses, manage budgets and goals, get real-time AI insights, parse receipts with OCR, and predict future spending — all in one place.

---

## ✨ Features

| Category | Capabilities |
|---|---|
| **AI** | Natural-language expense parsing · Smart categorization · Spending insights · Next-month predictions · Financial health score |
| **Expenses** | Full CRUD · Duplicate detection · Receipt OCR · Tags |
| **Budgets** | Category-level budgets · Real-time status · Email alerts |
| **Income** | Track multiple income streams · Tax-aware records |
| **Subscriptions** | Billing-cycle management · Upcoming renewal alerts |
| **Goals** | Savings goals · Progress tracking · Monthly savings calculator |
| **Analytics** | Dashboard · Monthly/Yearly reports · Predictions |
| **Auth** | JWT (access + refresh) · Google OAuth 2.0 · Email verification |
| **Infrastructure** | Redis cache · Asynq background jobs · Prometheus metrics · Docker |

---

## 🏗️ Architecture

```
AI Finance Tracker/
├── backend/                     # Go API server (Clean Architecture)
│   ├── cmd/api/main.go          # Entry point, DI wiring
│   ├── config/
│   ├── internal/
│   │   ├── domain/              # 13 GORM models
│   │   ├── interfaces/          # All contracts
│   │   ├── dto/                 # Request/response DTOs
│   │   ├── repository/          # GORM implementations
│   │   ├── usecase/             # Business logic
│   │   ├── handler/             # Gin HTTP handlers
│   │   └── middleware/
│   └── pkg/
│       ├── ai/                  # Gemini provider (pluggable)
│       └── auth/cache/database/email/monitoring/queue/storage/
├── frontend/                    # Next.js 14 App Router
│   └── src/
│       ├── app/                 # Pages (auth + dashboard)
│       ├── components/          # UI components
│       ├── hooks/               # TanStack Query hooks
│       ├── services/            # Axios API layer
│       ├── store/               # Zustand global state
│       └── lib/                 # Utils, query client
├── docker/                      # Dockerfiles + Prometheus
└── docker-compose.yml
```

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+, Node.js 20+, Docker & Docker Compose
- PostgreSQL (or [Neon](https://neon.tech) free tier)

### Configure

```bash
cp backend/.env.example backend/.env
# Set: DATABASE_URL, JWT_SECRET, GEMINI_API_KEY

cp frontend/.env.example frontend/.env.local
```

### Run with Docker

```bash
docker-compose up -d
```

### Run manually

```bash
# Backend
cd backend && go mod download && go run ./cmd/api

# Frontend
cd frontend && npm install && npm run dev
```

Open **http://localhost:3000**

---

## 🔑 Required Environment Variables

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection string |
| `JWT_SECRET` | Min 32 chars |
| `JWT_REFRESH_SECRET` | Min 32 chars |
| `GEMINI_API_KEY` | Google AI Studio API key |

---

## 📡 API Endpoints (`/api/v1`)

- **Auth:** `POST /auth/signup` · `/auth/login` · `/auth/refresh` · `/auth/logout`
- **Expenses:** `CRUD /expenses` · `/expenses/ai-parse` · `/expenses/search`
- **Income / Budgets / Subscriptions / Goals / Tags:** Full CRUD
- **Analytics:** `/analytics/dashboard` · `/analytics/monthly` · `/analytics/predictions` · `/analytics/health-score`
- **Notifications:** `GET /notifications` · `PATCH /notifications/{id}/read`

---

## 🤝 Tech Stack

**Backend:** Go 1.21, Gin, GORM, PostgreSQL, Redis, Asynq, Gemini AI, JWT, Prometheus, Zap

**Frontend:** Next.js 14, TypeScript, Tailwind CSS, shadcn/ui, Framer Motion, Recharts, TanStack Query v5, Zustand, React Hook Form + Zod

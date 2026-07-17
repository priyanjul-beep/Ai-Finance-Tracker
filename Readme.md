#  AI Finance Tracker

A **production-grade, AI-powered personal finance assistant** built with Go (Gin) + Next.js 14.

Track expenses, manage budgets and goals, get real-time AI insights, parse receipts with OCR, and predict future spending — all in one place.

---

## 🚀 Live Demo

| | URL |
|---|---|
| 🌐 **Frontend** | [https://ai-finance-tracker-2-q25a.onrender.com](https://ai-finance-tracker-2-q25a.onrender.com) |
| ⚙️ **Backend API** | [https://ai-finance-tracker-1-5js1.onrender.com](https://ai-finance-tracker-1-5js1.onrender.com) |
| 💓 **Health Check** | [https://ai-finance-tracker-1-5js1.onrender.com/api/v1/health](https://ai-finance-tracker-1-5js1.onrender.com/api/v1/health) |

> **Note:** The app is hosted on Render's free tier — the server may take **30–60 seconds to spin up** on the first request after a period of inactivity.

---

## ✨ Features

| Category | Capabilities |
|---|---|
| **AI** | Natural-language expense parsing · Smart categorization · Spending insights · Next-month predictions · Financial health score |
| **Expenses** | Full CRUD · Duplicate detection · Receipt OCR · Voice input · Tags |
| **Budgets** | Category-level budgets · Real-time status · 90% / custom threshold alerts |
| **Income** | Track multiple income streams · Tax-aware records |
| **Subscriptions** | Billing-cycle management · Upcoming renewal alerts |
| **Goals** | Savings goals · Progress tracking · Monthly savings calculator |
| **Analytics** | Dashboard · Monthly/Yearly reports · Predictions · Live financial health score |
| **Notifications** | In-app notification center · Priority badges · Infinite scroll · Unread badge |
| **Email** | Responsive HTML emails · Welcome · Budget warning · Budget exceeded |
| **Auth** | JWT (access + refresh) · Google OAuth 2.0 · Email verification |
| **Infrastructure** | Redis cache · Asynq background workers · Prometheus metrics · Docker |

---

##  Architecture

```
AI Finance Tracker/
├── backend/                         # Go API server (Clean Architecture)
│   ├── cmd/api/main.go              # Entry point, DI wiring, Asynq worker server
│   ├── config/
│   ├── internal/
│   │   ├── domain/                  # GORM models (Notification with Priority)
│   │   ├── interfaces/              # All repository + service contracts
│   │   ├── dto/                     # Request/response DTOs
│   │   ├── repository/              # GORM implementations
│   │   ├── usecase/                 # Business logic
│   │   │   ├── analytics_usecase.go # Live health score computation
│   │   │   ├── notification_usecase.go # Redis-cached unread count
│   │   │   └── auth_usecase.go      # Enqueues welcome notification on signup
│   │   ├── handler/                 # Gin HTTP handlers
│   │   └── middleware/
│   ├── migrations/
│   │   ├── 001_initial_schema.sql
│   │   └── 002_notification_enhancements.sql
│   └── pkg/
│       ├── ai/                      # Gemini provider (pluggable)
│       ├── email/                   # HTML email templates
│       ├── queue/                   # Asynq client + worker implementations
│       └── auth/cache/database/monitoring/storage/
├── frontend/                        # Next.js 14 App Router
│   ├── .vscode/                     # Tailwind CSS IntelliSense settings
│   └── src/
│       ├── app/(dashboard)/
│       │   └── notifications/       # Notification center (infinite scroll, filters)
│       ├── components/
│       │   └── layout/Header.tsx    # Live unread badge (30s polling)
│       ├── hooks/
│       ├── services/
│       ├── store/
│       └── types/index.ts           # NotificationType, NotificationPriority
├── docker/
└── docker-compose.yml
```

---

##  Quick Start

### Prerequisites

- Go 1.21+, Node.js 20+, Docker & Docker Compose
- PostgreSQL (or [Neon](https://neon.tech) free tier)
- Redis (or [Upstash](https://upstash.com) free tier)

### Configure

```bash
cp backend/.env.example backend/.env
# Required: DATABASE_URL, REDIS_URL, JWT_SECRET, JWT_REFRESH_SECRET, GEMINI_API_KEY

cp frontend/.env.example frontend/.env.local
# Required: NEXT_PUBLIC_API_URL
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

##  Environment Variables

| Variable | Description | Required |
|---|---|---|
| `DATABASE_URL` | PostgreSQL connection string | ✅ |
| `REDIS_URL` | Redis connection string | ✅ |
| `JWT_SECRET` | Min 32 chars | ✅ |
| `JWT_REFRESH_SECRET` | Min 32 chars | ✅ |
| `GEMINI_API_KEY` | Google AI Studio API key | ✅ |
| `SMTP_HOST` | SMTP server hostname | Optional |
| `SMTP_PORT` | SMTP port (default: 587) | Optional |
| `SMTP_USERNAME` | SMTP username | Optional |
| `SMTP_PASSWORD` | SMTP password | Optional |
| `SMTP_FROM` | Sender email address | Optional |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | Optional |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | Optional |

> **Note:** SMTP is optional — if not configured, emails are skipped silently and notifications are still saved to the database.

---

##  API Endpoints (`/api/v1`)

### Auth
| Method | Path | Description |
|---|---|---|
| `POST` | `/auth/signup` | Register + enqueues welcome notification |
| `POST` | `/auth/login` | Login |
| `POST` | `/auth/refresh` | Refresh token pair |
| `POST` | `/auth/forgot-password` | Send reset email |
| `GET` | `/auth/verify-email/:token` | Verify email |

### Expenses
| Method | Path | Description |
|---|---|---|
| `POST` | `/expenses` | Create + triggers budget check worker |
| `GET` | `/expenses` | List (paginated) |
| `GET` | `/expenses/search?q=` | AI natural-language search |
| `POST` | `/expenses/parse` | AI text → expense |
| `POST` | `/expenses/voice-parse` | Voice audio → expense |
| `POST` | `/expenses/scan` | Receipt image → expense |

### Notifications
| Method | Path | Description |
|---|---|---|
| `GET` | `/notifications` | List (paginated, filter by `?type=`) |
| `GET` | `/notifications/unread-count` | Cached unread badge count |
| `PATCH` | `/notifications/:id/read` | Mark one as read |
| `PATCH` | `/notifications/read-all` | Mark all as read |
| `DELETE` | `/notifications/:id` | Delete |

### Analytics
| Method | Path | Description |
|---|---|---|
| `GET` | `/analytics/dashboard` | Full dashboard payload |
| `GET` | `/analytics/health-score` | Live financial health score |
| `GET` | `/analytics/monthly/:month/:year` | Monthly report |
| `GET` | `/analytics/yearly/:year` | Yearly overview |
| `GET` | `/analytics/predictions` | AI spending forecast |
| `GET` | `/analytics/insights` | AI tips |

- **Income / Budgets / Subscriptions / Goals / Tags:** Full CRUD

---

##  Database Migrations

Run in order before starting the server:

```bash
psql $DATABASE_URL -f backend/migrations/001_initial_schema.sql
psql $DATABASE_URL -f backend/migrations/002_notification_enhancements.sql
```

**Migration 002** adds `priority` + `metadata` columns to `notifications` and creates composite indexes for fast unread-count queries.

---

##  Tech Stack

**Backend:** Go 1.21 · Gin · GORM · PostgreSQL · Redis · Asynq · Gemini AI · JWT · Prometheus · Zap Logger

**Frontend:** Next.js 14 · TypeScript · Tailwind CSS · TanStack Query v5 · Zustand · Recharts · Sonner (toasts)

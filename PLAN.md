# Plan

## Step 1 – Scaffold
- Create repository structure with backend (Go + Gin skeleton), frontend (Vite + React + TypeScript + Tailwind), and infra (Docker Compose).
- Add README and environment examples; ensure Docker Compose boots PostgreSQL and backend (health check).

## Step 2 – Database & migrations
- Configure PostgreSQL connection via environment variables.
- Add migration setup for core tables (sites, summaries, generation_requests, reports) and migration runner.

## Step 3 – Core API skeleton
- Implement health check and summary endpoints with stubbed data and throttling logic placeholders.
- Wire CORS and basic validation; ensure frontend can call APIs.

## Step 4 – OpenAI integration + worker
- Implement ToS fetching, text extraction, OpenAI prompt/validation, and background worker for generation requests with 7-day throttling.

## Step 5 – Frontend UI
- Build search experience, summary display states, report/human review actions, and stale/pending indicators.

## Step 6 – Admin UI
- Add basic admin view/edit for summaries with basic auth; integrate admin endpoints.

## Step 7 – Polish
- Improve UX states, error handling, production build flow, and documentation.

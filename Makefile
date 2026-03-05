SHELL := /bin/bash

.PHONY: setup setup-backend setup-frontend dev-backend dev-frontend test

setup: setup-backend setup-frontend

setup-backend:
	@if [ ! -f backend/.env ]; then \
		cp backend/.env.example backend/.env; \
		echo "backend/.env cree (pense a renseigner OPENAI_API_KEY)"; \
	else \
		echo "backend/.env deja present"; \
	fi

setup-frontend:
	cd frontend && npm install

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev -- --host --port 5173

test:
	cd backend && go test ./...
	cd frontend && npm run build

# TOSAI

![License](https://img.shields.io/github/license/gabrielb0x/TOSAI) ![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go) ![Stars](https://img.shields.io/github/stars/gabrielb0x/TOSAI?style=social)

> Plateforme francophone pour analyser et résumer les CGU/ToS : collecte par URL, résumé IA, notation A→E, quotas par domaine et signalements.

## Sommaire
- [Présentation](#présentation)
- [Fonctionnalités clés](#fonctionnalités-clés)
- [Stack technique](#stack-technique)
- [Arborescence](#arborescence)
- [Démarrage rapide (local)](#démarrage-rapide-local)
- [Configuration PostgreSQL](#configuration-postgresql)
- [Lancer le backend Go](#lancer-le-backend-go)
- [Lancer le frontend Vite](#lancer-le-frontend-vite-htmlcssjs)
- [Déploiement Raspberry Pi 5 + NGINX](#déploiement-raspberry-pi-5--nginx)
- [Variables d’environnement](#variables-denvironnement)
- [Tests](#tests)

## Présentation
TOSAI récupère les CGU/ToS d’un site, extrait le texte, interroge OpenAI puis renvoie un JSON d’analyse (note A→E, résumé, points clés, risques, recommandation). Le frontend Vite consomme directement cette API.

## Fonctionnalités clés
- Endpoint `/api/v1/summary` (`GET` ou `POST`) branché sur OpenAI.
- Extraction de texte HTML/plain text avant analyse IA.
- Résumé IA + notation A→E + points clés + risques + recommandation.
- Quotas journaliers par domaine (table limites + usage quotidien).
- Bouton de signalement et audit admin (optionnel).
- Panel admin sans comptes publics (protégé par un token côté backend).

## Stack technique
- **Backend** : Go (Gin, pgx), chargement .env prioritaire `backend/.env`, port HTTP par défaut **9000**.
- **Base** : PostgreSQL 15+ (UUID + JSONB), migrations idempotentes intégrées (lecture de `backend/config/database_init.sql`).
- **Frontend** : Vite + HTML/CSS/JS (vanilla).

## Arborescence
- `backend/` : API Go, migrations intégrées, fichiers `.env`.
- `frontend/` : client Vite en HTML/CSS/JS.
- `PLAN.md` : feuille de route.

## Démarrage rapide (local)
```bash
# 1) Cloner & se placer sur la racine
cd TOSAI

# Option A (rapide): préparer l'environnement automatiquement
make setup

# Option B (manuel): préparer l'environnement
cp backend/.env.example backend/.env
# renseigner OPENAI_API_KEY dans backend/.env

# 2) Démarrer le backend (terminal 1)
cd backend
go run ./cmd/server

# 3) Démarrer le frontend (terminal 2)
cd ../frontend
npm install
npm run dev -- --host --port 5173
```
- L'API écoute sur http://localhost:9000.
- Le frontend est sur http://localhost:5173 et proxy automatiquement `/api` vers `:9000`.
- La page de test API : http://localhost:9000/web/test.html

## Configuration PostgreSQL
1. Créer la base et l’utilisateur applicatif (commande à exécuter avec un superuser) :
   ```bash
   psql -U postgres <<'SQL'
   CREATE DATABASE tosai;
   CREATE USER tosai_app WITH PASSWORD 'motdepasse';
   GRANT ALL PRIVILEGES ON DATABASE tosai TO tosai_app;
   SQL
   ```
2. Appliquer le schéma si nécessaire (idempotent) :
   ```bash
   psql -d tosai -f backend/config/database_init.sql
   ```
3. Au démarrage, le backend applique automatiquement le même DDL via pgx. Si l’utilisateur n’a pas les droits `CREATE`/`ALTER`, le log indiquera l’erreur et la marche à suivre.

> Aucun rôle ni base n’est créé automatiquement par le code en production : fournissez un utilisateur ayant les droits sur la base cible.

## Lancer le backend Go
```bash
cd backend
cp .env.example .env  # ou utilisez backend/config/.env si besoin
APP_PORT=9000 go run ./cmd/server
```
- `OPENAI_API_KEY` est requis pour utiliser `/api/v1/summary`.
- `DATABASE_URL` est optionnel (mode API-only possible pour dev rapide).
- Si `DATABASE_URL` est fourni, le backend tente connexion + migration (`database_init.sql`).
- Variables supportées : voir [Variables d’environnement](#variables-denvironnement).

## Lancer le frontend Vite (HTML/CSS/JS)
```bash
cd frontend
npm install
npm run dev -- --host --port 5173
# ou build production
npm run build
```
- En local, aucune variable front n'est obligatoire (proxy Vite actif).
- Optionnel: `VITE_API_BASE_URL` pour forcer une base API externe.

## Déploiement Raspberry Pi 5 + NGINX
- OS recommandé : Raspberry Pi OS/Debian 12 (arm64). Go et Node fonctionnent nativement.
- Placer le code dans `/opt/tosai` et les artefacts frontend buildés dans `/var/www/tosai`.
- Exemple de service backend :
  ```bash
  cd /opt/tosai/backend
  APP_ENV=prod APP_PORT=9000 DATABASE_URL=postgres://tosai_app:motdepasse@127.0.0.1:5432/tosai?sslmode=disable \
    /usr/local/bin/go run ./cmd/server
  ```
- Exemple complet NGINX : `nginx.example.conf` à la racine du repo.
- Pensez à ouvrir le port 9000 localement uniquement (NGINX fait l’externalisation).

## Variables d’environnement
Fichier prioritaire : `backend/.env` (puis `backend/config/.env`).
- `APP_ENV` : `local` | `prod`
- `APP_PORT` : port HTTP (défaut `9000`)
- `DATABASE_URL` : optionnel, ex. `postgres://tosai_app:motdepasse@127.0.0.1:5432/tosai?sslmode=disable`
- `CORS_ORIGINS` : liste d’origines séparées par des virgules ou `*`
- `ADMIN_API_TOKEN` : token serveur pour sécuriser les endpoints admin
- `OPENAI_API_KEY` : clé requise pour l'analyse OpenAI
- `OPENAI_MODEL` : modèle OpenAI (défaut `gpt-4.1-mini`)
- `OPENAI_BASE_URL` : base URL API OpenAI (défaut `https://api.openai.com/v1`)
- `HTTP_TIMEOUT_SECONDS` : timeout HTTP global (défaut `25`)
- `ANALYSIS_INPUT_MAX_CHARS` : taille max du texte envoyé à OpenAI (défaut `12000`)
- `VITE_BACKEND_PROXY_TARGET` (frontend/dev) : cible proxy Vite (défaut `http://localhost:9000`)

## Tests
```bash
cd backend
go test ./...

cd ../frontend
npm run build
```

Bonne contribution !

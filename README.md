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
- [Lancer le frontend React](#lancer-le-frontend-react)
- [Déploiement Raspberry Pi 5 + NGINX](#déploiement-raspberry-pi-5--nginx)
- [Variables d’environnement](#variables-denvironnement)
- [Tests](#tests)

## Présentation
TOSAI récupère les CGU/ToS d’un site, les résume automatiquement (OpenAI), attribue une note A→E, stocke le résultat (texte + JSON) et applique des quotas par domaine. Un bouton de signalement et un panneau admin (protégé par token serveur) complètent l’expérience.

## Fonctionnalités clés
- Récupération des CGU/ToS par URL et stockage versionné (sha256).
- Résumés IA + notation A→E, avec métadonnées et coûts token.
- Quotas journaliers par domaine (table limites + usage quotidien).
- Bouton de signalement et audit admin (optionnel).
- Panel admin sans comptes publics (protégé par un token côté backend).

## Stack technique
- **Backend** : Go (Gin, pgx), chargement .env prioritaire `backend/.env`, port HTTP par défaut **9000**.
- **Base** : PostgreSQL 15+ (UUID + JSONB), migrations idempotentes intégrées (lecture de `backend/config/database_init.sql`).
- **Frontend** : React + Vite + TypeScript + Tailwind.
- **Infra** : Docker Compose de dev, reverse proxy recommandé via NGINX.

## Arborescence
- `backend/` : API Go, migrations intégrées, fichiers `.env`.
- `frontend/` : client Vite/React/TS.
- `infra/` : docker-compose pour dev local.
- `PLAN.md` : feuille de route.

## Démarrage rapide (local)
```bash
# 1) Cloner & se placer sur la racine
cd TOSAI

# 2) Préparer l'environnement backend
touch backend/.env
cp backend/.env.example backend/.env

# 3) Installer les dépendances frontend
cd frontend
npm install
npm run dev -- --host --port 5173
```
- L’API écoute sur http://localhost:9000 (proxy facile avec NGINX).
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
- Le serveur logge la connexion DB, applique les migrations (`database_init.sql`) et démarre sur le port configuré.
- Variables supportées : voir [Variables d’environnement](#variables-denvironnement).

## Lancer le frontend React
```bash
cd frontend
npm install
npm run dev -- --host --port 5173
# ou build production
npm run build
```
Configurez `VITE_API_BASE_URL` selon votre proxy (par défaut http://localhost:9000).

## Déploiement Raspberry Pi 5 + NGINX
- OS recommandé : Raspberry Pi OS/Debian 12 (arm64). Go et Node fonctionnent nativement.
- Placer le code dans `/opt/tosai` et les artefacts frontend buildés dans `/var/www/tosai`.
- Exemple de service backend :
  ```bash
  cd /opt/tosai/backend
  APP_ENV=prod APP_PORT=9000 DATABASE_URL=postgres://tosai_app:motdepasse@127.0.0.1:5432/tosai?sslmode=disable \
    /usr/local/bin/go run ./cmd/server
  ```
- Exemple de bloc NGINX (reverse proxy) :
  ```nginx
  server {
    listen 80;
    server_name tosai.fr;

    location / {
      root /var/www/tosai;
      try_files $uri $uri/ /index.html;
    }

    location /api/ {
      proxy_pass http://127.0.0.1:9000;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
    }

    location /healthz {
      proxy_pass http://127.0.0.1:9000/healthz;
    }
  }
  ```
- Pensez à ouvrir le port 9000 localement uniquement (NGINX fait l’externalisation).

## Variables d’environnement
Fichier prioritaire : `backend/.env` (puis `backend/config/.env`).
- `APP_ENV` : `local` | `prod`
- `APP_PORT` : port HTTP (défaut `9000`)
- `DATABASE_URL` : ex. `postgres://tosai_app:motdepasse@127.0.0.1:5432/tosai?sslmode=disable`
- `CORS_ORIGINS` : liste d’origines séparées par des virgules ou `*`
- `ADMIN_API_TOKEN` : token serveur pour sécuriser les endpoints admin
- `OPENAI_API_KEY` : clé pour les futurs appels de génération

## Tests
```bash
cd backend
go test ./...

cd ../frontend
npm run lint
npm run build
```

Bonne contribution !

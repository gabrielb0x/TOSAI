# Backend TOSAI

Ce dossier contient le backend Go de TOSAI. Il expose maintenant une API simple:

- route principale: `POST /summary`
- lecture rapide: `GET /summary?url=https://example.com/terms`
- alias de compatibilite: `/v1/summary`

Il n'y a plus de route `/healthz` ni de page web embarquee dans le backend. Le processus a lancer est simplement le serveur API.

## Prerequis

- Go 1.22+
- PostgreSQL 15+ si vous voulez de la persistance
- une cle `OPENAI_API_KEY`

## Mise en place locale

Depuis ce dossier:

```bash
cp .env.example .env
nano .env
go run ./cmd/server
```

Variables minimales a renseigner:

```env
APP_ENV=local
APP_PORT=9000
OPENAI_API_KEY=sk-...
```

Si `DATABASE_URL` est vide, le backend demarre en mode API-only, sans PostgreSQL.

## Variables d'environnement

Le backend charge en priorite `backend/.env`, puis `backend/config/.env`.

- `APP_ENV`: `local` ou `prod`
- `APP_PORT`: port HTTP local du backend
- `DATABASE_URL`: URL PostgreSQL complete
- `CORS_ORIGINS`: liste CSV des origines autorisees
- `ADMIN_API_TOKEN`: reserve pour des endpoints admin
- `OPENAI_API_KEY`: obligatoire pour `/summary`
- `OPENAI_MODEL`: defaut `gpt-4.1-mini`
- `OPENAI_BASE_URL`: defaut `https://api.openai.com/v1`
- `HTTP_TIMEOUT_SECONDS`: timeout reseau global
- `ANALYSIS_INPUT_MAX_CHARS`: taille max du texte envoye a OpenAI

Exemple complet:

```env
APP_ENV=prod
APP_PORT=9000
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
DATABASE_URL=postgres://tosai_app:change-me@127.0.0.1:5432/tosai?sslmode=disable
CORS_ORIGINS=https://tosai.fr,https://www.tosai.fr
HTTP_TIMEOUT_SECONDS=25
ANALYSIS_INPUT_MAX_CHARS=12000
```

## Configuration PostgreSQL

Le backend applique automatiquement le schema embarque depuis `config/database_init.sql` au demarrage, si `DATABASE_URL` est defini et valide.

### 1. Creer un role et une base

Exemple sur un serveur Linux avec PostgreSQL local:

```bash
sudo -u postgres psql <<'SQL'
CREATE ROLE tosai_app LOGIN PASSWORD 'change-me';
CREATE DATABASE tosai OWNER tosai_app;
SQL
```

### 2. Donner l'extension si necessaire

Le schema utilise `pgcrypto`. Si votre utilisateur n'a pas le droit de creer les extensions, faites-le une fois avec un compte admin:

```bash
sudo -u postgres psql -d tosai -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
```

### 3. Configurer `DATABASE_URL`

```env
DATABASE_URL=postgres://tosai_app:change-me@127.0.0.1:5432/tosai?sslmode=disable
```

### 4. Demarrer le backend

Au premier demarrage valide:

- connexion PostgreSQL
- `Ping`
- application automatique du schema SQL

Si la connexion echoue, le backend continue de demarrer sans DB. Cela permet un mode API-only, mais aucune persistance PostgreSQL n'est alors disponible.

## Build production

```bash
mkdir -p /opt/tosai/bin
go build -o /opt/tosai/bin/tosai-backend ./cmd/server
```

Le binaire final sera:

```bash
/opt/tosai/bin/tosai-backend
```

## Service qui tourne en boucle sur un serveur

Sur un serveur Linux classique, le plus simple est `systemd`. Le redemarrage automatique est gere par `Restart=always`.

Creer `/etc/systemd/system/tosai-backend.service`:

```ini
[Unit]
Description=TOSAI Backend API
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/tosai/backend
EnvironmentFile=/opt/tosai/backend/.env
ExecStart=/opt/tosai/bin/tosai-backend
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Activer le service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now tosai-backend
sudo systemctl status tosai-backend
sudo journalctl -u tosai-backend -f
```

Resultat:

- demarrage automatique au boot
- relance automatique si le process tombe
- logs centralises dans `journalctl`

## Exemples d'appel API

### POST

```bash
curl -X POST http://127.0.0.1:9000/summary \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/terms"}'
```

### GET

```bash
curl 'http://127.0.0.1:9000/summary?url=https://example.com/terms'
```

## Reverse proxy conseille

Si vous exposez le service derriere NGINX:

- backend Go sur `127.0.0.1:9000`
- TLS et domaine public geres par NGINX
- CORS limite aux domaines frontend reels dans `CORS_ORIGINS`

Le fichier `../nginx.example.conf` peut servir de base.

## Depannage rapide

- `missing_openai_api_key`: verifier `OPENAI_API_KEY`
- `fetch_failed`: la page cible n'est pas accessible depuis le serveur
- `openai_http_error`: verifier cle, modele, quota ou proxy OpenAI
- erreur SQL au boot: verifier `DATABASE_URL`, droits de l'utilisateur et extension `pgcrypto`

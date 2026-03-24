# Backend TOSAI

Ce dossier contient le backend Go de TOSAI. Il expose maintenant une API simple:

- racine API versionnee: `GET /v1/`
- route principale: `POST /v1/summary`
- lecture rapide: `GET /v1/summary?url=https://example.com/terms`
- alias de compatibilite conserve: `/summary`

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

Si vous voulez preparer PostgreSQL en une commande:

```bash
cd /opt/tosai
make db-init
```

ou directement:

```bash
./scripts/init_db.sh
```

Le script:

- cree le role PostgreSQL si besoin
- cree la base si besoin
- applique `config/database_init.sql`
- renseigne `DATABASE_URL` dans `backend/.env`

## Variables d'environnement

Le backend charge en priorite `backend/.env`, puis `backend/config/.env`.

- `APP_ENV`: `local` ou `prod`
- `APP_PORT`: port HTTP local du backend
- `DATABASE_URL`: URL PostgreSQL complete
- `CORS_ORIGINS`: liste CSV des origines autorisees
- `ADMIN_API_TOKEN`: reserve pour des endpoints admin
- `OPENAI_API_KEY`: obligatoire pour `/v1/summary`
- `OPENAI_MODEL`: defaut `gpt-5-nano`
- `OPENAI_BASE_URL`: defaut `https://api.openai.com/v1`
- `HTTP_TIMEOUT_SECONDS`: timeout reseau global
- `ANALYSIS_INPUT_MAX_CHARS`: taille max du texte envoye a OpenAI
- `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` / `DB_SSLMODE`: utilises par `./scripts/init_db.sh`
- `POSTGRES_ADMIN_URL`: connexion admin optionnelle pour `./scripts/init_db.sh` si vous ne voulez pas utiliser `sudo -u postgres`

Exemple complet:

```env
APP_ENV=prod
APP_PORT=9000
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-5-nano
OPENAI_BASE_URL=https://api.openai.com/v1
DATABASE_URL=postgres://tosai_app:change-me@127.0.0.1:5432/tosai?sslmode=disable
CORS_ORIGINS=https://tosai.fr,https://www.tosai.fr
HTTP_TIMEOUT_SECONDS=25
ANALYSIS_INPUT_MAX_CHARS=12000
```

## Configuration PostgreSQL

Le backend applique automatiquement le schema embarque depuis `config/database_init.sql` au demarrage, si `DATABASE_URL` est defini et valide.

Le plus simple est maintenant:

```bash
cd /opt/tosai
make db-init
```

Par defaut, le script utilise:

```env
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=tosai
DB_USER=tosai_app
DB_PASSWORD=change-me
DB_SSLMODE=disable
```

Workflow:

- si `backend/.env` manque, il est cree depuis `.env.example`
- le script cree/met a jour le role applicatif
- le script cree la base si besoin
- le schema SQL est applique
- `DATABASE_URL` est ecrit dans `backend/.env`

Le script essaie d'abord `POSTGRES_ADMIN_URL` si vous l'avez defini, sinon il passe par `sudo -u postgres psql`.

Si la connexion echoue au demarrage du backend, le serveur continue en mode API-only. Cela permet un mode sans DB, mais aucune persistance PostgreSQL n'est alors disponible.

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
curl -X POST http://127.0.0.1:9000/v1/summary \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/terms"}'
```

### GET

```bash
curl 'http://127.0.0.1:9000/v1/summary?url=https://example.com/terms'
```

### Racine API

```bash
curl 'http://127.0.0.1:9000/v1/'
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
- chaque erreur API renvoie maintenant un `request_id`; utilisez-le pour retrouver la ligne `api_error ...` dans les logs backend

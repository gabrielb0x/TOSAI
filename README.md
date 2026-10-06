# 🚀 TOSAI

[![TOSAI en 15 secondes](promo/tosai-promo.gif)](promo/tosai-promo.mp4)

**TOSAI** est une plateforme pour **analyser des CGU/ToS avec l'IA** et obtenir rapidement:
- **une note globale (A→E)**
- **un resume clair**
- **les points cles**
- **les risques detectes**
- **une recommendation actionnable**

Le frontend est maintenant organise en 2 pages:
- **`/`**: page de presentation premium (style vitrine moderne)
- **`/tosai`**: page outil pour lancer les analyses

---

## ✨ Fonctionnalites

- **API Go (Gin)**: endpoint principal `POST /v1/summary` sur `https://api.tosai.fr`
- **Extraction de contenu** depuis une URL cible
- **Cache PostgreSQL par URL** avant tout appel OpenAI
- **Analyse OpenAI en 2 temps**: web research puis analyse finale JSON stricte
- **Rate limit**: 1 demande par minute et par IP
- **Frontend Vite** sombre, anime, responsive

---

## 🧱 Stack

- **Backend**: Go 1.22+ (`backend/`)
- **Frontend**: Vite + HTML/CSS/JS (`frontend/`)
- **Database**: PostgreSQL 15+ (optionnelle en mode dev API-only)
- **Reverse proxy**: NGINX avec vhosts separes frontend/API (exemple fourni: `nginx.example.conf`)

---

## 📁 Arborescence utile

- `backend/` → serveur API, config, schema SQL
- `backend/README.MD` → guide de mise en place du backend, DB et service systemd
- `frontend/` → application web (vitrine + `/tosai`)
- `nginx.example.conf` → configuration NGINX de reference
- `promo/` → video de presentation 15 s (`tosai-promo.mp4`) et son code d'animation
- `Makefile` → commandes rapides (`setup`, `dev-backend`, `dev-frontend`, `test`)

---

## ⚡ Demarrage rapide (local)

### 1) Preparation

```bash
cd /opt/tosai
make setup
```

Puis editez `backend/.env` et renseignez au minimum:

```env
OPENAI_API_KEY=sk-...
```

Si vous voulez la base locale tout de suite:

```bash
make db-init
```

### 2) Lancer le backend

```bash
cd /opt/tosai
make dev-backend
```

Backend: **http://localhost:9000**

### 3) Lancer le frontend

```bash
cd /opt/tosai
make dev-frontend
```

Frontend: **http://localhost:5173**

### 4) Tester

- Vitrine: `http://localhost:5173/`
- Outil: `http://localhost:5173/tosai`
- API summary: `curl -X POST http://localhost:9000/v1/summary -H 'Content-Type: application/json' -d '{"url":"https://example.com/terms"}'`
- API root: `curl http://localhost:9000/v1/`

---

## 🏠 Guide auto-hebergement (clair et direct)

### 1) Ou placer le projet

Recommande en production:
- **Code source**: `/opt/tosai`
- **Build frontend servi par NGINX**: `/opt/tosai/frontend/dist`
- **Binaire backend**: `/opt/tosai/bin/tosai-backend`

### 2) Installer le projet

```bash
sudo mkdir -p /opt
cd /opt
sudo git clone https://github.com/gabrielb0x/TOSAI.git tosai
sudo chown -R $USER:$USER /opt/tosai
cd /opt/tosai
```

### 3) Configurer l'environnement backend

```bash
cp backend/.env.example backend/.env
nano backend/.env
```

Variables minimales:

```env
APP_ENV=prod
APP_PORT=9000
OPENAI_API_KEY=sk-...
DATABASE_URL=postgres://tosai_app:motdepasse@127.0.0.1:5432/tosai?sslmode=disable
CORS_ORIGINS=https://tosai.fr,https://www.tosai.fr
```

Pour preparer PostgreSQL automatiquement sur une machine locale:

```bash
make db-init
```

Le script `backend/scripts/init_db.sh` cree le role, la base, applique le schema et renseigne `DATABASE_URL` dans `backend/.env`.

### 4) Build backend + frontend

```bash
cd /opt/tosai/backend
go build -o /opt/tosai/bin/tosai-backend ./cmd/server

cd /opt/tosai/frontend
npm install
npm run build
```

### 5) Service systemd backend

Creer `/etc/systemd/system/tosai-backend.service`:

```ini
[Unit]
Description=TOSAI Backend
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/tosai/backend
EnvironmentFile=/opt/tosai/backend/.env
ExecStart=/opt/tosai/bin/tosai-backend
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

Activer:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now tosai-backend
sudo systemctl status tosai-backend
```

### 6) NGINX

- Utilisez `nginx.example.conf` comme base.
- Le fichier fourni separe `api.tosai.fr` (backend Go sur `127.0.0.1:9000`) et `tosai.fr`/`www.tosai.fr` (frontend statique).
- Le `try_files ... /index.html;` est indispensable pour supporter **`/tosai`**.

Exemple d'installation:

```bash
sudo cp /opt/tosai/nginx.example.conf /etc/nginx/sites-available/tosai
sudo ln -sf /etc/nginx/sites-available/tosai /etc/nginx/sites-enabled/tosai
sudo nginx -t
sudo systemctl reload nginx
```

---

## 🧪 Commandes utiles

```bash
# Installer dependances + .env
make setup

# Developpement
make dev-backend
make dev-frontend

# Verification rapide
make test
```

---

## 🔐 Variables d'environnement backend

Fichier prioritaire: `backend/.env`.

- `APP_ENV` (`local` ou `prod`)
- `APP_PORT` (defaut `9000`)
- `DATABASE_URL` (optionnelle mais recommandee en prod)
- `CORS_ORIGINS` (liste CSV ou `*`)
- `ADMIN_API_TOKEN`
- `OPENAI_API_KEY` (**obligatoire pour l'analyse**)
- `OPENAI_MODEL` (defaut `gpt-5.4-mini`)
- `OPENAI_RESEARCH_MODEL` (optionnel, phase web research)
- `OPENAI_BASE_URL` (defaut `https://api.openai.com/v1`)
- `HTTP_TIMEOUT_SECONDS` (defaut `45`)
- `ANALYSIS_INPUT_MAX_CHARS` (defaut `12000`)
- `ANALYSIS_CACHE_MAX_AGE_DAYS` (defaut `90`)
- `ANALYSIS_RATE_LIMIT_PER_MINUTE` (defaut `1`)
- `API_DEBUG_MODE` (`true` = responses API tres verbeuses cote backend)
- `TRUSTED_PROXIES` (defaut `127.0.0.1,::1`)
- `VITE_BACKEND_PROXY_TARGET` (frontend dev, defaut `http://localhost:9000`)
- `VITE_MASK_BUILD_FILENAMES` (frontend build, `true` = noms de fichiers hashes sans prefixe, ex `assets/abc123.css`)

---

## ✅ Etat du frontend

- **`/`**: page de presentation moderne
- **`/tosai`**: page outil d'analyse
- **Theme sombre**, animations, et typographie **Poppins (600/800)**

---

## 🛟 Depannage rapide

- Si `/tosai` ne charge pas en prod: verifier le `try_files` NGINX vers `/index.html`.
- Si la DB echoue au demarrage: tester `DATABASE_URL` avec `psql` et verifier les droits `CREATE/ALTER`.
- Si `/v1/summary` renvoie une erreur: regardez le `request_id` dans la reponse puis retrouvez la ligne `api_error` correspondante dans les logs backend.
- Si l'API renvoie `rate_limited`: attendez 1 minute ou ajustez `ANALYSIS_RATE_LIMIT_PER_MINUTE`.

---

## 🤝 Licence

Projet sous licence **MIT** (voir `LICENSE`).

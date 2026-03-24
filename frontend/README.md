# 🎨 Frontend TOSAI (Vite vanilla)

Frontend moderne en **HTML/CSS/JS pur** avec:
- **`/`**: page de presentation
- **`/tosai`**: page outil d'analyse

## 🚀 Lancer en dev

```bash
npm install
npm run dev -- --host --port 5173
```

- URL locale: `http://localhost:5173`
- Proxy API actif en dev sur `/v1` vers `http://localhost:9000`

## ⚙️ Variables optionnelles

- `VITE_API_BASE_URL` -> base API explicite (ex: `https://api.tosai.fr`)
- `VITE_BACKEND_PROXY_TARGET` -> cible proxy Vite (defaut `http://localhost:9000`)
- `VITE_MASK_BUILD_FILENAMES` -> `true` pour masquer les noms en build

Sans `VITE_API_BASE_URL`, le frontend utilise automatiquement `https://api.tosai.fr` lorsqu'il tourne sur `tosai.fr` ou `www.tosai.fr`.

L'outil appelle l'endpoint versionne `https://api.tosai.fr/v1/summary`.

Exemple:

```bash
VITE_MASK_BUILD_FILENAMES=true npm run build
```

- `false` (defaut): `assets/index-abc123.css`
- `true`: `assets/abc123.css`

## 🧩 Build production

```bash
npm run build
```

Le build statique est genere dans `frontend/dist/`.

# Vite + HTML/CSS/JS (vanilla)

Ce frontend utilise Vite avec du HTML, du CSS et du JavaScript pur (sans React/TypeScript).

## Démarrer

```bash
npm install
npm run dev -- --host --port 5173
```

Par défaut, Vite proxy `/api` et `/healthz` vers `http://localhost:9000`, donc aucune variable front n'est requise en local.

## Variables optionnelles

- `VITE_API_BASE_URL` : base API explicite (ex: `https://api.tosai.fr`).
- `VITE_BACKEND_PROXY_TARGET` : cible du proxy Vite en dev (défaut `http://localhost:9000`).

## Paramètres faciles

Les paramètres clés sont regroupés dans `src/main.js` (objet `CONFIG`) pour pouvoir
changer rapidement l'URL de l'API, l'endpoint ou le timeout.

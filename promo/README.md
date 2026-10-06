# 🎬 Promo TOSAI (15 s)

Vidéo motion design 1920x1080 / 60 fps : `tosai-promo.mp4` (aperçu GIF : `tosai-promo.gif`).

La vidéo n'est pas montée dans un logiciel : c'est une page HTML animée avec
**GSAP**, filmée image par image par Chromium (Playwright) puis encodée avec
**ffmpeg**. Donc tu la modifies comme du code front.

## Fichiers

- `scene.html` → les éléments (textes, barre de recherche, carte de résultat…)
- `scene.css` → le style (mêmes couleurs/police que `frontend/src/style.css`)
- `scene.js` → **toute l'animation** : une seule timeline GSAP, chaque ligne
  `tl.to(cible, {...}, temps)` dit quoi animer et à quelle seconde
- `render.mjs` → serveur local + capture + encodage

## Déroulé

| Temps | Scène |
| --- | --- |
| 0 – 2.5 s | Mur de CGU qui défile, « Personne ne lit les CGU. », 18 min de lecture |
| 2.5 – 4.3 s | Tout est aspiré → logo + anneaux Google, « TOSAI », tagline |
| 4.3 – 11 s | Démo : 1. coller le lien (Ctrl+V) → 2. cliquer Analyser → 3. note B + points |
| 11 – 12.9 s | « Colle. Clique. Compris. » |
| 12.9 – 15 s | Logo + tosai.fr |

## Commandes

Prérequis : Node 18+, ffmpeg dans le PATH.

```bash
cd promo
npm install
npx playwright install chromium   # une seule fois

npm run preview                    # http://localhost:4173 → lecture en temps réel
                                   # (?t=8 pour démarrer à 8 s)
node render.mjs --still 3,9.5      # captures PNG dans frames/ pour vérifier un instant
npm run render                     # → tosai-promo.mp4
```

Options de `render.mjs` : `--fps 30`, `--blur 1` (sans flou de mouvement, 2x plus
rapide), `--blur 4` (flou plus marqué), `--out autre.mp4`, `--workers 3` (nombre de
Chromium en parallèle, par défaut nb de cœurs - 1), `--keep-frames` (garde les PNG
dans `frames/video/`). Si tu as déjà un Chrome/Chromium :
`CHROMIUM_PATH=/chemin/chrome npm run render`.

## Aperçu GIF du README

GitHub ne lit pas un mp4 du repo directement dans un README, donc le README affiche
`tosai-promo.gif` (cliquable vers le mp4). Pour le regénérer après un nouveau rendu
(il faut aussi `gifsicle`) :

```bash
ffmpeg -i tosai-promo.mp4 -filter_complex "fps=12,scale=800:-1:flags=lanczos,hqdn3d=4:4:8:8,split[a][b];[a]palettegen=stats_mode=diff[p];[b][p]paletteuse=dither=bayer:bayer_scale=5:diff_mode=rectangle" -loop 0 tmp.gif
gifsicle -O3 --lossy=80 tmp.gif -o tosai-promo.gif && rm tmp.gif
```

## Modifier

- **Un texte** → `scene.html`.
- **Un timing** → le dernier argument des `tl.to(...)` dans `scene.js`
  (en secondes). Pense à décaler ce qui suit dans la même scène.
- **La durée totale** → `DURATION` en haut de `scene.js`.
- **Le côté « smooth »** vient des easings `expo.out` / `expo.inOut` (départ
  rapide, arrivée très douce) et du flou de mouvement du rendu (`--blur`).

Pas de son dans le mp4 : ajoute ta musique/tes whooshs au montage (CapCut,
Premiere…). Les gros temps forts tombent à 2.5 s, 7.2 s, 11.2 s et 12.9 s.

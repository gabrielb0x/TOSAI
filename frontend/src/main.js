import './style.css'
import logoUrl from './logo.webp'

const CONFIG = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? '',
  endpoint: '/api/v1/summary',
  defaultUrl: 'https://openai.com/policies/privacy-policy/',
  labels: {
    idle: 'Pret',
    loading: 'Analyse en cours...',
    success: 'Analyse terminee',
    error: 'Erreur de traitement',
  },
  requestTimeoutMs: 25000,
  emptySummary: 'Aucun resume pour le moment.',
}

const ROUTES = {
  home: '/',
  tool: '/tosai',
}

const app = document.getElementById('app')

let cleanupFns = []
let routerInstalled = false

const normalizePath = (pathname) => {
  const safePath = (pathname || '/').replace(/\/+$/, '') || '/'
  if (safePath === ROUTES.tool || safePath.startsWith(`${ROUTES.tool}/`)) {
    return ROUTES.tool
  }
  return ROUTES.home
}

const addCleanup = (fn) => {
  cleanupFns.push(fn)
}

const runCleanup = () => {
  cleanupFns.forEach((fn) => fn())
  cleanupFns = []
}

const setLogoSources = () => {
  document.querySelectorAll('[data-logo]').forEach((img) => {
    img.src = logoUrl
  })
}

const setYear = () => {
  document.querySelectorAll('[data-year]').forEach((node) => {
    node.textContent = String(new Date().getFullYear())
  })
}

const setupReveal = () => {
  const items = [...document.querySelectorAll('[data-reveal]')]
  if (items.length === 0) {
    return
  }

  if (!('IntersectionObserver' in window)) {
    items.forEach((item) => item.classList.add('is-visible'))
    return
  }

  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('is-visible')
          observer.unobserve(entry.target)
        }
      })
    },
    {
      threshold: 0.16,
      rootMargin: '0px 0px -8% 0px',
    },
  )

  items.forEach((item) => observer.observe(item))
  addCleanup(() => observer.disconnect())
}

const homeTemplate = () => `
  <div class="layout">
    <div class="bg-aura bg-aura-home" aria-hidden="true"></div>
    <div class="bg-grid" aria-hidden="true"></div>

    <header class="topbar">
      <a href="/" data-nav class="brand" aria-label="Accueil TOSAI">
        <img data-logo alt="Logo TOSAI" class="brand-logo" />
        <span>TOSAI</span>
      </a>

      <nav class="nav-links" aria-label="Navigation principale">
        <a href="#how">Fonctionnement</a>
        <a href="#features">Avantages</a>
        <a href="/tosai" data-nav class="btn btn-primary btn-sm">Utiliser TOSAI</a>
      </nav>
    </header>

    <main>
      <section class="hero hero-home" data-reveal>
        <div class="hero-copy">
          <p class="eyebrow">Product Landing</p>
          <h1>Les CGU deviennent enfin lisibles.</h1>
          <p>
            TOSAI analyse une page de conditions d'utilisation et la transforme en lecture claire:
            note globale, zones sensibles, synthese courte et recommendation utile.
          </p>
          <div class="hero-inline-metrics">
            <div>
              <strong>1 URL</strong>
              <span>entree simple</span>
            </div>
            <div>
              <strong>A -> E</strong>
              <span>note lisible</span>
            </div>
            <div>
              <strong>JSON</strong>
              <span>sortie exploitable</span>
            </div>
          </div>
          <div class="hero-actions">
            <a href="/tosai" data-nav class="btn btn-primary">Ouvrir l'analyseur</a>
            <a href="#features" class="btn btn-ghost">Voir la plateforme</a>
          </div>
        </div>

        <aside class="hero-side hero-preview card">
          <div class="preview-head">
            <div class="preview-brand">
              <img data-logo alt="Logo TOSAI" class="preview-logo" />
              <div>
                <span class="preview-label">TOSAI Snapshot</span>
                <strong>Rapport instantane</strong>
              </div>
            </div>
            <span class="preview-badge">Live</span>
          </div>
          <div class="preview-grade-row">
            <div>
              <span class="muted-cap">Note</span>
              <strong class="preview-grade">B</strong>
            </div>
            <div class="preview-copy">
              <p>Le document est lisible mais comporte plusieurs clauses a surveiller.</p>
            </div>
          </div>
          <div class="preview-points">
            <article>
              <span>Points cles</span>
              <strong>3 resumes</strong>
            </article>
            <article>
              <span>Risques</span>
              <strong>4 alertes</strong>
            </article>
            <article>
              <span>Format</span>
              <strong>JSON + UI</strong>
            </article>
          </div>
        </aside>
      </section>

      <section class="metrics-row" data-reveal>
        <article class="metric-card card">
          <span class="muted-cap">Experience</span>
          <strong>Interface claire</strong>
          <p>Concue pour lire vite, sans surcharge visuelle.</p>
        </article>
        <article class="metric-card card">
          <span class="muted-cap">Analyse</span>
          <strong>Sortie orientee decision</strong>
          <p>Resume, risques et recommandation dans le meme flux.</p>
        </article>
        <article class="metric-card card">
          <span class="muted-cap">Integration</span>
          <strong>Local et auto-heberge</strong>
          <p>Frontend Vite, backend Go, proxy `/api` en dev.</p>
        </article>
      </section>

      <section id="how" class="section" data-reveal>
        <div class="section-head">
          <p class="eyebrow">Workflow</p>
          <h2>Une interface simple, un resultat vraiment utile.</h2>
        </div>

        <div class="steps-grid">
          <article class="card step-card">
            <span class="step-index">1</span>
            <h3>Entrer une URL</h3>
            <p>Collez l'adresse d'une page de conditions d'utilisation, de politique ou de mentions.</p>
          </article>

          <article class="card step-card">
            <span class="step-index">2</span>
            <h3>Analyser</h3>
            <p>Le backend extrait le texte, normalise le contenu puis interroge l'analyse IA.</p>
          </article>

          <article class="card step-card">
            <span class="step-index">3</span>
            <h3>Decider</h3>
            <p>Vous obtenez un rapport utilisable tout de suite, cote interface et cote API.</p>
          </article>
        </div>
      </section>

      <section id="features" class="section" data-reveal>
        <div class="section-head">
          <p class="eyebrow">Ce que TOSAI apporte</p>
          <h2>Une vitrine nette devant, un vrai outil derriere.</h2>
        </div>

        <div class="features-grid">
          <article class="card feature-card">
            <h3>Lecture immediate</h3>
            <p>Un design plus editorial, avec du rythme, de l'espace et une meilleure hierarchie.</p>
          </article>

          <article class="card feature-card">
            <h3>Usage produit</h3>
            <p>La page `/tosai` garde le focus sur l'action: saisir, analyser, comprendre.</p>
          </article>

          <article class="card feature-card">
            <h3>Base solide</h3>
            <p>Le frontend reste compatible avec le backend local et les builds de production.</p>
          </article>
        </div>
      </section>

      <section class="showcase card" data-reveal>
        <div class="showcase-copy">
          <p class="eyebrow">Auto-hebergement</p>
          <h2>Concu pour etre lance proprement en local, puis deploie sans friction.</h2>
          <p>
            Le site de presentation met en valeur le produit. La page outil reste pratique et
            directe. On garde donc un front plus mature sans sacrifier le fonctionnement actuel.
          </p>
        </div>
        <div class="showcase-panel">
          <div class="showcase-line">
            <span>Frontend</span>
            <strong>Vite SPA</strong>
          </div>
          <div class="showcase-line">
            <span>Backend</span>
            <strong>Go API</strong>
          </div>
          <div class="showcase-line">
            <span>Dev local</span>
            <strong>Proxy `/api`</strong>
          </div>
        </div>
      </section>
    </main>

    <footer class="site-footer" data-reveal>
      <a href="/tosai" data-nav>Ouvrir la page outil</a>
      <span>•</span>
      <span>Copyright <span data-year></span> TOSAI</span>
    </footer>
  </div>
`

const toolTemplate = () => `
  <div class="layout">
    <div class="bg-aura bg-aura-tool" aria-hidden="true"></div>
    <div class="bg-grid" aria-hidden="true"></div>

    <header class="topbar">
      <a href="/" data-nav class="brand" aria-label="Accueil TOSAI">
        <img data-logo alt="Logo TOSAI" class="brand-logo" />
        <span>TOSAI</span>
      </a>

      <nav class="nav-links" aria-label="Navigation principale">
        <a href="/" data-nav>Presentation</a>
        <a href="/tosai" data-nav class="btn btn-primary btn-sm">Analyseur</a>
      </nav>
    </header>

    <main class="tool-main">
      <section class="tool-hero card" data-reveal>
        <div class="tool-hero-copy">
          <p class="eyebrow">TOSAI / Tool</p>
          <h1>Un analyseur clair, rapide et centre sur le resultat.</h1>
          <p>Collez une URL puis obtenez une note, un resume, des risques et la reponse complete.</p>
        </div>
        <aside class="tool-hero-side">
          <span class="muted-cap">Sortie</span>
          <strong>Resume + JSON brut</strong>
          <p>La meme analyse est lisible a l'ecran et reutilisable cote produit.</p>
        </aside>
      </section>

      <div class="tool-grid tool-grid-top">
        <form id="analyze-form" class="card tool-form" data-reveal>
          <label for="tos-url">URL a analyser</label>
          <div class="field-row">
            <input id="tos-url" name="tos-url" type="url" placeholder="https://example.com/terms" required />
            <button type="submit">Analyser</button>
          </div>
          <p class="hint">Backend: <strong id="api-base">-</strong> | Endpoint: <code id="api-endpoint">-</code></p>
        </form>

        <section class="card tool-info-card" data-reveal>
          <span class="muted-cap">Bon a savoir</span>
          <h2>Le front parle au backend local sans configuration supplementaire.</h2>
          <p>En dev, Vite proxy `/api` et `/healthz` vers le serveur Go.</p>
        </section>
      </div>

      <div class="tool-grid tool-grid-results">
        <section class="card" data-reveal>
          <h2>Etat de la requete</h2>
          <div class="status-grid">
            <article class="status-item">
              <span class="label">Statut</span>
              <p id="status-text" class="status-value">Pret</p>
            </article>
            <article class="status-item">
              <span class="label">HTTP</span>
              <p id="status-code" class="status-value">-</p>
            </article>
            <article class="status-item">
              <span class="label">Maj</span>
              <p id="status-time" class="status-value">-</p>
            </article>
            <article class="status-item">
              <span class="label">Confiance</span>
              <p id="confidence" class="status-value">-</p>
            </article>
          </div>
        </section>

        <section class="card summary-card" data-reveal>
          <div class="summary-head">
            <div>
              <span class="label">Note globale</span>
              <p id="rating" class="rating" data-rating="X">-</p>
            </div>
          </div>
          <p id="summary-text" class="summary-text">Aucun resume pour le moment.</p>
          <p id="recommendation" class="recommendation">Recommendation en attente.</p>
        </section>

        <section class="card split-panel list-panel" data-reveal>
          <article>
            <h3>Points cles</h3>
            <ul id="highlights-list" class="bullet-list">
              <li>En attente d'analyse.</li>
            </ul>
          </article>
          <article>
            <h3>Risques detectes</h3>
            <ul id="risks-list" class="bullet-list">
              <li>En attente d'analyse.</li>
            </ul>
          </article>
        </section>

        <section class="card" data-reveal>
          <h2>Reponse brute</h2>
          <pre id="raw-response" class="raw">En attente...</pre>
        </section>
      </div>
    </main>

    <footer class="site-footer" data-reveal>
      <a href="/" data-nav>Retour accueil</a>
      <span>•</span>
      <span>Copyright <span data-year></span> TOSAI</span>
    </footer>
  </div>
`

const render = () => {
  runCleanup()

  const route = normalizePath(window.location.pathname)
  app.innerHTML = route === ROUTES.tool ? toolTemplate() : homeTemplate()

  setLogoSources()
  setYear()
  setupReveal()

  if (route === ROUTES.tool) {
    initToolPage()
  }
}

const installRouter = () => {
  if (routerInstalled) {
    return
  }

  routerInstalled = true

  document.addEventListener('click', (event) => {
    const anchor = event.target.closest('a[data-nav]')
    if (!anchor) {
      return
    }

    const href = anchor.getAttribute('href') || ''
    if (!href.startsWith('/')) {
      return
    }

    event.preventDefault()
    const target = normalizePath(href)
    const current = normalizePath(window.location.pathname)

    if (target !== current) {
      window.history.pushState({}, '', target)
      render()
    }

    window.scrollTo({ top: 0, behavior: 'smooth' })
  })

  window.addEventListener('popstate', () => {
    render()
  })
}

const formatTime = (date = new Date()) =>
  date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit', second: '2-digit' })

const withTimeout = async (promise, timeoutMs) => {
  let timeoutId
  const timeoutPromise = new Promise((_, reject) => {
    timeoutId = setTimeout(() => reject(new Error('Delai depasse')), timeoutMs)
  })

  try {
    return await Promise.race([promise, timeoutPromise])
  } finally {
    clearTimeout(timeoutId)
  }
}

const buildSummaryUrl = () => {
  if (!CONFIG.apiBaseUrl) {
    return CONFIG.endpoint
  }
  return new URL(CONFIG.endpoint, CONFIG.apiBaseUrl).toString()
}

const toPrettyJSON = (value) => {
  if (typeof value === 'string') {
    return value
  }

  try {
    return JSON.stringify(value, null, 2)
  } catch (error) {
    return String(value)
  }
}

const extractSummary = (payload) => {
  if (!payload || typeof payload !== 'object') {
    return {
      rating: '-',
      confidence: '-',
      summary: CONFIG.emptySummary,
      recommendation: 'Recommendation indisponible.',
      highlights: [],
      risks: [],
    }
  }

  const analysis = payload.analysis && typeof payload.analysis === 'object' ? payload.analysis : payload

  return {
    rating: analysis.rating || analysis.note || analysis.grade || '-',
    confidence: analysis.confidence || payload.confidence || '-',
    summary: analysis.summary || analysis.summary_md || payload.message || CONFIG.emptySummary,
    recommendation: analysis.recommendation || 'Lire les clauses sensibles avant de continuer.',
    highlights: analysis.highlights || [],
    risks: analysis.risks || [],
  }
}

const initToolPage = () => {
  const elements = {
    form: document.getElementById('analyze-form'),
    urlInput: document.getElementById('tos-url'),
    statusText: document.getElementById('status-text'),
    statusCode: document.getElementById('status-code'),
    statusTime: document.getElementById('status-time'),
    apiBase: document.getElementById('api-base'),
    apiEndpoint: document.getElementById('api-endpoint'),
    rating: document.getElementById('rating'),
    confidence: document.getElementById('confidence'),
    summary: document.getElementById('summary-text'),
    recommendation: document.getElementById('recommendation'),
    highlights: document.getElementById('highlights-list'),
    risks: document.getElementById('risks-list'),
    raw: document.getElementById('raw-response'),
  }

  if (!elements.form) {
    return
  }

  const updateStatus = ({ label, code = '-', time = '-' }) => {
    elements.statusText.textContent = label
    elements.statusCode.textContent = code
    elements.statusTime.textContent = time
  }

  const setRating = (value = '-') => {
    const rating = String(value || '-')
      .trim()
      .toUpperCase()

    elements.rating.textContent = rating
    elements.rating.dataset.rating = ['A', 'B', 'C', 'D', 'E'].includes(rating) ? rating : 'X'
  }

  const setList = (target, items, fallback) => {
    target.innerHTML = ''
    const safeItems = Array.isArray(items) && items.length > 0 ? items : [fallback]

    safeItems.forEach((item) => {
      const li = document.createElement('li')
      li.textContent = item
      target.appendChild(li)
    })
  }

  const updateSummary = ({
    rating = '-',
    confidence = '-',
    summary = CONFIG.emptySummary,
    recommendation = 'Recommendation en attente.',
    highlights = [],
    risks = [],
  }) => {
    setRating(rating)
    elements.confidence.textContent = confidence
    elements.summary.textContent = summary
    elements.recommendation.textContent = recommendation
    setList(elements.highlights, highlights, 'Aucun point cle pour le moment.')
    setList(elements.risks, risks, 'Aucun risque detecte pour le moment.')
  }

  const setRaw = (content) => {
    elements.raw.textContent = content
  }

  const resetUI = () => {
    updateStatus({ label: CONFIG.labels.idle })
    updateSummary({})
    setRaw('En attente...')
  }

  const submitForm = async (event) => {
    event.preventDefault()

    const urlValue = elements.urlInput.value.trim()

    updateStatus({ label: CONFIG.labels.loading, time: formatTime() })
    updateSummary({
      rating: '-',
      confidence: '-',
      summary: 'Chargement...',
      recommendation: 'Analyse en cours.',
      highlights: ['Extraction du contenu...'],
      risks: ['Evaluation en cours...'],
    })
    setRaw('Requete en cours...')

    try {
      const response = await withTimeout(
        fetch(buildSummaryUrl(), {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ url: urlValue }),
        }),
        CONFIG.requestTimeoutMs,
      )

      const text = await response.text()
      let payload

      try {
        payload = JSON.parse(text)
      } catch (error) {
        payload = text
      }

      const { rating, confidence, summary, recommendation, highlights, risks } = extractSummary(payload)
      updateSummary({ rating, confidence, summary, recommendation, highlights, risks })

      updateStatus({
        label: response.ok ? CONFIG.labels.success : CONFIG.labels.error,
        code: response.status,
        time: formatTime(),
      })

      setRaw(toPrettyJSON(payload))

      if (!response.ok) {
        const message = payload && typeof payload === 'object' ? payload.message : null
        throw new Error(message || `HTTP ${response.status}`)
      }
    } catch (error) {
      updateStatus({ label: CONFIG.labels.error, code: '-', time: formatTime() })
      updateSummary({
        rating: '-',
        confidence: '-',
        summary: "Impossible de contacter l'API.",
        recommendation: 'Verifier backend, reseau et OPENAI_API_KEY.',
        highlights: ['Aucune donnee disponible.'],
        risks: ['Requete interrompue ou invalide.'],
      })
      setRaw(error instanceof Error ? error.message : String(error))
    }
  }

  elements.urlInput.value = CONFIG.defaultUrl
  elements.apiBase.textContent = CONFIG.apiBaseUrl || 'proxy Vite (localhost:9000)'
  elements.apiEndpoint.textContent = CONFIG.endpoint
  elements.form.addEventListener('submit', submitForm)
  resetUI()

  addCleanup(() => {
    elements.form.removeEventListener('submit', submitForm)
  })
}

installRouter()
render()

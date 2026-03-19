import './style.css'
import { gsap } from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

gsap.registerPlugin(ScrollTrigger)

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

let cleanupFns = []
let routerInstalled = false

const getAppRoot = () => document.getElementById('app')

const getTemplateMarkup = (id) => {
  const template = document.getElementById(id)
  if (!template) {
    throw new Error(`Template introuvable: ${id}`)
  }

  return template.innerHTML
}

const escapeHtml = (value) =>
  String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')

const prefersReducedMotion = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches

const showFatalError = (error) => {
  const app = getAppRoot()
  if (!app) {
    return
  }

  const message = error instanceof Error ? `${error.name}: ${error.message}` : String(error)
  app.innerHTML = `
    <div class="layout">
      <section class="panel fatal-error-panel">
        <p class="section-eyebrow">Erreur frontend</p>
        <h1>L'interface n'a pas pu demarrer.</h1>
        <p>Une erreur JavaScript a bloque le chargement local. Le message brut est ci-dessous.</p>
        <pre class="raw">${escapeHtml(message)}</pre>
      </section>
    </div>
  `
}

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
  ScrollTrigger.getAll().forEach((trigger) => trigger.kill())
}

const activateNavState = (route) => {
  document.querySelectorAll('[data-route]').forEach((link) => {
    link.classList.toggle('is-active', link.getAttribute('data-route') === route)
  })
}

const trackTween = (tween) => {
  addCleanup(() => {
    tween.scrollTrigger?.kill()
    tween.kill()
  })

  return tween
}

const setupMotion = () => {
  if (prefersReducedMotion()) {
    return
  }

  const nav = document.querySelector('.nav-shell')
  const heroItems = [...document.querySelectorAll('.motion-item')]
  const bubbles = [...document.querySelectorAll('.bubble')]
  const sectionItems = [...document.querySelectorAll('[data-animate]')]
  const storyRows = [...document.querySelectorAll('[data-story]')]
  const footer = document.querySelector('.site-footer')
  const scrollHint = document.querySelector('.scroll-hint')

  if (nav) {
    trackTween(
      gsap.from(nav, {
        y: -20,
        autoAlpha: 0,
        duration: 0.9,
        ease: 'power3.out',
      }),
    )
  }

  if (heroItems.length > 0) {
    trackTween(
      gsap.from(heroItems, {
        y: 44,
        autoAlpha: 0,
        duration: 1.05,
        stagger: 0.1,
        ease: 'power3.out',
        delay: 0.05,
      }),
    )
  }

  if (footer) {
    trackTween(
      gsap.from(footer, {
        y: 28,
        autoAlpha: 0,
        duration: 0.9,
        ease: 'power3.out',
      }),
    )
  }

  if (scrollHint) {
    const tween = gsap.to(scrollHint, {
      y: 8,
      duration: 1.4,
      repeat: -1,
      yoyo: true,
      ease: 'sine.inOut',
    })

    addCleanup(() => tween.kill())
  }

  bubbles.forEach((bubble, index) => {
    const tween = gsap.to(bubble, {
      x: index === 0 ? 38 : index === 1 ? -30 : 20,
      y: index === 0 ? 24 : index === 1 ? 32 : -18,
      scale: index === 2 ? 1.08 : 0.94,
      duration: 8 + index * 1.5,
      repeat: -1,
      yoyo: true,
      ease: 'sine.inOut',
    })

    addCleanup(() => tween.kill())
  })

  sectionItems.forEach((item) => {
    trackTween(
      gsap.from(item, {
        y: 56,
        autoAlpha: 0,
        duration: 1,
        ease: 'power3.out',
        scrollTrigger: {
          trigger: item,
          start: 'top 84%',
          once: true,
        },
      }),
    )
  })

  storyRows.forEach((row) => {
    const media = row.querySelector('.story-media')
    const copy = row.querySelector('.story-copy')

    if (media) {
      trackTween(
        gsap.from(media, {
          x: -72,
          autoAlpha: 0,
          duration: 1.05,
          ease: 'power3.out',
          scrollTrigger: {
            trigger: row,
            start: 'top 80%',
            once: true,
          },
        }),
      )
    }

    if (copy) {
      trackTween(
        gsap.from(copy, {
          x: 56,
          autoAlpha: 0,
          duration: 1.05,
          ease: 'power3.out',
          scrollTrigger: {
            trigger: row,
            start: 'top 80%',
            once: true,
          },
        }),
      )
    }
  })

  ScrollTrigger.refresh()
}

const render = () => {
  const app = getAppRoot()
  if (!app) {
    return
  }

  runCleanup()

  const route = normalizePath(window.location.pathname)
  const templateId = route === ROUTES.tool ? 'tool-template' : 'home-template'

  app.innerHTML = getTemplateMarkup(templateId)

  activateNavState(route)
  setupMotion()

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

const startApp = () => {
  installRouter()
  render()
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

window.addEventListener('error', (event) => {
  if (event.error) {
    showFatalError(event.error)
  } else if (event.message) {
    showFatalError(event.message)
  }
})

window.addEventListener('unhandledrejection', (event) => {
  showFatalError(event.reason || 'Promesse rejetee sans details.')
})

const bootApp = () => {
  try {
    startApp()
  } catch (error) {
    showFatalError(error)
    throw error
  }
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', bootApp, { once: true })
} else {
  bootApp()
}

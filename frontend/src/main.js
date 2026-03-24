import './style.css'
import { gsap } from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

gsap.registerPlugin(ScrollTrigger)

const DEFAULT_API_BASE_URL = 'https://api.tosai.fr'
const PUBLIC_API_HOSTNAMES = new Set(['tosai.fr', 'www.tosai.fr', 'api.tosai.fr'])

const sanitizeBaseUrl = (value) => String(value || '').trim().replace(/\/+$/, '')

const resolveApiBaseUrl = () => {
  const explicitBaseUrl = sanitizeBaseUrl(import.meta.env.VITE_API_BASE_URL)
  if (explicitBaseUrl) {
    return explicitBaseUrl
  }

  if (typeof window === 'undefined') {
    return ''
  }

  return PUBLIC_API_HOSTNAMES.has(window.location.hostname) ? DEFAULT_API_BASE_URL : ''
}

const CONFIG = {
  apiBaseUrl: resolveApiBaseUrl(),
  endpoint: '/v1/summary',
  defaultUrl: '',
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
  legal: '/mentions-legales',
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

  if (safePath === ROUTES.legal || safePath.startsWith(`${ROUTES.legal}/`)) {
    return ROUTES.legal
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

const setupMenus = () => {
  const header = document.querySelector('.topbar')
  const toggle = document.querySelector('[data-menu-toggle]')
  const panel = document.querySelector('[data-menu-panel]')

  if (!header || !toggle || !panel) {
    return
  }

  const setOpen = (open) => {
    header.classList.toggle('is-open', open)
    toggle.setAttribute('aria-expanded', String(open))
  }

  const toggleMenu = () => {
    setOpen(!header.classList.contains('is-open'))
  }

  const closeMenu = () => {
    setOpen(false)
  }

  const handleResize = () => {
    if (window.innerWidth > 760) {
      closeMenu()
    }
  }

  const links = [...panel.querySelectorAll('a')]

  toggle.addEventListener('click', toggleMenu)
  links.forEach((link) => link.addEventListener('click', closeMenu))
  window.addEventListener('resize', handleResize)

  addCleanup(() => {
    toggle.removeEventListener('click', toggleMenu)
    links.forEach((link) => link.removeEventListener('click', closeMenu))
    window.removeEventListener('resize', handleResize)
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
        y: -16,
        autoAlpha: 0,
        duration: 0.8,
        ease: 'power3.out',
      }),
    )
  }

  if (heroItems.length > 0) {
    trackTween(
      gsap.from(heroItems, {
        y: 36,
        autoAlpha: 0,
        duration: 0.95,
        stagger: 0.08,
        ease: 'power3.out',
        delay: 0.04,
      }),
    )
  }

  if (footer) {
    trackTween(
      gsap.from(footer, {
        y: 20,
        autoAlpha: 0,
        duration: 0.85,
        ease: 'power3.out',
      }),
    )
  }

  if (scrollHint) {
    const tween = gsap.to(scrollHint, {
      y: 6,
      duration: 1.3,
      repeat: -1,
      yoyo: true,
      ease: 'sine.inOut',
    })

    addCleanup(() => tween.kill())
  }

  bubbles.forEach((bubble, index) => {
    const tween = gsap.to(bubble, {
      x: index === 0 ? 30 : index === 1 ? -24 : 16,
      y: index === 0 ? 18 : index === 1 ? 26 : -14,
      scale: index === 2 ? 1.06 : 0.95,
      duration: 7 + index * 1.5,
      repeat: -1,
      yoyo: true,
      ease: 'sine.inOut',
    })

    addCleanup(() => tween.kill())
  })

  sectionItems.forEach((item) => {
    trackTween(
      gsap.from(item, {
        y: 44,
        autoAlpha: 0,
        duration: 0.9,
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
          x: -56,
          autoAlpha: 0,
          duration: 0.95,
          ease: 'power3.out',
          scrollTrigger: {
            trigger: row,
            start: 'top 82%',
            once: true,
          },
        }),
      )
    }

    if (copy) {
      trackTween(
        gsap.from(copy, {
          x: 42,
          autoAlpha: 0,
          duration: 0.95,
          ease: 'power3.out',
          scrollTrigger: {
            trigger: row,
            start: 'top 82%',
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
  const templateId =
    route === ROUTES.tool ? 'tool-template' : route === ROUTES.legal ? 'legal-template' : 'home-template'

  app.innerHTML = getTemplateMarkup(templateId)

  activateNavState(route)
  setupMenus()
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

const normalizeAnalyzedDomain = (value) => {
  const rawValue = String(value || '').trim()
  if (!rawValue) {
    return ''
  }

  const withScheme = /^[a-z][a-z0-9+.-]*:\/\//i.test(rawValue) ? rawValue : `https://${rawValue}`

  try {
    const parsed = new URL(withScheme)
    return parsed.hostname.trim().toLowerCase()
  } catch {
    return rawValue
      .replace(/^[a-z][a-z0-9+.-]*:\/\//i, '')
      .split(/[/?#]/, 1)[0]
      .trim()
      .toLowerCase()
  }
}

class ApiRequestError extends Error {
  constructor(message, { code = '', requestId = '', detailItems = [], status = 0 } = {}) {
    super(message)
    this.name = 'ApiRequestError'
    this.code = code
    this.requestId = requestId
    this.detailItems = detailItems
    this.status = status
  }
}

const humanizeKey = (value) =>
  String(value || '')
    .replaceAll('_', ' ')
    .trim()

const stringifyDetailValue = (value) => {
  if (value == null) {
    return ''
  }

  if (Array.isArray(value)) {
    return value.map((item) => stringifyDetailValue(item)).filter(Boolean).join(', ')
  }

  if (typeof value === 'object') {
    return Object.entries(value)
      .map(([key, item]) => {
        const formatted = stringifyDetailValue(item)
        return formatted ? `${humanizeKey(key)}: ${formatted}` : ''
      })
      .filter(Boolean)
      .join(', ')
  }

  return String(value).trim()
}

const extractApiErrorDetailItems = (payload, status) => {
  const items = []

  if (payload && typeof payload === 'object') {
    if (payload.code) {
      items.push(`Code API: ${payload.code}`)
    }

    if (payload.details && typeof payload.details === 'object') {
      Object.entries(payload.details).forEach(([key, value]) => {
        const formatted = stringifyDetailValue(value)
        if (formatted) {
          items.push(`${humanizeKey(key)}: ${formatted}`)
        }
      })
    }
  }

  if (items.length === 0 && status) {
    items.push(`HTTP ${status}`)
  }

  return items
}

const buildApiRequestError = (response, payload) => {
  const message =
    payload && typeof payload === 'object' && typeof payload.message === 'string' && payload.message.trim()
      ? payload.message.trim()
      : `HTTP ${response.status}`

  return new ApiRequestError(message, {
    code: payload && typeof payload === 'object' && payload.code ? String(payload.code) : '',
    requestId: payload && typeof payload === 'object' && payload.request_id ? String(payload.request_id) : '',
    detailItems: extractApiErrorDetailItems(payload, response.status),
    status: response.status,
  })
}

const extractSummary = (payload) => {
  if (!payload || typeof payload !== 'object') {
    return {
      rating: '-',
      summary: CONFIG.emptySummary,
      highlights: [],
      risks: [],
      isContestable: false,
    }
  }

  const analysis = payload.analysis && typeof payload.analysis === 'object' ? payload.analysis : payload

  return {
    rating: analysis.rating || analysis.note || analysis.grade || '-',
    summary: analysis.summary || analysis.summary_md || payload.message || CONFIG.emptySummary,
    highlights: analysis.highlights || [],
    risks: analysis.risks || [],
    isContestable: Boolean(analysis.is_contestable),
  }
}

const initToolPage = () => {
  const form = document.getElementById('analyze-form')
  const urlInput = document.getElementById('tos-url')
  const rating = document.getElementById('rating')
  const summary = document.getElementById('summary-text')
  const highlights = document.getElementById('highlights-list')
  const risks = document.getElementById('risks-list')
  const retryButton = document.getElementById('retry-analysis')
  const humanButton = document.getElementById('human-check')
  const humanFeedback = document.getElementById('human-feedback')
  const submitButton = form?.querySelector('button[type="submit"]')

  if (!form || !urlInput || !rating || !summary || !highlights || !risks || !submitButton) {
    return
  }

  let retryAllowed = false

  const setRating = (value = '-') => {
    const normalized = String(value || '-')
      .trim()
      .toUpperCase()

    rating.textContent = normalized
    rating.dataset.rating = ['A', 'B', 'C', 'D', 'E'].includes(normalized) ? normalized : 'X'
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

  const updateResult = ({
    ratingValue = '-',
    summaryValue = CONFIG.emptySummary,
    highlightsValue = [],
    risksValue = [],
    isContestableValue = false,
  }) => {
    retryAllowed = Boolean(isContestableValue)
    setRating(ratingValue)
    summary.textContent = summaryValue
    setList(highlights, highlightsValue, 'Les bons points apparaitront ici apres analyse.')
    setList(risks, risksValue, 'Les points sensibles apparaitront ici apres analyse.')
    if (retryButton) {
      retryButton.disabled = submitButton.disabled || !retryAllowed
      retryButton.setAttribute('aria-disabled', String(retryButton.disabled))
    }
  }

  const setLoadingState = (loading) => {
    submitButton.disabled = loading
    submitButton.classList.toggle('is-busy', loading)
    submitButton.setAttribute('aria-disabled', String(loading))
    submitButton.textContent = loading ? 'Analyse...' : 'Analyser'

    if (retryButton) {
      retryButton.disabled = loading || !retryAllowed
      retryButton.classList.toggle('is-busy', loading)
      retryButton.setAttribute('aria-disabled', String(retryButton.disabled))
    }
  }

  const showHumanFeedback = (message) => {
    if (!humanFeedback) {
      return
    }

    humanFeedback.hidden = false
    humanFeedback.textContent = message
  }

  const hideHumanFeedback = () => {
    if (!humanFeedback) {
      return
    }

    humanFeedback.hidden = true
    humanFeedback.textContent = ''
  }

  const resetUI = () => {
    updateResult({})
    hideHumanFeedback()
  }

  const runAnalysis = async ({ forceRefresh = false } = {}) => {
    const normalizedDomain = normalizeAnalyzedDomain(urlInput.value)
    if (!normalizedDomain) {
      urlInput.focus()
      return
    }
    urlInput.value = normalizedDomain

    hideHumanFeedback()
    setLoadingState(true)
    updateResult({
      ratingValue: '-',
      summaryValue: 'Analyse en cours...',
      highlightsValue: ['Lecture de la page en cours...'],
      risksValue: ['Detection des points sensibles en cours...'],
    })

    try {
      const response = await withTimeout(
        fetch(buildSummaryUrl(), {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ url: normalizedDomain, force_refresh: forceRefresh }),
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

      if (!response.ok) {
        throw buildApiRequestError(response, payload)
      }

      const {
        rating: ratingValue,
        summary: summaryValue,
        highlights: highlightsValue,
        risks: risksValue,
        isContestable: isContestableValue,
      } = extractSummary(payload)

      updateResult({
        ratingValue,
        summaryValue,
        highlightsValue,
        risksValue,
        isContestableValue,
      })
    } catch (error) {
      const message = error instanceof Error && error.message ? error.message : "Impossible de contacter l'API."
      const detailItems =
        error instanceof ApiRequestError && error.detailItems.length > 0
          ? error.detailItems
          : ['La requete a echoue ou le backend est indisponible.']
      const highlightItems = ['Verifie les logs du backend si le probleme persiste.']

      if (error instanceof ApiRequestError && error.status) {
        highlightItems.unshift(`HTTP ${error.status}`)
      }

      updateResult({
        ratingValue: '-',
        summaryValue: message,
        highlightsValue: highlightItems,
        risksValue: detailItems,
        isContestableValue: false,
      })

      if (error instanceof ApiRequestError && error.requestId) {
        showHumanFeedback(`Request ID backend: ${error.requestId}`)
      }
    } finally {
      setLoadingState(false)
    }
  }

  const submitForm = (event) => {
    event.preventDefault()
    void runAnalysis({ forceRefresh: false })
  }

  const normalizeInputValue = () => {
    const normalizedDomain = normalizeAnalyzedDomain(urlInput.value)
    if (normalizedDomain) {
      urlInput.value = normalizedDomain
    }
  }

  const retryAnalysis = () => {
    if (!retryAllowed) {
      return
    }

    hideHumanFeedback()
    void runAnalysis({ forceRefresh: true })
  }

  const requestHumanCheck = () => {
    showHumanFeedback('La verification humaine sera disponible lors de l ouverture publique du service.')
  }

  urlInput.value = CONFIG.defaultUrl
  form.addEventListener('submit', submitForm)
  urlInput.addEventListener('blur', normalizeInputValue)
  retryButton?.addEventListener('click', retryAnalysis)
  humanButton?.addEventListener('click', requestHumanCheck)
  resetUI()

  addCleanup(() => {
    form.removeEventListener('submit', submitForm)
    urlInput.removeEventListener('blur', normalizeInputValue)
    retryButton?.removeEventListener('click', retryAnalysis)
    humanButton?.removeEventListener('click', requestHumanCheck)
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

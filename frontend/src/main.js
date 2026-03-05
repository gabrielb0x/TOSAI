import './style.css'

const CONFIG = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? '',
  endpoints: {
    summary: '/api/v1/summary',
  },
  defaults: {
    url: 'https://openai.com/policies/privacy-policy/',
  },
  labels: {
    idle: 'Pret',
    loading: 'Analyse en cours...',
    success: 'Analyse terminee',
    error: 'Erreur de traitement',
  },
  requestTimeoutMs: 15000,
  emptySummary: 'Aucun resume pour le moment.',
}

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

const formatTime = (date = new Date()) =>
  date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit', second: '2-digit' })

const updateStatus = ({ label, code = '—', time = '—' }) => {
  elements.statusText.textContent = label
  elements.statusCode.textContent = code
  elements.statusTime.textContent = time
}

const setRating = (value = '-') => {
  const rating = String(value || '-').trim().toUpperCase()
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
  recommendation = 'Recommandation en attente.',
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
    return CONFIG.endpoints.summary
  }
  return new URL(CONFIG.endpoints.summary, CONFIG.apiBaseUrl).toString()
}

const extractSummary = (payload) => {
  if (!payload || typeof payload !== 'object') {
    return {
      rating: '-',
      confidence: '-',
      summary: CONFIG.emptySummary,
      recommendation: 'Recommandation indisponible.',
      highlights: [],
      risks: [],
    }
  }
  const analysis = payload.analysis && typeof payload.analysis === 'object' ? payload.analysis : payload
  return {
    rating: analysis.rating || analysis.note || analysis.grade || '-',
    confidence: analysis.confidence || payload.confidence || '-',
    summary: analysis.summary || analysis.summary_md || payload.message || CONFIG.emptySummary,
    recommendation: analysis.recommendation || 'Lire attentivement les clauses sensibles avant de continuer.',
    highlights: analysis.highlights || [],
    risks: analysis.risks || [],
  }
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
  setRaw('Requête en cours...')

  const target = buildSummaryUrl()

  try {
    const response = await withTimeout(
      fetch(target, {
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

const init = () => {
  elements.urlInput.value = CONFIG.defaults.url
  elements.apiBase.textContent = CONFIG.apiBaseUrl || 'proxy Vite (localhost:9000)'
  elements.apiEndpoint.textContent = CONFIG.endpoints.summary
  elements.form.addEventListener('submit', submitForm)
  resetUI()
}

init()

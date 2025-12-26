import './style.css'

const CONFIG = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:9000',
  endpoints: {
    summary: '/api/v1/summary',
  },
  defaults: {
    url: 'https://example.com/terms',
  },
  labels: {
    idle: 'Prêt',
    loading: "Analyse en cours...",
    success: 'Réponse reçue',
    error: 'Erreur lors de la requête',
  },
  requestTimeoutMs: 15000,
  emptySummary: 'Aucun résumé pour le moment.',
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
  summary: document.getElementById('summary-text'),
  raw: document.getElementById('raw-response'),
}

const formatTime = (date = new Date()) =>
  date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit', second: '2-digit' })

const updateStatus = ({ label, code = '—', time = '—' }) => {
  elements.statusText.textContent = label
  elements.statusCode.textContent = code
  elements.statusTime.textContent = time
}

const updateSummary = ({ rating = '—', summary = CONFIG.emptySummary }) => {
  elements.rating.textContent = rating
  elements.summary.textContent = summary
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
    timeoutId = setTimeout(() => reject(new Error('Délai dépassé')), timeoutMs)
  })
  try {
    return await Promise.race([promise, timeoutPromise])
  } finally {
    clearTimeout(timeoutId)
  }
}

const buildSummaryUrl = (tosUrl) => {
  const url = new URL(CONFIG.endpoints.summary, CONFIG.apiBaseUrl)
  if (tosUrl) {
    url.searchParams.set('url', tosUrl)
  }
  return url
}

const extractSummary = (payload) => {
  if (!payload || typeof payload !== 'object') {
    return { rating: '—', summary: CONFIG.emptySummary }
  }
  const rating = payload.rating || payload.note || payload.grade || '—'
  const summary = payload.summary || payload.summary_md || payload.message || CONFIG.emptySummary
  return { rating, summary }
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
  updateSummary({ rating: '—', summary: 'Chargement...' })
  setRaw('Requête en cours...')

  const target = buildSummaryUrl(urlValue)

  try {
    const response = await withTimeout(fetch(target), CONFIG.requestTimeoutMs)
    const text = await response.text()
    let payload
    try {
      payload = JSON.parse(text)
    } catch (error) {
      payload = text
    }

    const { rating, summary } = extractSummary(payload)
    updateSummary({ rating, summary })

    updateStatus({
      label: response.ok ? CONFIG.labels.success : CONFIG.labels.error,
      code: response.status,
      time: formatTime(),
    })

    setRaw(toPrettyJSON(payload))
  } catch (error) {
    updateStatus({ label: CONFIG.labels.error, code: '—', time: formatTime() })
    updateSummary({ rating: '—', summary: 'Impossible de contacter l’API.' })
    setRaw(error instanceof Error ? error.message : String(error))
  }
}

const init = () => {
  elements.urlInput.value = CONFIG.defaults.url
  elements.apiBase.textContent = CONFIG.apiBaseUrl
  elements.apiEndpoint.textContent = CONFIG.endpoints.summary
  elements.form.addEventListener('submit', submitForm)
  resetUI()
}

init()

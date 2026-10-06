/*
 * Promo TOSAI - 15 s, 1920x1080.
 * Tout est pilote par UNE timeline GSAP en pause (`tl`) : le rendu video
 * appelle window.__seek(t) image par image, la preview la joue en temps reel.
 * Pour modifier le montage : chaque scene a ses temps en secondes (2e argument
 * des tl.to / tl.fromTo), il suffit de les decaler.
 */
;(async () => {
  const DURATION = 15
  const params = new URLSearchParams(location.search)
  const RENDER = params.has('render')

  const $ = (sel) => document.querySelector(sel)
  const $$ = (sel) => [...document.querySelectorAll(sel)]

  await document.fonts.ready
  await Promise.all($$('img').map((img) => img.decode().catch(() => {})))

  // ---------- Mur de CGU (scene 1) ----------
  const clauses = [
    "L'Utilisateur reconnait avoir pris connaissance des presentes Conditions et les accepter sans reserve.",
    'La Societe se reserve le droit de modifier a tout moment et sans preavis tout ou partie des presentes.',
    "Les donnees collectees peuvent etre transmises a des partenaires commerciaux a des fins d'analyse.",
    "En aucun cas la responsabilite de la Societe ne saurait etre engagee pour tout dommage indirect.",
    "L'acces au Service peut etre suspendu ou resilie sans notification prealable et sans indemnite.",
    "L'Utilisateur concede une licence mondiale, non exclusive et gratuite sur les contenus publies.",
    'Les presentes sont regies par le droit applicable au siege social de la Societe, sauf exception.',
    "Toute utilisation du Service vaut acceptation pleine et entiere de l'ensemble des clauses ci-dessus.",
  ]
  const wallScroll = document.createElement('div')
  wallScroll.className = 'wall-scroll'
  let html = ''
  for (let i = 0; i < 90; i++) {
    const n = `${1 + Math.floor(i / 4)}.${1 + (i % 4)}`
    const body = [0, 1, 2].map((k) => clauses[(i * 3 + k * 5) % clauses.length]).join(' ')
    html += `<p><b>Article ${n}</b>${body}</p>`
  }
  wallScroll.innerHTML = html
  $('#wall').appendChild(wallScroll)

  // Les 3 mots de fin doivent tenir dans le cadre, meme si on change le texte
  const wordsRow = $('#words')
  while (wordsRow.offsetWidth > 1500) {
    wordsRow.style.fontSize = `${parseFloat(getComputedStyle(wordsRow).fontSize) - 2}px`
  }

  // ---------- Mesures (avant toute transformation) ----------
  const rect = (el) => el.getBoundingClientRect()

  const title = $('.s1-title')
  const em = $('#s1-l2 em')
  const scribble = $('.scribble')
  const emR = rect(em)
  const titleR = rect(title)
  Object.assign(scribble.style, {
    left: `${emR.left - titleR.left - 14}px`,
    top: `${emR.bottom - titleR.top - 34}px`,
    width: `${emR.width + 28}px`,
  })

  const brand = $('#brand')
  const wordmarkMask = $('#brand .wordmark-mask')
  const wordmarkW = wordmarkMask.offsetWidth
  const brandW = brand.offsetWidth
  const brandH = brand.offsetHeight

  const stepBoxes = ['#step-1', '#step-2', '#step-3'].map((sel) => ({
    x: $(sel).offsetLeft - 4,
    width: $(sel).offsetWidth + 8,
  }))

  const words = $$('.word')
  const wordsW = wordsRow.offsetWidth
  // Decalage pour garder centres les mots deja affiches
  const wordShift = words.map((w) => (wordsW - (w.offsetLeft + w.offsetWidth)) / 2)

  const urlW = $('#url').offsetWidth
  const inputR = rect($('.input'))
  const btnR = rect($('#btn'))
  const inputPoint = { x: inputR.left + 640, y: inputR.top + inputR.height / 2 }
  const btnPoint = { x: btnR.left + btnR.width * 0.56, y: btnR.top + btnR.height * 0.58 }

  const outroMask = $('#outro .wordmark-mask')
  const outroWordmarkW = outroMask.offsetWidth
  outroMask.style.width = '0px'
  const outroLogoR = rect($('#outro-logo'))
  const outroLogoCenter = {
    x: outroLogoR.left + outroLogoR.width / 2,
    y: outroLogoR.top + outroLogoR.height / 2,
  }

  // ---------- Etats initiaux ----------
  const arc = $('#grade-arc')
  const ARC_LEN = 2 * Math.PI * 104
  const scribblePath = $('#scribble')
  const SCRIBBLE_LEN = scribblePath.getTotalLength()
  const rings = $$('.ring')
  const taglineWords = $$('#tagline .w')
  const points = $$('.point')

  gsap.set('.wall', { opacity: 0 })
  gsap.set(['#s1-l1', '#s1-l2'], { yPercent: 125, rotation: 5 })
  gsap.set(scribblePath, { strokeDasharray: SCRIBBLE_LEN, strokeDashoffset: SCRIBBLE_LEN, opacity: 0 })
  gsap.set('#chip-time', { scale: 0.4, opacity: 0 })
  gsap.set('#clock-hand', { svgOrigin: '12 12', rotation: 0 })
  gsap.set('#dot', { scale: 0 })
  gsap.set(rings, { attr: { r: 0, 'stroke-width': 0 }, opacity: 1 })

  gsap.set('#logo', { scale: 0, rotation: -25 })
  gsap.set(wordmarkMask, { width: 0 })
  gsap.set('#wordmark', { xPercent: -30, opacity: 0 })
  gsap.set(taglineWords, { y: 50, opacity: 0, filter: 'blur(12px)' })

  gsap.set('#steps', { y: 30, opacity: 0 })
  gsap.set('#step-indicator', { ...stepBoxes[0], scale: 0 })
  gsap.set('#search', { y: 80, scale: 0.94, opacity: 0 })
  gsap.set(['#url', '#caret', '#spinner', '#btn-check'], { opacity: 0 })
  gsap.set('#btn-check', { scale: 0.4 })
  gsap.set('#progress-bar', { scaleX: 0 })
  gsap.set('#ripple', { scale: 0, opacity: 0 })
  gsap.set('.status-item', { yPercent: 100 })
  gsap.set('#keys', { y: 24, opacity: 0, scale: 0.92 })
  gsap.set('#result', { y: 90, scale: 0.96, opacity: 0 })
  gsap.set(['.grade-col', '.service', '#summary'], { opacity: 0, y: 24 })
  gsap.set(points, { opacity: 0, x: 60 })
  gsap.set('#pill-time', { scale: 0.5, opacity: 0 })
  gsap.set(arc, { strokeDasharray: ARC_LEN, strokeDashoffset: ARC_LEN })
  gsap.set('#grade-letter', { scale: 0.3, opacity: 0 })
  gsap.set('#cursor', { x: 1680, y: 1180, opacity: 0 })
  gsap.set('#click-ring', { scale: 0, opacity: 0 })

  gsap.set(words, { opacity: 0, scale: 1.7, filter: 'blur(22px)' })
  gsap.set('#outro-logo', { scale: 0, rotation: -25 })
  gsap.set('#outro-wordmark', { xPercent: -30, opacity: 0 })
  gsap.set('#outro-line', { y: 36, opacity: 0, filter: 'blur(10px)' })
  gsap.set('#cta', { scale: 0.7, opacity: 0, y: 20 })

  // ---------- Timeline ----------
  const tl = gsap.timeline({ paused: true, defaults: { ease: 'expo.out' } })

  // Fond qui respire en continu
  tl.to('.b-blue', { x: 260, y: 120, scale: 1.15, duration: DURATION, ease: 'sine.inOut' }, 0)
  tl.to('.b-yellow', { x: -220, y: 160, scale: 0.9, duration: DURATION, ease: 'sine.inOut' }, 0)
  tl.to('.b-green', { x: -300, y: -140, scale: 1.2, duration: DURATION, ease: 'sine.inOut' }, 0)
  tl.to('.b-red', { x: 380, y: -60, duration: DURATION, ease: 'sine.inOut' }, 0)

  // Camera
  tl.to('#camera', { scale: 1.07, rotation: -0.6, duration: 2.45, ease: 'power1.in' }, 0)
  tl.set('#camera', { scale: 1.16, rotation: 0 }, 2.45)
  tl.to('#camera', { scale: 1, duration: 1.8 }, 2.45)
  tl.to('#camera', { scale: 1.03, duration: 6.4, ease: 'sine.inOut' }, 4.4)
  tl.to('#camera', { scale: 1, duration: 0.4, ease: 'power3.in' }, 10.85)
  tl.to('#camera', { scale: 1.05, duration: 1.4, ease: 'sine.out' }, 11.25)
  tl.set('#camera', { scale: 1.1 }, 12.9)
  tl.to('#camera', { scale: 1, duration: 1.6 }, 12.9)
  tl.to('#camera', { scale: 1.025, duration: 1.5, ease: 'sine.inOut' }, 14.5)

  // ===== SCENE 1 : personne ne lit les CGU (0 -> 2.5) =====
  tl.to('.wall', { opacity: 1, duration: 0.7, ease: 'power2.out' }, 0)
  tl.fromTo(wallScroll, { y: 0 }, { y: -1500, duration: 2.5, ease: 'power1.in' }, 0)
  tl.to('#s1-l1', { yPercent: 0, rotation: 0, duration: 0.95 }, 0.08)
  tl.to('#s1-l2', { yPercent: 0, rotation: 0, duration: 0.95 }, 0.28)
  tl.set(scribblePath, { opacity: 1 }, 0.72)
  tl.to(scribblePath, { strokeDashoffset: 0, duration: 0.5, ease: 'power2.inOut' }, 0.72)
  tl.to('#chip-time', { scale: 1, opacity: 1, duration: 0.6, ease: 'back.out(2.2)' }, 0.92)
  tl.to('#clock-hand', { rotation: 1080, duration: 1.4, ease: 'power2.in' }, 0.92)

  tl.to(['#s1-l1', '#s1-l2'], { yPercent: -125, duration: 0.45, ease: 'expo.in', stagger: 0.05 }, 1.92)
  tl.to('.scribble', { opacity: 0, duration: 0.2, ease: 'none' }, 2.05)
  tl.to('#chip-time', { scale: 0, opacity: 0, duration: 0.38, ease: 'back.in(2)' }, 1.95)
  tl.to('.wall', { scale: 0.04, opacity: 0, duration: 0.55, ease: 'expo.in' }, 1.95)
  tl.to('.wall-fade', { opacity: 0, duration: 0.4, ease: 'power2.in' }, 2.0)
  tl.to('#dot', { scale: 1, duration: 0.18, ease: 'power4.out' }, 2.3)
  tl.to('#dot', { scale: 0, duration: 0.2, ease: 'power3.in' }, 2.45)

  // ===== SCENE 2 : logo (2.5 -> 4.3) =====
  const burst = (at, radius) => {
    rings.forEach((ring, i) => {
      tl.fromTo(
        ring,
        { attr: { r: 0, 'stroke-width': 46 }, opacity: 1 },
        { attr: { r: radius + i * 46, 'stroke-width': 0 }, duration: 1.15, ease: 'expo.out', immediateRender: false },
        at + i * 0.055,
      )
    })
  }
  burst(2.5, 290)
  tl.to('#logo', { scale: 1, rotation: 0, duration: 1.0, ease: 'back.out(1.6)' }, 2.52)
  tl.to(wordmarkMask, { width: wordmarkW, duration: 0.95, ease: 'expo.inOut' }, 2.95)
  tl.to('#wordmark', { xPercent: 0, opacity: 1, duration: 0.9, ease: 'expo.out' }, 3.25)
  tl.to(taglineWords, { y: 0, opacity: 1, filter: 'blur(0px)', duration: 0.75, stagger: 0.045 }, 3.35)
  tl.to(taglineWords, { y: -36, opacity: 0, filter: 'blur(10px)', duration: 0.4, ease: 'power3.in', stagger: 0.025 }, 4.2)

  // Le logo file en haut a gauche, comme la navbar du site
  const navScale = 0.3
  tl.to(
    brand,
    {
      x: 74 + (brandW * navScale) / 2 - 960,
      y: 98 - (300 + brandH / 2),
      scale: navScale,
      duration: 0.95,
      ease: 'expo.inOut',
    },
    4.3,
  )

  // ===== SCENE 3 : la demo (4.5 -> 11) =====
  // Indicateur noir qui glisse d'une etape a l'autre (comme un segmented control)
  const activate = (i, at) => {
    if (i === 0) {
      tl.to('#step-indicator', { scale: 1, duration: 0.6 }, at)
    } else {
      tl.to('#step-indicator', { ...stepBoxes[i], duration: 0.65, ease: 'expo.inOut' }, at)
    }
    // Le texte passe en blanc quand l'indicateur arrive dessous
    const sel = `#step-${i + 1}`
    const textAt = at + (i === 0 ? 0.1 : 0.34)
    tl.to(sel, { color: '#ffffff', duration: 0.15, ease: 'none' }, textAt)
    tl.to(`${sel} b`, { backgroundColor: '#ffffff', color: '#11131a', duration: 0.15, ease: 'none' }, textAt)
  }
  const done = (i, at) => {
    // ... et repasse en noir quand il s'en va
    const sel = `#step-${i + 1}`
    tl.to(`${sel} b`, { backgroundColor: '#34a853', color: '#ffffff', duration: 0.3, ease: 'none' }, at)
    tl.to(sel, { color: '#11131a', duration: 0.12, ease: 'none' }, at + 0.27)
  }
  const click = (at) => {
    tl.to('#cursor', { scale: 0.8, duration: 0.08, ease: 'power2.out' }, at)
    tl.to('#cursor', { scale: 1, duration: 0.35, ease: 'back.out(3)' }, at + 0.08)
    tl.fromTo('#click-ring', { scale: 0.2, opacity: 1 }, { scale: 2.2, opacity: 0, duration: 0.55, ease: 'expo.out', immediateRender: false }, at)
  }
  const cursorTo = (p, at, duration) =>
    tl.to('#cursor', { x: p.x - 11, y: p.y - 6, duration, ease: 'power3.inOut' }, at)

  tl.to('#steps', { y: 0, opacity: 1, duration: 0.85 }, 4.98)
  tl.to('#search', { y: 0, scale: 1, opacity: 1, duration: 1.05 }, 4.7)
  activate(0, 5.2)

  // 1. Colle un lien
  tl.set('#cursor', { opacity: 1 }, 4.85)
  cursorTo(inputPoint, 4.85, 0.72)
  click(5.57)
  tl.to('#keys', { y: 0, opacity: 1, scale: 1, duration: 0.5, ease: 'back.out(1.8)' }, 5.78)
  tl.to('#key-ctrl', { y: 6, boxShadow: '0 1px 0 rgba(17,19,26,0.12), 0 8px 20px rgba(47,63,89,0.1)', backgroundColor: '#eef3fe', duration: 0.1, ease: 'power2.out' }, 6.03)
  tl.to('#key-v', { y: 6, boxShadow: '0 1px 0 rgba(17,19,26,0.12), 0 8px 20px rgba(47,63,89,0.1)', backgroundColor: '#eef3fe', duration: 0.1, ease: 'power2.out' }, 6.11)
  tl.to(['#key-ctrl', '#key-v'], { y: 0, boxShadow: '0 7px 0 rgba(17,19,26,0.12), 0 18px 40px rgba(47,63,89,0.12)', backgroundColor: '#ffffff', duration: 0.3, ease: 'power2.out' }, 6.32)
  tl.to('#placeholder', { opacity: 0, x: 16, duration: 0.08, ease: 'power2.in' }, 6.14)
  tl.fromTo('#url', { x: -24, opacity: 0 }, { x: 0, opacity: 1, duration: 0.5, immediateRender: false }, 6.2)
  tl.to('#url-sel', { opacity: 0, duration: 0.5, ease: 'power2.inOut' }, 6.55)
  tl.set('#caret', { x: urlW + 8 }, 6.2)
  tl.to('#keys', { y: 16, opacity: 0, duration: 0.35, ease: 'power2.in' }, 6.6)

  // 2. Clique
  done(0, 6.65)
  activate(1, 6.65)
  cursorTo(btnPoint, 6.62, 0.62)
  click(7.22)
  tl.to('#cursor', { x: '+=70', y: '+=64', duration: 0.6, ease: 'power3.out' }, 7.42)
  tl.fromTo('#btn', { scale: 0.93 }, { scale: 1, duration: 0.45, ease: 'back.out(2.5)', immediateRender: false }, 7.24)
  tl.fromTo('#ripple', { scale: 0, opacity: 1 }, { scale: 4, opacity: 0, duration: 0.7, immediateRender: false }, 7.22)
  tl.to('#btn-label', { yPercent: -110, opacity: 0, duration: 0.3, ease: 'power3.in' }, 7.26)
  tl.fromTo('#spinner', { opacity: 0, scale: 0.5 }, { opacity: 1, scale: 1, duration: 0.35, ease: 'back.out(2)', immediateRender: false }, 7.42)
  tl.to('#spinner', { rotation: 1100, duration: 1.15, ease: 'none' }, 7.3)
  tl.to('#progress-bar', { scaleX: 1, duration: 1.1, ease: 'power2.inOut' }, 7.28)
  // Ticker de statut : chaque message pousse le precedent vers le haut
  ;['#status-1', '#status-2', '#status-3'].forEach((sel, i) => {
    const at = 7.3 + i * 0.4
    tl.to(sel, { yPercent: 0, duration: 0.45, ease: 'expo.inOut' }, at)
    if (i < 2) tl.to(sel, { yPercent: -100, duration: 0.45, ease: 'expo.inOut' }, at + 0.4)
  })
  tl.to('#spinner', { opacity: 0, scale: 0.5, duration: 0.2, ease: 'power2.in' }, 8.36)
  tl.to('#btn-check', { opacity: 1, scale: 1, duration: 0.5, ease: 'back.out(2.6)' }, 8.42)
  tl.to('#cursor', { opacity: 0, y: '+=40', duration: 0.35, ease: 'power2.in' }, 8.42)

  // 3. C'est compris
  done(1, 8.5)
  activate(2, 8.5)
  tl.to('.progress', { opacity: 0, duration: 0.3, ease: 'none' }, 8.5)
  tl.to('#status-3', { yPercent: -100, duration: 0.45, ease: 'expo.inOut' }, 8.5)
  tl.to('#search', { y: -110, duration: 0.9, ease: 'expo.inOut' }, 8.55)
  tl.to('#result', { y: 0, scale: 1, opacity: 1, duration: 1.05 }, 8.75)
  tl.to('.grade-col', { y: 0, opacity: 1, duration: 0.8 }, 8.88)
  tl.to(arc, { strokeDashoffset: ARC_LEN * 0.2, duration: 1.3, ease: 'power3.out' }, 8.95)
  tl.to('#grade-letter', { scale: 1, opacity: 1, duration: 0.8, ease: 'back.out(2.2)' }, 9.05)
  tl.to(['.service', '#summary'], { y: 0, opacity: 1, duration: 0.8, stagger: 0.1 }, 8.95)
  tl.to(points, { x: 0, opacity: 1, duration: 0.8, stagger: 0.12 }, 9.2)
  tl.to('#pill-time', { scale: 1, opacity: 1, duration: 0.6, ease: 'back.out(2.4)' }, 9.75)

  // Sortie de la demo
  tl.to(['#result', '#search'], { y: '-=90', opacity: 0, filter: 'blur(8px)', duration: 0.45, ease: 'expo.in', stagger: 0.05 }, 10.82)
  tl.to('#steps', { y: -30, opacity: 0, duration: 0.4, ease: 'expo.in' }, 10.8)
  tl.to(brand, { opacity: 0, duration: 0.35, ease: 'power2.in' }, 10.85)

  // ===== SCENE 4 : Colle. Clique. Compris. (11.2 -> 12.9) =====
  gsap.set(wordsRow, { x: wordShift[0] })
  words.forEach((word, i) => {
    const at = 11.22 + i * 0.3
    tl.to(word, { opacity: 1, scale: 1, filter: 'blur(0px)', duration: 0.6 }, at)
    if (i > 0) tl.to(wordsRow, { x: wordShift[i], duration: 0.55 }, at)
  })
  tl.to(words, { y: -70, opacity: 0, filter: 'blur(14px)', duration: 0.45, ease: 'expo.in', stagger: 0.05 }, 12.55)

  // ===== OUTRO (12.9 -> 15) =====
  tl.set('#burst', { x: outroLogoCenter.x - 960, y: outroLogoCenter.y - 540 }, 12.85)
  burst(12.9, 230)
  tl.to('#outro-logo', { scale: 1, rotation: 0, duration: 1.0, ease: 'back.out(1.6)' }, 12.92)
  tl.to(outroMask, { width: outroWordmarkW, duration: 0.9, ease: 'expo.inOut' }, 13.25)
  tl.to('#outro-wordmark', { xPercent: 0, opacity: 1, duration: 0.85 }, 13.5)
  tl.to('#outro-line', { y: 0, opacity: 1, filter: 'blur(0px)', duration: 0.85 }, 13.62)
  tl.to('#cta', { scale: 1, opacity: 1, y: 0, duration: 0.75, ease: 'back.out(1.8)' }, 13.8)
  tl.to('#cta-arrow', { x: 10, duration: 0.3, ease: 'power2.out' }, 14.4)
  tl.to('#cta-arrow', { x: 0, duration: 0.5, ease: 'power2.inOut' }, 14.7)

  tl.set({}, {}, DURATION)

  // ---------- Elements proceduraux (dependent seulement du temps) ----------
  const caret = $('#caret')
  const procedural = (t) => {
    const on = t >= 5.6 && t < 7.26
    caret.style.opacity = on && Math.floor((t - 5.6) / 0.4) % 2 === 0 ? '1' : '0'
  }

  window.__duration = DURATION
  window.__seek = (t) => {
    tl.time(Math.min(t, DURATION), false)
    procedural(t)
  }

  if (RENDER) {
    window.__seek(Number(params.get('t') || 0))
    window.__ready = true
  } else {
    tl.eventCallback('onUpdate', () => procedural(tl.time()))
    tl.eventCallback('onComplete', () => gsap.delayedCall(1.2, () => tl.restart()))
    tl.play(Number(params.get('t') || 0))
  }
})()

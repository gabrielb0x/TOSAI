// Rendu de la promo TOSAI.
//   node render.mjs                 -> tosai-promo.mp4 (1920x1080, 60 fps)
//   node render.mjs --preview       -> sert la scene sur http://localhost:4173 (lecture temps reel)
//   node render.mjs --still 1,5.5,9 -> captures PNG dans frames/ pour verifier un instant precis
// Options : --fps 60, --blur 2 (sous-images moyennees = flou de mouvement), --out fichier.mp4
// Chromium : celui de Playwright, ou CHROMIUM_PATH=/chemin/vers/chrome.
import { createServer } from 'node:http'
import { readFile, mkdir, writeFile } from 'node:fs/promises'
import { extname, join, normalize, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import { spawn } from 'node:child_process'
import { chromium } from 'playwright'

const ROOT = dirname(fileURLToPath(import.meta.url))
const args = process.argv.slice(2)
const opt = (name, fallback) => {
  const i = args.indexOf(`--${name}`)
  return i === -1 ? fallback : args[i + 1]
}

const MIME = {
  '.html': 'text/html',
  '.css': 'text/css',
  '.js': 'text/javascript',
  '.png': 'image/png',
  '.woff2': 'font/woff2',
}

const server = createServer(async (req, res) => {
  const path = normalize(decodeURIComponent(new URL(req.url, 'http://x').pathname)).replace(/^(\.\.[/\\])+/, '')
  try {
    const body = await readFile(join(ROOT, path === '/' ? 'scene.html' : path))
    res.writeHead(200, { 'Content-Type': MIME[extname(path)] || 'application/octet-stream' })
    res.end(body)
  } catch {
    res.writeHead(404).end()
  }
})

const port = Number(opt('port', 4173))
await new Promise((resolve) => server.listen(port, resolve))
const base = `http://localhost:${port}/scene.html`

if (args.includes('--preview')) {
  console.log(`Preview : ${base}  (ajoute ?t=8 pour demarrer a 8 s)`)
} else {
  const browser = await chromium.launch({
    executablePath: process.env.CHROMIUM_PATH || undefined,
    args: ['--force-color-profile=srgb', '--hide-scrollbars'],
  })
  const page = await browser.newPage({ viewport: { width: 1920, height: 1080 }, deviceScaleFactor: 1 })
  await page.goto(`${base}?render`)
  await page.waitForFunction(() => window.__ready === true)
  const duration = await page.evaluate(() => window.__duration)
  const cdp = await page.context().newCDPSession(page)

  const shoot = async (t) => {
    await page.evaluate((time) => {
      window.__seek(time)
      return new Promise((r) => requestAnimationFrame(() => requestAnimationFrame(r)))
    }, t)
    const { data } = await cdp.send('Page.captureScreenshot', { format: 'png' })
    return Buffer.from(data, 'base64')
  }

  const still = opt('still', null)
  if (still) {
    await mkdir(join(ROOT, 'frames'), { recursive: true })
    for (const t of still.split(',').map(Number)) {
      const file = join(ROOT, 'frames', `t-${t.toFixed(2)}.png`)
      await writeFile(file, await shoot(t))
      console.log(file)
    }
  } else {
    const fps = Number(opt('fps', 60))
    const blur = Math.max(1, Number(opt('blur', 2)))
    const out = join(ROOT, opt('out', 'tosai-promo.mp4'))
    const total = Math.round(duration * fps * blur)

    const ffmpeg = spawn(
      'ffmpeg',
      [
        '-y', '-loglevel', 'error',
        '-f', 'image2pipe', '-framerate', String(fps * blur), '-c:v', 'png', '-i', '-',
        '-vf', `${blur > 1 ? `tmix=frames=${blur},` : ''}fps=${fps},format=yuv420p`,
        '-c:v', 'libx264', '-preset', 'slow', '-crf', '14', '-profile:v', 'high',
        '-color_primaries', 'bt709', '-color_trc', 'bt709', '-colorspace', 'bt709',
        '-movflags', '+faststart', out,
      ],
      { stdio: ['pipe', 'inherit', 'inherit'] },
    )
    const done = new Promise((resolve, reject) =>
      ffmpeg.on('close', (code) => (code === 0 ? resolve() : reject(new Error(`ffmpeg exit ${code}`)))),
    )

    const started = Date.now()
    for (let i = 0; i < total; i++) {
      const png = await shoot(i / (fps * blur))
      if (!ffmpeg.stdin.write(png)) await new Promise((r) => ffmpeg.stdin.once('drain', r))
      if (i % 60 === 0 || i === total - 1) {
        const pct = (((i + 1) / total) * 100).toFixed(1)
        process.stdout.write(`\r${i + 1}/${total} images (${pct} %) - ${((Date.now() - started) / 1000).toFixed(0)} s`)
      }
    }
    ffmpeg.stdin.end()
    await done
    console.log(`\nOK -> ${out}`)
  }

  await browser.close()
  server.close()
}

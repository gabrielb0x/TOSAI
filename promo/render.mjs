import { createServer } from 'node:http'
import { readFile, mkdir, writeFile, rm } from 'node:fs/promises'
import { cpus } from 'node:os'
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

const openScene = async () => {
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
    const { data } = await cdp.send('Page.captureScreenshot', { format: 'png', optimizeForSpeed: true })
    return Buffer.from(data, 'base64')
  }

  return { browser, duration, shoot }
}

if (args.includes('--preview')) {
  console.log(`Preview : ${base}  (ajoute ?t=8 pour demarrer a 8 s)`)
} else if (opt('still', null)) {
  const { browser, shoot } = await openScene()
  await mkdir(join(ROOT, 'frames'), { recursive: true })
  for (const t of opt('still').split(',').map(Number)) {
    const file = join(ROOT, 'frames', `t-${t.toFixed(2)}.png`)
    await writeFile(file, await shoot(t))
    console.log(file)
  }
  await browser.close()
  server.close()
} else {
  const fps = Number(opt('fps', 60))
  const blur = Math.max(1, Number(opt('blur', 2)))
  const workers = Math.max(1, Number(opt('workers', Math.max(1, cpus().length - 1))))
  const out = join(ROOT, opt('out', 'tosai-promo.mp4'))
  const dir = join(ROOT, 'frames', 'video')
  await rm(dir, { recursive: true, force: true })
  await mkdir(dir, { recursive: true })

  const scenes = await Promise.all(Array.from({ length: workers }, openScene))
  const total = Math.round(scenes[0].duration * fps * blur)
  const started = Date.now()
  let next = 0
  let doneCount = 0
  await Promise.all(
    scenes.map(async ({ shoot }) => {
      while (next < total) {
        const i = next++
        await writeFile(join(dir, `${String(i).padStart(5, '0')}.png`), await shoot(i / (fps * blur)))
        doneCount++
        if (doneCount % 30 === 0 || doneCount === total) {
          const secs = ((Date.now() - started) / 1000).toFixed(0)
          process.stdout.write(`\r${doneCount}/${total} images (${workers} workers) - ${secs} s`)
        }
      }
    }),
  )
  await Promise.all(scenes.map(({ browser }) => browser.close()))
  server.close()

  console.log('\nEncodage...')
  await new Promise((resolve, reject) => {
    const ffmpeg = spawn(
      'ffmpeg',
      [
        '-y', '-loglevel', 'error',
        '-framerate', String(fps * blur), '-i', join(dir, '%05d.png'),
        '-vf', `${blur > 1 ? `tmix=frames=${blur},` : ''}fps=${fps},noise=alls=2:allf=t,format=yuv420p`,
        '-c:v', 'libx264', '-preset', 'slow', '-crf', '15', '-profile:v', 'high', '-tune', 'animation',
        '-color_primaries', 'bt709', '-color_trc', 'bt709', '-colorspace', 'bt709',
        '-movflags', '+faststart', out,
      ],
      { stdio: 'inherit' },
    )
    ffmpeg.on('close', (code) => (code === 0 ? resolve() : reject(new Error(`ffmpeg exit ${code}`))))
  })
  if (!args.includes('--keep-frames')) await rm(dir, { recursive: true, force: true })
  console.log(`OK -> ${out}`)
}

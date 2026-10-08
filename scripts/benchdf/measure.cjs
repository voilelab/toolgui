// Measures a Toggle rerun in the browser, web and wasm alike (TG-94):
// click -> first pack, -> DataFrame pack, -> result pack, -> next paint, plus
// the main thread time spent in JSON.parse of the big packs.
//
//	go run ./scripts/benchdf -addr 127.0.0.1:3100
//	go run ./cmd/toolgui-wasm serve -addr 127.0.0.1:3200 ./scripts/benchdf
//	npm i -g playwright && npx playwright install chromium
//	NODE_PATH=$(npm root -g) node scripts/benchdf/measure.cjs
const { chromium } = require('playwright')

const WEB = process.env.WEB_URL || 'http://127.0.0.1:3100/'
const WASM = process.env.WASM_URL || 'http://127.0.0.1:3200/#/'
const ROWS = [0, 1000, 3000, 10000]
const REPS = Number(process.env.REPS || 30)

// Runs before the app: taps the transport ahead of the app's own listener.
function probe() {
  const log = window.__bench = { msgs: [], parse: 0 }

  // A pack's kind, read without parsing it so the probe costs next to nothing.
  const kindOfText = s =>
    s.includes('"name":"dataframe_component"') ? 'df'
      : (s.length < 4096 && s.includes('"success"')) ? 'result' : 'other'
  const kindOfPack = p =>
    p?.component?.name === 'dataframe_component' ? 'df'
      : (p && 'success' in p) ? 'result' : 'other'

  const NativeWS = window.WebSocket
  window.WebSocket = class extends NativeWS {
    constructor(...args) {
      super(...args)
      this.addEventListener('message', e => {
        log.msgs.push({ t: performance.now(), kind: kindOfText(e.data), size: e.data.length })
      })
    }
  }

  const NativeWorker = window.Worker
  window.Worker = class extends NativeWorker {
    constructor(...args) {
      super(...args)
      this.addEventListener('message', e => {
        if (e.data?.kind === 'pack') {
          log.msgs.push({ t: performance.now(), kind: kindOfPack(e.data.pack), size: 0 })
        }
      })
    }
  }

  // The web transport parses each pack on the main thread; time the big ones.
  const parse = JSON.parse
  JSON.parse = function (text, ...rest) {
    const t = performance.now()
    const v = parse.call(this, text, ...rest)
    if (typeof text === 'string' && text.length > 4096) {
      log.parse += performance.now() - t
    }
    return v
  }
}

// Resolves once the result pack is in and the frame after it is painted.
function waitDone(since) {
  return new Promise(resolve => {
    const check = () => {
      const r = window.__bench.msgs.find(m => m.t >= since && m.kind === 'result')
      if (!r) {
        setTimeout(check, 0)
        return
      }
      setTimeout(() => requestAnimationFrame(() => setTimeout(() =>
        resolve(performance.now()), 0)), 0)
    }
    check()
  })
}

async function measure(page, url) {
  await page.goto(url)
  await page.waitForSelector('input[id=toggle_component_Flag]', { timeout: 120000 })
  await page.waitForTimeout(500)

  const samples = []
  for (let i = 0; i < REPS + 1; i++) {
    const s = await page.evaluate(async () => {
      const log = window.__bench
      log.parse = 0
      const t0 = performance.now()
      document.querySelector('input[id=toggle_component_Flag]').click()
      const done = await waitDoneFn(t0)
      const msgs = log.msgs.filter(m => m.t >= t0)
      const at = kind => (msgs.find(m => m.kind === kind)?.t ?? NaN) - t0
      return {
        first: msgs.length ? msgs[0].t - t0 : NaN,
        df: at('df'),
        result: at('result'),
        paint: done - t0,
        parse: log.parse,
        dfSize: msgs.find(m => m.kind === 'df')?.size ?? 0,
      }
    })
    if (i > 0) samples.push(s) // the first click warms up
    await page.waitForTimeout(200)
  }
  return samples
}

const median = xs => {
  const s = xs.filter(x => !Number.isNaN(x)).sort((a, b) => a - b)
  return s.length ? s[Math.floor(s.length / 2)] : NaN
}

;(async () => {
  const browser = await chromium.launch()
  for (const [mode, base] of [['web', WEB], ['wasm', WASM]]) {
    console.log(`\n${mode} (median of ${REPS} toggles, ms)`)
    console.log('rows    first     df  result   paint   parse')
    for (const n of ROWS) {
      const page = await browser.newPage()
      await page.addInitScript(probe)
      await page.addInitScript(`window.waitDoneFn = ${waitDone.toString()}`)
      const samples = await measure(page, base + 'rows' + n)
      const m = k => median(samples.map(s => s[k])).toFixed(1).padStart(7)
      console.log(`${String(n).padEnd(6)}${m('first')}${m('df')}${m('result')}${m('paint')}${m('parse')}`)
      await page.close()
    }
  }
  await browser.close()
})()

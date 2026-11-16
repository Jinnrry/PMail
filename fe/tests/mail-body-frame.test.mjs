/* eslint-env node, browser */
import assert from 'node:assert/strict'
import {existsSync} from 'node:fs'
import http from 'node:http'
import {fileURLToPath} from 'node:url'
import {after, before, test} from 'node:test'
import {setTimeout as delay} from 'node:timers/promises'
import {createServer} from 'vite'
import vue from '@vitejs/plugin-vue'
import puppeteer from 'puppeteer-core'

let server, browser, baseURL
before(async () => {
  const executablePath = process.env.CHROME_BIN || [
    '/usr/bin/google-chrome', '/usr/bin/chromium', '/usr/bin/chromium-browser',
    '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
  ].find(existsSync)
  assert.ok(executablePath, 'Install Chrome/Chromium or set CHROME_BIN to its executable')
  server = await createServer({
    root: fileURLToPath(new URL('../', import.meta.url)),
    configFile: false,
    plugins: [vue()],
    optimizeDeps: {entries: ['tests/fixtures/mail-body.html']},
    server: {host: '127.0.0.1', port: 0},
    logLevel: 'error',
  })
  await server.listen()
  baseURL = `http://127.0.0.1:${server.httpServer.address().port}`
  browser = await puppeteer.launch({executablePath, headless: true,
    args: process.getuid?.() === 0 ? ['--no-sandbox'] : []})
})
after(async () => { await browser?.close(); await server?.close() })

async function newPage(t) {
  const page = await browser.newPage()
  t.after(() => page.close())
  await page.goto(`${baseURL}/tests/fixtures/mail-body.html`)
  await page.waitForSelector('iframe')
  return page
}

let generation = 0
async function setHTML(page, html) {
  const token = String(++generation)
  await page.evaluate(value => window.setMailHTML(value),
    `${html}<span hidden id="fixture-ready" data-generation="${token}"></span>`)
  await page.waitForFunction(value => {
    return document.querySelector('iframe')?.contentDocument
      ?.querySelector('#fixture-ready')?.dataset.generation === value
  }, {}, token)
}

test('viewport-dependent CSS stops resizing and oversized messages remain scrollable', async t => {
  const page = await newPage(t)
  await setHTML(page, '<div id="message-ready" style="min-height:100vh;padding:24px">Message</div>')
  const heights = []
  for (let i = 0; i < 10; i++) {
    await delay(150)
    heights.push(await page.$eval('iframe', el => el.getBoundingClientRect().height))
  }
  assert.ok(heights.every(height => height <= 1200), `Unbounded frame height: ${heights}`)
  assert.equal(new Set(heights.slice(-3)).size, 1, `Frame keeps resizing: ${heights}`)

  await setHTML(page, '<div id="message-ready" style="height:10000px">Long message</div>')
  await page.waitForFunction(() => document.querySelector('iframe').clientHeight === 1200)
  assert.ok(await page.$eval('iframe', el => el.contentDocument.documentElement.scrollHeight > el.clientHeight))

  await setHTML(page, '<p id="message-ready">Short message</p>')
  await page.waitForFunction(() => document.querySelector('iframe').clientHeight === 200)
})

test('email scripts, event handlers and javascript links remain sandboxed', async t => {
  const page = await newPage(t)
  await setHTML(page, `<div id="message-ready">
    <script>parent.scriptExecutions++</script>
    <img src="data:image/png;base64,AA==" onerror="parent.scriptExecutions++">
    <svg onload="parent.scriptExecutions++"></svg>
    <a id="script-link" target="_self" href="javascript:parent.scriptExecutions++">Link</a>
  </div>`)
  const frame = await (await page.$('iframe')).contentFrame()
  await frame.click('#script-link')
  await delay(150)
  assert.equal(await page.evaluate(() => window.scriptExecutions), 0)
  assert.ok(!await page.$eval('iframe', el => el.sandbox.contains('allow-scripts')))
})

test('cross-origin navigation does not throw in the frame load handler', async t => {
  const external = http.createServer((req, res) => res.end('External page'))
  await new Promise(resolve => external.listen(0, '127.0.0.1', resolve))
  t.after(() => new Promise(resolve => external.close(resolve)))
  const page = await newPage(t)
  const errors = []
  page.on('pageerror', err => errors.push(err.message))
  const url = `http://127.0.0.1:${external.address().port}/`
  await setHTML(page, `<a id="message-ready" target="_self" href="${url}">External</a>`)
  const frame = await (await page.$('iframe')).contentFrame()
  await Promise.all([frame.waitForNavigation(), frame.click('#message-ready')])
  assert.equal(frame.url(), url)
  assert.deepEqual(errors, [])
})

/**
 * Drive the mini-program in WeChat DevTools and capture screenshots for the
 * demo docs. Prereqs:
 *   - WeChat DevTools installed and logged in
 *   - DevTools -> Settings -> Security -> Service Port enabled
 *   - Local backend running on http://127.0.0.1:8081 (scripts/dev-backend.sh)
 * Usage: node scripts/mp-screenshots.js
 */
const path = require('path')
const fs = require('fs')
const { execFileSync } = require('child_process')
const automator = require('miniprogram-automator')

const ROOT = path.resolve(__dirname, '..')
const OUT_DIR = path.join(ROOT, 'docs', 'screenshots')
const CLI_PATH = 'C:/Program Files (x86)/Tencent/微信web开发者工具/cli.bat'
const CAPTURE_PS1 = path.join(__dirname, 'capture-devtools.ps1')

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
let shotIndex = 0
const failures = []
const SKIP_SHOTS = process.env.SKIP_SHOTS === '1'

/** Some DevTools builds never resolve navigateTo/switchTab promises although
 *  the navigation happens; race them so the flow keeps moving. */
async function navigate(miniProgram, kind, url) {
  try {
    await Promise.race([
      miniProgram[kind](url),
      new Promise((_, reject) => setTimeout(() => reject(new Error('nav ack timeout')), 5000)),
    ])
  } catch (e) {
    console.log('  (ignored ' + kind + ' ack: ' + e.message + ')')
  }
  await sleep(800)
}

async function shot(miniProgram, name) {
  if (SKIP_SHOTS) {
    console.log('  [skip screenshot]', name)
    return
  }
  shotIndex += 1
  const file = path.join(OUT_DIR, String(shotIndex).padStart(2, '0') + '-' + name + '.png')
  // Prefer the automation protocol; on some DevTools builds it never returns
  // (locked/occluded rendering), so fall back to a window capture.
  try {
    await Promise.race([
      miniProgram.screenshot({ path: file }),
      new Promise((_, reject) => setTimeout(() => reject(new Error('screenshot timeout')), 25000)),
    ])
    console.log('  screenshot:', path.relative(ROOT, file))
  } catch (e) {
    console.log('  native screenshot failed (' + e.message + '), capturing window...')
    execFileSync(
      'powershell',
      [
        '-NoProfile',
        '-ExecutionPolicy', 'Bypass',
        '-File', CAPTURE_PS1,
        '-OutFile', file,
      ],
      { stdio: 'inherit' }
    )
  }
  await sleep(400)
}

async function waitForPage(miniProgram, match, timeoutMs = 30000) {
  const started = Date.now()
  while (Date.now() - started < timeoutMs) {
    const page = await miniProgram.currentPage()
    if (page && page.path.indexOf(match) !== -1) return page
    await sleep(500)
  }
  throw new Error('timeout waiting for page ' + match)
}

async function run() {
  fs.mkdirSync(OUT_DIR, { recursive: true })
  // Reuse an IDE that was started with `cli auto --auto-port 9420`; only
  // launch a new one when nothing is listening yet.
  let miniProgram
  try {
    miniProgram = await Promise.race([
      automator.connect({ wsEndpoint: 'ws://127.0.0.1:9420' }),
      new Promise((_, reject) => setTimeout(() => reject(new Error('connect timeout')), 10000)),
    ])
    console.log('connected to a running DevTools on port 9420')
  } catch (connectErr) {
    console.log('no reachable DevTools automation port (' + connectErr.message + '), launching...')
    miniProgram = await automator.launch({
      cliPath: CLI_PATH,
      projectPath: ROOT,
      port: 9420,
    })
  }

  try {
    // 1. Home
    await miniProgram.reLaunch('/pages/index/index')
    await sleep(4000)
    await shot(miniProgram, 'home')

    // 2. Wechat one-tap login
    await navigate(miniProgram, 'navigateTo', '/pages/user/login/index')
    const loginPage = await waitForPage(miniProgram, 'user/login')
    await sleep(800)
    await shot(miniProgram, 'login')
    const wechatBtn = await loginPage.$('.wechat-btn')
    if (wechatBtn) {
      await Promise.race([
        wechatBtn.tap(),
        new Promise((_, rej) => setTimeout(() => rej(new Error('tap timeout')), 5000)),
      ]).catch((e) => {
        console.log('  (tap fallback: ' + e.message + ')')
        return loginPage.callMethod('onWechatLogin')
      })
    } else {
      await loginPage.callMethod('onWechatLogin')
    }
    await sleep(3000)
    console.log('  logged in (dev-mode wechat login)')

    // 3. Route generation
    await navigate(miniProgram, 'switchTab', '/pages/user/routeInput/index')
    const routePage = await waitForPage(miniProgram, 'routeInput')
    await sleep(800)
    await routePage.setData({
      interestTags: [
        { label: 'Foodie', value: 'foodie', icon: 'fork', selected: true },
        { label: 'History', value: 'history', icon: 'temple', selected: false },
        { label: 'Nature', value: 'nature', icon: 'tree', selected: false },
        { label: 'Art', value: 'art', icon: 'palette', selected: false },
        { label: 'Shopping', value: 'shopping', icon: 'shop', selected: false },
      ],
    })
    await shot(miniProgram, 'route-form')
    try {
      await routePage.callMethod('onGenerate')
    } catch (e) {
      console.log('  onGenerate resolved late:', e.message)
    }
    console.log('  generating AI route (up to ~60s)...')
    const mapPage = await waitForPage(miniProgram, 'user/map', 120000)
    // Wait until the polyline is rendered.
    for (let i = 0; i < 60; i++) {
      const data = await mapPage.data()
      if (data.polyline && data.polyline.length > 0) break
      await sleep(500)
    }
    await sleep(800)
    await shot(miniProgram, 'route-map')

    // 4. AI guide
    await navigate(miniProgram, 'switchTab', '/pages/user/guide/index')
    const guidePage = await waitForPage(miniProgram, 'user/guide')
    await sleep(800)
    await guidePage.setData({ inputValue: '广州早茶有哪些必点菜？' })
    await guidePage.callMethod('onSend').catch((e) => console.log('  onSend ack:', e.message))
    console.log('  asking the AI guide...')
    await sleep(14000)
    await shot(miniProgram, 'guide')

    // 5. Travel blog: AI draft -> publish
    await navigate(miniProgram, 'switchTab', '/pages/user/travelBlog/index')
    const blogPage = await waitForPage(miniProgram, 'travelBlog')
    await sleep(800)
    await blogPage.setData({ topic: '周末夜游珠江看灯光秀' })
    await blogPage.callMethod('onGenerate').catch((e) => console.log('  onGenerate ack:', e.message))
    console.log('  drafting blog with AI...')
    await sleep(14000)
    await shot(miniProgram, 'blog-draft')
    await blogPage.callMethod('onPublish').catch((e) => console.log('  onPublish ack:', e.message))
    await sleep(8000)
    await shot(miniProgram, 'blog-published')

    // 6. Profile
    await navigate(miniProgram, 'switchTab', '/pages/user/profile/index')
    await waitForPage(miniProgram, 'user/profile')
    await sleep(5000)
    await shot(miniProgram, 'profile')

    // 7. Merchant copy + poster
    await navigate(miniProgram, 'navigateTo', '/pages/merchant/generate/index')
    const genPage = await waitForPage(miniProgram, 'merchant/generate')
    await sleep(800)
    await genPage.setData({
      inputText: '粤香茶楼招牌虾饺皇，皮薄馅足',
      languages: ['zh'],
    })
    await genPage.callMethod('onGenerate').catch((e) => console.log('  onGenerate ack:', e.message))
    console.log('  generating marketing copy...')
    await sleep(18000)
    await shot(miniProgram, 'merchant-copy')
    await genPage.callMethod('onGeneratePoster').catch((e) => console.log('  onGeneratePoster ack:', e.message))
    console.log('  generating poster (polling ~40s)...')
    await sleep(45000)
    await shot(miniProgram, 'merchant-poster')

    // Note: menuScan needs a real photo; capture it manually during the demo.
    console.log('\nDone. Screenshots in docs/screenshots/')
    console.log('menuScan is manual: choose a photo in DevTools, then Culture Card / Translate.')
  } finally {
    // Leave the IDE open so the user can keep recording; just drop the ws link.
    await miniProgram.disconnect().catch(() => {})
    if (failures.length) {
      console.log('Failed steps:', failures)
    }
  }
}

run().catch((err) => {
  console.error('\nScreenshot run failed:', err.message || err)
  if (String(err.message || err).indexOf('Service Port') !== -1) {
    console.error('Enable it in DevTools: 设置 -> 安全设置 -> 服务端口')
  }
  process.exitCode = 1
})

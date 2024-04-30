import * as cheerio from 'cheerio'
import { listDir, join } from './utils.js'

export class Connector {
  constructor (data) {
    Object.assign(this, data)
  }
}

export async function fetchNextPage (url) {
  const res = await fetch(url)
  const html = await res.text()
  const page = cheerio.load(html)
  const nextData = JSON.parse(page('#__NEXT_DATA__').text())
  const data = nextData.props.pageProps
  return data
}

export async function loadConnectors (api) {
  const cdir = './connectors'

  const out = {}
  for (const cf of listDir(cdir)) {
    const d = await import(join('../connectors', cf))
    const n = cf.replace('.js', '')
    out[n] = d.default
    out[n].id = n
  }
  return out
}

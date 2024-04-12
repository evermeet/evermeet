
import { initDatabase, loadMockData } from './db.js'
import { parse, stringify } from 'yaml'
import fs from 'node:fs'

import { makeRoutes } from './routes.js';
import { makeCollections } from './collections.js';

export class API {

  constructor () {
    this.models = {}
    this.config = parse(fs.readFileSync('../../config.yaml', 'utf-8'))
    this.pkg = JSON.parse(fs.readFileSync('../../package.json', 'utf-8'))

    console.log(stringify(this.config))
  }

  async init () {
    this.db = await initDatabase(this)
    this.cols = await this.db.addCollections(makeCollections(this))

    // import mock-data
    await loadMockData(this)
  }

  async makeRoutes(app) {
    return makeRoutes(app, this)
  }

  async fetchRemoteInstance (domain, id) {
    let res;
    try {
      const url = `https://${domain}/api/query/${id}`
      console.log(url)
      res = await fetch(url)
      if (!res) {
        return null
      }
    } catch (e) {
      console.log(e);
      return null
    }
    const json = await res.json()
    return json
  }

  async authorizeSession (req, reply) {
    const sessionId = req.headers['evermeet-session-id'] || req.cookies['evermeet-session-id'];
    if (!sessionId) {
        return [ false, null, 'no sessionId ("evermeet-session-id" header)' ]
    }
    const session = await this.cols.sessions.findOne({ selector: { id: sessionId }}).exec()
    if (!session) {
       return [ false, null, 'session not found' ] 
    }
    const user = await this.cols.users.findOne(session.user).exec()
    if (!user) {
        return [ false, null, 'user not found' ]
    }
    return [ true, user ]
  }

  async objectGet (id, opts={}) {
    let item = null;
    if (id.includes(':')) {
      const [ rid, domain ] = id.split(':')
      const res = await this.fetchRemoteInstance(domain, rid)
      if (res) {
        if (res.type === 'calendar') {
          // fix subevents _remote
          for (const ev of res.item.events) {
            ev._remote = domain
            ev.slug = ev.slug + ':' + domain
            ev.id = ev.id + ':' + domain
          }
        }
        return {
          type: res.type,
          item: Object.assign({ _remote: domain }, res.item, { slug: res.item.slug + ':' + domain, id: res.item.id + ':' + domain })
        }
      }
    } else {
      const cols = { calendars: 'calendar', events: 'event' }
      for (const c of Object.keys(cols)) {
        const found = await this.cols[c].findOne({ selector: { $or: [ { slug: id }, { id } ] }}).exec()
        if (found) {
          return {
            type: cols[c],
            item: await found.view(opts)
          }
        }
      }
    }
    return null
  }
}
import { Lexicons } from '@atproto/lexicon'
import xrpc from '@atproto/xrpc'

import pkg from '../../../package.json' with { type: "json" };
import endpoints from '../endpoints/index.js';

import { loadYaml, defaultsDeep, stringify, readdirSync, join } from './utils.js'
import { runtime, exitRuntime, env } from './runtime.js';
import { initDatabase, loadMockData } from './db.js'

export class Evermeet {

  constructor (opts = { configFile: '../../config.yaml' }) {
    this.runtime = runtime
    this.env = env('NODE_ENV') || 'development'
    const defaults = loadYaml('../../config.defaults.yaml')
    const localConfig = loadYaml(opts.configFile)
    this.config = defaultsDeep(localConfig, defaults)
    this.initialized = false
    this.pkg = pkg
    this.models = {}
    this.authVerifier = {
      accessUser: ({ user }) => {
        if (!user) {
          throw new Error('NotAuthorized')
        }
      },
      accessAdmin: ({ user }) => {
        if (!user) {
          throw new Error('NotAuthorized')
        }
        /*if (!user.roles.includes('admin')) {
          throw new Error('NotAuthorized')
        }*/
      }
    }
    
    console.log(stringify({ RUNTIME: this.runtime }))
    console.log(stringify({ CONFIG: this.config }))
    console.log('ENV:', this.env)
  }

  async init () {
    try {
      this.lexicons = await this.loadLexicons()

      this.db = await initDatabase(this, this.config.api.db)

      this.xrpc = xrpc.default || xrpc
      for (const lex of this.lexicons.docs) {
        this.xrpc.addLexicon(lex[1])
      }

      // compile & init endpoints
      // must be compiled before http adapter
      this.endpoints = await this.compileXrpcEndpoints(endpoints)

      // import mock-data
      await loadMockData(this)

      // load http adapter
      this.adapterMake = (await import("../adapters/" + this.config.api.adapter + ".js")).default
      this.adapterCtl = await this.adapterMake({ evermeet: this })
      this.adapter = await this.adapterCtl.init()

      // OLD `/api/*` optional for backward compatibility
      //makeRoutes(this.adapter, this)

    } catch (e) {
      console.error(e)
      this.exit()
    }
  }

  exit () {
    return exitRuntime()
  }

  async start () {
    if (!this.initialized) {
      await this.init()
    }
    await this.adapterCtl.start()
    console.log(`[${this.config.api.adapter}] HTTP adapter started at: ${this.config.api.host}:${this.config.api.port}`)
  }

  async loadLexicons() {
    const lexicons = new Lexicons()
    const lexiconDir = '../../lexicons';
    for (const ns of readdirSync(lexiconDir)) {
      for (const cat of readdirSync(join(lexiconDir, ns))) {
        for (const lex of readdirSync(join(lexiconDir, ns, cat))) {
          const defFn = join(lexiconDir, ns, cat, lex)
          const def = loadYaml(defFn)
          if (def) {
            const id = [ns, cat, lex.replace(/\.yaml$/, '')].join('.')
            const x = { id, ...def }
            lexicons.add(x)
            console.log(`Lexicon loaded: ${id}`)
          }
        }
      }
    }
    return lexicons
  }

  async compileXrpcEndpoints(_ep) {
    const struct = {}
    const list = []
    for (const ns of Object.keys(_ep)) {
      struct[ns] = {}
      for (const cat of Object.keys(_ep[ns])) {
        struct[ns][cat] = {}
        for (const cmd of Object.keys(_ep[ns][cat])) {
          const id = [ ns, cat, cmd ].join('.')
          const lex = this.lexicons.get(id)
          if (!lex) {
            console.error(`Endpoint (skipped): ${id} [missing lexicon]`)
            continue;
          }
          let handler;
          let auth;
          _ep[ns][cat][cmd]({
            endpoint: (opts) => {
              if (typeof opts === 'function') {
                handler = opts
              } else {
                handler = opts.handler
                auth = opts.auth
              }
            }
          }, {
            api: this
          })
          const endpoint = {
            id,
            lex,
            handler,
            auth,
          }
          struct[ns][cat][cmd] = endpoint
          list.push(endpoint)

          console.log(`Endpoint compiled: ${id}`)
        }
      }
    }
    //console.table(list)
    return { list, struct }
  }

  async request (id, { input, headers, session }) {
    const endpoint = this.endpoints.list.find(ep => ep.id === id)
    let out = {};
    try {
      // check input parameters
      this.lexicons.assertValidXrpcParams(id, input)
      // authorize session and basic context
      let authData = {};
      if (endpoint.auth) {
        const [ _, user, authError ] = await this.authorizeSession(session)
        authData = { user, authError }
        await endpoint.auth(authData)
      }
      const base = {
        db: this.db.getContext(),
        ...authData,
       }
      // run handler
      out = await endpoint.handler({ input, ...base })
      if (!out.error) {
        // if not error - validate output
        const res = this.lexicons.assertValidXrpcOutput(id, out.body)
      }
      } catch(e) {
        if (e.constructor.name !== 'Error') {
          out = { error: e.constructor.name, message: e.message }
        } else {
          out = { error: e.message }
        }
    }
    return out
  }

  async fetchRemoteInstance (domain, id) {
    let res, timeout;
    try {
      //timeout = setTimeout(() => {
      //  throw new Error
      //}, 200);
      const url = `https://${domain}/api/query/${id}`
      console.log('REMOTE:', url)
      res = await fetch(url, { timeout: 100 })
      //clearTimeout(timeout)
      if (!res) {
        return null
      }
    } catch (e) {
      console.log(e)
      return null
    }
    const json = await res.json()
    return json
  }

  async authorizeSession (sessionId) {

    return [ false, null ]

    const sessionName = this.config.api.sessionName
    if (!sessionId) {
      return [ false, null, `no sessionId ("${sessionName}" header)` ]
    }
    const session = await this.cols.sessions.findOne({ selector: { id: sessionId }}).exec()
    if (!session) {
      throw new AuthError('InvalidSession')
    }
    const user = await this.cols.users.findOne(session.user).exec()
    if (!user) {
      throw new AuthError('UserNotFound')
    }
    return [ true, user ]
  }

  async objectGet (db, id, opts={}) {
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
      const cols = { calendar: 'calendars', event: 'events' }
      for (const c of Object.keys(cols)) {
        const found = await db[cols[c]].findOne({ $or: [ { slug: id }, { _id: id } ] })
        if (found) {
          return {
            type: c,
            item: await found.view(opts)
          }
        }
      }
    }
    return null
  }

  internalEndpoints () {
    return [
      {
        id: '_health',
        handler: () => {
          return { ok: true }
        }
      },
      {
        id: '_lexicons',
        handler: () => {
          const out = []
          for (const [id, def] of this.lexicons.docs) {
            out.push(def)
          }
          return out
        }
      },
      {
        id: '_config',
        handler: () => {
          const c = this.config
          return {
            domain: c.domain,
            sitename: c.sitename,
            options: c.options,
          }
        }
      }
    ]
  }
}

class AuthError extends Error {}
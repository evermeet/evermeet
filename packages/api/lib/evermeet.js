import logger from 'pino'

import { Lexicons } from '@atproto/lexicon'
import * as xrpc from '@atproto/xrpc'

import pkg from '../../../package.json' with { type: "json" };
import endpoints from '../endpoints/index.js';

import { loadYaml, loadYamlDir, loadYamlDirList, stringify, readdirSync, join } from './utils.js'
import { runtime, exitRuntime, env } from './runtime.js';
import { initDatabase } from './db.js'
import { loadMockData } from './mock.js';
import { loadConnectors } from './connector.js';
import { authVerifier } from './auth.js';
import { loadConfig } from './config.js';
import { BlobStore } from './blobstore.js';

export class Evermeet {
  
  constructor (opts = { configFile: '../../config.yaml' }) {
    this.paths = {
      lexicons: '../../lexicons',
      entities: './entities',
      schema: './schema',
      mockData: './mock-data',
      configDefaults: '../../config.defaults.yaml',
      // `adapters` need be relative to script because its dynamic import
      adapters: '../adapters/',
    }
    this.initialized = false
    this.runtime = runtime
    this.env = env('NODE_ENV') || 'development'
    this.schema = loadYamlDir(this.paths.schema)
    this.config = loadConfig(opts.configFile, this.paths.configDefaults, this.schema.config)
    this.pkg = pkg
    this.authVerifier = authVerifier
    this.network = this.config.network || 'sandbox'

    // check if runtime if config is correct
    if (runtime.name !== this.config.api.runtime) {
      throw new Error(`Wrong runtime! in config: ${this.config.api.runtime}, current: ${runtime.name}`)
    }

    this.logLevel = this.config.api.trace ? 'trace' : (this.env === 'development' ? 'info' : 'error')
    this.logTransport = this.env === 'development' && {
      target: 'pino-pretty',
      options: {
        colorize: true
      }
    }
    this.logger = logger({
      level: this.logLevel,
      transport: this.logTransport,
    })
    this.logger.info({ domain: this.config.domain, network: this.network, env: this.env, runtime: this.runtime }, `${pkg.name} ${pkg.version}`)
  }

  async init () {
    try {
      // initialize lexicons
      this.lexicons = new Lexicons()
      for (const { id, data } of loadYamlDirList(this.paths.lexicons)) {
        this.lexicons.add({ id, ...data })
      }

      // initialize database
      this.db = await initDatabase(this, this.config.api.db)

      // init blobStore
      this.blobStore = new BlobStore(this)

      // init connectors
      this.connectors = await loadConnectors(this)

      // construct XPRC client and add lexicons
      this.xrpc = new xrpc.Client()
      for (const lex of this.lexicons.docs) {
        this.xrpc.addLexicon(lex[1])
      }

      // compile & init endpoints
      // must be compiled before http adapter
      this.endpoints = await this.compileXrpcEndpoints(endpoints)

      // import mock-data
      await loadMockData(this)

      // load http adapter
      this.adapterMake = (await import(this.paths.adapters + this.config.api.adapter + ".js")).default
      this.adapterCtl = await this.adapterMake({ evermeet: this })
      this.adapter = await this.adapterCtl.init()

      // OLD `/api/*` optional for backward compatibility
      //makeRoutes(this.adapter, this)

    } catch (e) {
      console.error(e)
      this.exit()
    }
  }

  async start () {
    if (!this.initialized) {
      await this.init()
    }
    await this.adapterCtl.start()
    this.logger.info({ adapter: this.config.api.adapter, host: this.config.api.host, port: this.config.api.port }, 'HTTP server started')
    this.logger.info('@evermeet/api started')
  }

  exit () {
    return exitRuntime()
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
        }
      }
    }
    this.logger.debug({ endpoints: list.map(e => e.id) }, `XRPC endpoints compiled (${list.length})`)
    return { list, struct }
  }

  async request (id, { input, encoding, headers, session }) {
    const endpoint = this.endpoints.list.find(ep => ep.id === id)
    let out = {};
    try {
      // check input parameters
      const lex = this.lexicons.get(id)
      if (lex.defs.main.type === 'procedure') {
        this.lexicons.assertValidXrpcInput(id, input)
      } else {
        this.lexicons.assertValidXrpcParams(id, input)
      }
      // authorize session and basic context
      const db = this.db.getContext()
      const [ _, user, authError ] = await this.authorizeSession(session, { db })
      const authData = { user, authError };
      if (endpoint.auth) {
        await endpoint.auth(authData)
      }
      // run handler
      this.logger.debug({ endpoint: id, user: authData?.user?.did || 'nil' }, 'API request')
      out = await endpoint.handler({ input, encoding, ...authData, db, api: this })
      if (!out.error) {
        // if not error - validate output
        const res = this.lexicons.assertValidXrpcOutput(id, out.body)
      }
      } catch(e) {
        this.logger.error(e)
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

  async authorizeSession (sessionId, { db }) {

    const sessionName = this.config.api.sessionName
    if (!sessionId) {
      return [ false, null, `no sessionId ("${sessionName}" header)` ]
    }
    const session = await db.sessions.findOne({ token: sessionId })
    if (!session) {
      throw new AuthError('InvalidSession')
    }
    const user = await db.users.findOne({ id: session.userId })
    if (!user) {
      throw new AuthError('UserNotFound')
    }
    return [ true, user ]
  }

  async objectGet (ctx, id, opts={}) {
    id = id.toLowerCase()
    let item = null;    
    /*(if (id.includes(':')) {
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
    } else {*/

    // if have `/` we know its event
    if (id.match(/:/)) {
      const [ calendarDid, eventId ] = id.split(':')
      const calendar = await ctx.db.calendars.findOne({ $or: [ { handle: calendarDid }, { handle: `${calendarDid}.${this.config.domain}` } ] })
      if (calendar) {
        const found = await ctx.db.events.findOne({ slug: eventId })
        if (found) {
          return {
            type: 'event',
            item: await found.view(ctx, opts)
          }
        }
      }
    }

    // otherwise ...
    const cols = { calendar: 'calendars', event: 'events', user: 'users' }
    for (const c of Object.keys(cols)) {
      const found = await ctx.db[cols[c]].findOne({ $or: [ { handle: id }, { handle: `${id}.${this.config.domain}` }, { id }, { did: id } ] })
      if (found) {
        return {
          type: c,
          item: await found.view(ctx, opts)
        }
      }
    }
    return null
  }

  findConnector (url) {
    for (const c of Object.keys(this.connectors)) {
      const conn = this.connectors[c]
      for (const re of conn.urlPatterns) {
        if (url.match(re)) {
          return conn
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
            plcServer: c.plcServer,
            network: this.network,
            env: this.env,
          }
        }
      }
    ]
  }
}

class AuthError extends Error {}
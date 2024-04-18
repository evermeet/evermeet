
import { parse, stringify } from 'yaml'
import _ from 'lodash'
import fs, { chmod } from 'node:fs'
import path from 'node:path'
import { Lexicons } from '@atproto/lexicon'

import { initDatabase, loadMockData } from './db.js'
import { makeRoutes } from './routes.js';
import { makeCollections } from './collections.js';
import { runtime, env } from './runtime.js';
import endpoints from '../endpoints/index.js';
import pkg from '../../../package.json' with { type: "json" };

class AuthError extends Error {}

export class Evermeet {

  constructor (opts) {
    this.runtime = runtime
    this.env = env('NODE_ENV') || 'development'
    const defaults = parse(fs.readFileSync('../../config.defaults.yaml', 'utf-8'))
    const localConfig = parse(fs.readFileSync('../../config.yaml', 'utf-8'))
    this.config = _.defaultsDeep(localConfig, defaults)
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
      this.db = await initDatabase(this)
      this.cols = await this.db.addCollections(makeCollections(this))
      this.endpoints = await this.loadEndpoints()

      // import mock-data
      await loadMockData(this)

      // load http adapter
      this.adapter = (await import("../adapters/" + this.config.api.adapter + ".js")).default({ evermeet: this })
      await this.adapter.init()

    } catch (e) {
      console.error(e)
      this.exit()
    }
  }

  exit () {
    if (Deno) {
      Deno.exit(1)
    } else {
      process.exit(1)
    }
  }

  async start () {
    if (!this.initialized) {
      await this.init()
    }
    await this.adapter.start()
    console.log(`[${this.config.api.adapter}] HTTP adapter started at: ${this.config.api.host}:${this.config.api.port}`)
  }

  async loadLexicons() {
    const lexicons = new Lexicons()
    const lexiconDir = '../../lexicons';
    for (const ns of fs.readdirSync(lexiconDir)) {
      for (const cat of fs.readdirSync(path.join(lexiconDir, ns))) {
        for (const lex of fs.readdirSync(path.join(lexiconDir, ns, cat))) {
          const defFn = path.join(lexiconDir, ns, cat, lex)
          const def = parse(fs.readFileSync(defFn, 'utf-8'))
          if (def) {
            const id = [ns, cat, lex.replace(/\.yaml$/, '')].join('.')
            lexicons.add({ id, ...def })
            console.log(`Lexicon loaded: ${id}`)
          }
        }
      }
    }
    return lexicons
  }

  async loadEndpoints() {
    const struct = {}
    const list = []
    for (const ns of Object.keys(endpoints)) {
      struct[ns] = {}
      for (const cat of Object.keys(endpoints[ns])) {
        struct[ns][cat] = {}
        for (const cmd of Object.keys(endpoints[ns][cat])) {
          const id = [ ns, cat, cmd ].join('.')
          const lex = this.lexicons.get(id)
          if (!lex) {
            console.error(`Endpoint (skipped): ${id} [missing lexicon]`)
            continue;
          }
          let handler;
          let auth;
          endpoints[ns][cat][cmd]({
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

          console.log(`Endpoint loaded: ${id}`)
        }
      }
    }
    //console.log(struct['app.evermeet'].server.describeServer)
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
      const base = { ...authData }
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

  async makeRoutes(app) {
    for (const ns of Object.keys(endpoints)) {
      for (const cat of Object.keys(endpoints[ns])) {
        for (const endpoint of Object.keys(endpoints[ns][cat])) {
          const id = [ ns, cat, endpoint ].join('.')
          const url = `/xrpc/${id}`

          let handler;
          let config = {};
          endpoints[ns][cat][endpoint]({
            endpoint: (opts) => {
              if (typeof opts === 'function') {
                handler = opts
              } else {
                config = opts
                handler = opts.handler
              }
            }
          }, {
            api: this
          })
          if (handler) {
            const lex = this.lexicons.get(id)
            if (!lex) {
              console.error(`Missing lexicon: ${id}, skipping this handler`)
              continue;
            }
            const callback = async (req, reply) => {
              let out = {};
              try {
                const [ _, user, authError ] = await this.authorizeSession(req, reply)
                const input = req.query

                this.lexicons.assertValidXrpcParams(id, input)
                out = await handler({ input, user, authError })
                if (!out.error) {
                  this.lexicons.assertValidXrpcOutput(id, out.body)
                }

              } catch(e) {
                out = { error: e.constructor.name, message: e.message }
              }
              if (out.error) {
                return reply.code(501).send({ error: out.error, message: out.message })
              }
              return reply.send(out.body)
            }

            //console.log(lex.defs.main.input.schena)
            app.get(url, callback)
          }
        }
      }
    }

    return makeRoutes(app, this)
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
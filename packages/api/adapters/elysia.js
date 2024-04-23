import { Elysia, t } from 'elysia'
import { swagger } from '@elysiajs/swagger'
import { logger } from '@bogeychan/elysia-logger'
import { cors } from '@elysiajs/cors'

export default function ({ evermeet }) {
  let app
  return {
    async init () {
      // check environment
      if (evermeet.runtime.name !== 'bun') {
        throw new Error(`Elysia: Only works with Bun runtime (current=${evermeet.runtime.name})`)
      }

      // start adapter
      app = new Elysia({
        prefix: evermeet.config.api.prefix
      })
        .use(cors())
        .use(swagger({
          path: '/_swagger',
          excludeStaticFile: false,
          exclude: [
            '/xrpc/_swagger',
            '/xrpc/_swagger/json'
          ],
          documentation: {
            info: {
              title: 'evermeet HTTP API (XRPC)',
              version: evermeet.pkg.version,
              description: 'More info: https://docs.evermeet.app/specs/xrpc'
            },
            paths: {
              '/xrpc/_swagger': {
                get: {
                  description: 'Swagger UI'
                }
              }
            },
            swaggerOptions: {
              persistAuthorization: true
            }
          }
        }))
        .use(logger({ level: process.env.NODE_ENV === 'development' ? 'debug' : 'error' }))

      for (const ep of evermeet.endpoints.list) {
        const method = ep.lex.defs.main.type === 'procedure' ? 'post' : 'get'

        app[method](ep.id, async ({ error, query, body, headers, cookie }) => {
          let out = {}
          const input = ep.lex.defs.main.type === 'procedure' ? body : query
          const session = headers.authorization?.replace(/^Bearer /, '') || cookie[evermeet.config.api.sessionName].value
          out = await evermeet.request(ep.id, { input, headers, session })
          if (out.error) {
            return error(501, { error: out.error, message: out.message })
          }
          if (out.cookies) {
            for (const cd of out.cookies) {
              cookie[cd.key].set({ ...cd.config })
            }
          }
          return out.body
        }, {
          query: ep.lex.defs.main.type === 'query' ? t.Any(ep.lex.defs.main.parameters) : null,
          body: ep.lex.defs.main.type === 'procedure' ? t.Any(ep.lex.defs.main.input?.schema) : null,
          response: {
            200: t.Any(ep.lex.defs.main.output.schema),
            501: t.Object({
              error: t.String(),
              message: t.Optional(t.String())
            })
          },
          detail: {
            summary: ep.id.replace('app.evermeet.', ''),
            description: ep.lex.defs.main.description,
            tags: ['app.evermeet.*']
          }
        })
        console.log(`[elysia] Route created: [${method.toUpperCase()}] ${evermeet.config.api.prefix}/${ep.id}`)
      }
      for (const ih of evermeet.internalEndpoints()) {
        const url = ih.id
        app.get(url, ih.handler)
        console.log(`[elysia] Internal route created: [${ih.id}] ${url}`)
      }
      return app
    },
    async start () {
      return app.listen({
        hostname: evermeet.config.api.host,
        port: evermeet.config.api.port
      })
    },
    getCookie (req, key) {
      return req.cookie[key]?.value
    }
  }
}

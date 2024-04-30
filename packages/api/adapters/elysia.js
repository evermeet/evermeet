import { Elysia, t } from 'elysia'
import { swagger } from '@elysiajs/swagger'
import { cors } from '@elysiajs/cors'
// import { logger as httpLogger } from '@bogeychan/elysia-logger'

export default function ({ evermeet }) {
  let app
  return {
    async init () {
      // check environment
      /* if (evermeet.runtime.name !== 'bun') {
        throw new Error(`Elysia: Only works with Bun runtime (current=${evermeet.runtime.name})`)
      } */
      const logger = evermeet.logger.child({ module: 'http', adapter: 'elysia' })

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
        /* .use(httpLogger({
          level: evermeet.logLevel === 'debug' ? 'info' : 'error',
          transport: evermeet.logTransport
        })) */

      for (const ep of evermeet.endpoints.list) {
        const method = ep.lex.defs.main.type === 'procedure' ? 'post' : 'get'

        app[method](ep.id, async ({ error, query, body, headers, cookie, request, set }) => {
          let out = {}
          let input = ep.lex.defs.main.type === 'procedure' ? body : query
          const encoding = headers['content-type'] || 'application/json'

          if (input === undefined && ep.lex.defs.main.type === 'procedure' && ep.lex.defs.main.input.encoding === '*/*') {
            // decode binary data
            input = await request.arrayBuffer()
          }

          const session = headers.authorization?.replace(/^Bearer /, '') || cookie[evermeet.config.api.sessionName].value
          out = await evermeet.request(ep.id, { input, encoding, headers, session })
          if (out.error) {
            return error(501, { error: out.error, message: out.message })
          }
          if (out.cookies) {
            for (const cd of out.cookies) {
              cookie[cd.key].set({ ...cd.config })
            }
          }
          if (out.encoding && out.encoding !== 'application/json') {
            set.headers['content-type'] = out.encoding

            return Buffer.from(out.body)
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
        logger.trace({ method: method.toUpperCase(), route: `${evermeet.config.api.prefix}/${ep.id}` }, 'Route created')
      }
      for (const ih of evermeet.internalEndpoints()) {
        const url = ih.id
        app.get(url, ih.handler)
        logger.trace({ method: ih.id, route: url }, 'Internal route created')
      }
      return app
    },
    async start () {
      if (evermeet.runtime.name === 'bun') {
        return app.listen({
          hostname: evermeet.config.api.host,
          port: evermeet.config.api.port
        })
      } else if (evermeet.runtime.name === 'deno') {
        return Deno.serve({
          hostname: evermeet.config.api.host,
          port: evermeet.config.api.port,
          handler: app.fetch
        })
      } else {
        throw new Error(`Runtime not supported: ${evermeet.runtime.name}`)
      }
    },
    getCookie (req, key) {
      return req.cookie[key]?.value
    }
  }
}

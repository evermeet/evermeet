import Fastify from 'fastify'
import cors from '@fastify/cors'
import cookie from '@fastify/cookie';
import middie from '@fastify/middie';

export default function ({ evermeet }) {
    let app;
    return {
        async init () {
            app = Fastify({ logger: evermeet.env === 'development' })
        
            await app.register(middie)
            await app.register(cors, {
              origin: evermeet.env === 'development' 
                ? `http://${evermeet.config.web.host}:${evermeet.config.web.port}`
                : `https://${evermeet.config.domain}`,
              credentials: true
            })
            
            app.addContentTypeParser('application/json', { parseAs: 'string' }, function (req, body, done) {
                try {
                  var json = JSON.parse(body)
                  done(null, json)
                } catch (err) {
                  err.statusCode = 400
                  done(err, undefined)
                }
            })
            
            app.register(cookie, {
                //secret: "my-secret", // for cookies signature
                //hook: 'onRequest', // set to false to disable cookie autoparsing or set autoparsing on any of the following hooks: 'onRequest', 'preParsing', 'preHandler', 'preValidation'. default: 'onRequest'
                parseOptions: {}  // options for parsing cookies
            })

            for (const ep of evermeet.endpoints.list) {
                const url = evermeet.config.api.prefix + '/' + ep.id
                const method = ep.lex.defs.main.type === "procedure" ? 'post' : 'get'

                app[method](url, async (req, reply) => {
                    let out = {};
                    const input = ep.lex.defs.main.type === "procedure" ? req.body : req.query
                    out = await evermeet.request(ep.id, { input, headers: req.headers, session: req.cookies[evermeet.config.api.sessionName] })
                    if (out.error) {
                        return reply.code(501).send({ error: out.error, message: out.message })
                    }
                    return reply.send(out.body)
                })
                console.log(`[fastify] Route created: [${method.toUpperCase()}] ${url}`)
            }
            for (const ih of evermeet.internalEndpoints()) {
                const url = evermeet.config.api.prefix + '/' + ih.id
                app.get(url, ih.handler)
                console.log(`[fastify] Internal route created: [${ih.id}] ${url}`)
            }

            return app
        },
        async start () {
            return app.listen({ port: evermeet.config.api.port, host: evermeet.config.api.host }, function (err, address) {
                if (err) {
                    app.log.error(err)
                    console.error(err)
                    evermeet.exit()
                }
                // Server is now listening on ${address}
            })
        },
        getCookie (req, key) {
            return req.cookies[key]
        }
    }
}


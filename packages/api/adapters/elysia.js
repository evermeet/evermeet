import { Elysia, t } from 'elysia'
import { swagger } from '@elysiajs/swagger'

export default function ({ evermeet }) {
    let app;
    return {
        async init () {
            app = new Elysia()
            app.use(swagger({ excludeStaticFile: false }))

            for (const ep of evermeet.endpoints.list) {
                console.log(ep.lex.defs.main.output.schema)
                const url = evermeet.config.api.prefix + '/' + ep.id
                const method = ep.lex.defs.main.type === "procedure" ? 'post' : 'get'

                app[method](url, async ({ error, query, body, headers, cookie }) => {
                    let out = {};   
                    const input = ep.lex.defs.main.type === "procedure" ? body : query
                    out = await evermeet.request(ep.id, { input, headers, session: cookie[evermeet.config.api.sessionName] })
                    if (out.error) {
                        return error(501, { error: out.error, message: out.message })
                    }
                    return out.body
                }, {
                    //query: ep.lex.defs.main.parameters
                    //response: ep.lex.defs.main.output.schema
                })
                console.log(`[elysia] Route created: [${method.toUpperCase()}] ${url}`)
            }
            return app
        },
        async start () {
            return app.listen({
                hostname: evermeet.config.api.host,
                port: evermeet.config.api.port
            })
        }
    }
}
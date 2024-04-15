import { API } from './lib/api.js';
import Fastify from 'fastify'
import cors from '@fastify/cors'
import cookie from '@fastify/cookie';
import middie from '@fastify/middie';

const api = new API()
await api.init()

const app = Fastify({ logger: process.env.NODE_ENV === 'development' })

await app.register(middie)
await app.register(cors, {
  //origin: `http://localhost:${api.config.web.port},http://${api.config.web.host}:${api.config.web.port},https://${api.config.domain}`,
  origin: process.env.NODE_ENV === 'development' 
    ? `http://${api.config.web.host}:${api.config.web.port}`
    : `https://${api.config.domain}`,
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


api.makeRoutes(app)

app.listen({ port: api.config.api.port, host: api.config.api.host }, function (err, address) {
    if (err) {
        app.log.error(err)
        process.exit(1)
    }
    // Server is now listening on ${address}
})
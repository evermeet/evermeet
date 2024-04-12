import { API } from './lib/api.js';
import Fastify from 'fastify'
import cors from '@fastify/cors'
import cookie from '@fastify/cookie';
import middie from '@fastify/middie';

import { handler } from '../web/build/handler.js';

const api = new API()
await api.init()

const app = Fastify({ logger: true })

await app.register(middie)
await app.register(cors, {
  origin: 'http://localhost:5173',
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

app.listen({ port: 3000, host: "0.0.0.0" }, function (err, address) {
    if (err) {
        app.log.error(err)
        process.exit(1)
    }
    // Server is now listening on ${address}
})
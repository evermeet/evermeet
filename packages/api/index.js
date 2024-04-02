import Fastify from 'fastify'
import cors from '@fastify/cors'
import { initDatabase } from './lib/db.js'


const db = await initDatabase()

function eventView (event, opts = {}) {
  event = Object.assign({}, event)
  if (event.calendarId) {
    const cal = db.calendars.findOne({ id: event.calendarId })
    if (opts.calendar !== false && cal) {
      event.calendar = cal
    }
  }
  return event
}
function calendarView (calendar) {
  calendar = Object.assign({}, calendar)
  calendar.events = db.events.find({ calendarId: calendar.id }).fetch().map(e => eventView(e, { calendar: false }))
  return calendar
}

const api = Fastify({ logger: true })
api.register(cors)

// Declare a routes
api.get('/', function (request, reply) {
  reply.send({ hello: 'world' })
})

api.get('/query/:id', (req, reply) => {
  const { id } = req.params;

  const calendar = db.calendars.findOne({ slug: id })
  if (calendar) {
    return reply.send({
      type: 'calendar',
      item: calendarView(calendar)
    })
  }

  const event = db.events.findOne({ slug: id })
  if (event) {
    return reply.send({
      type: 'event',
      item: eventView(event)
    })
  }

  return reply.code(404).send({ error: 'notfound' })
})

api.get('/events', (req, reply) => {
  const cursor = db.events.find({})
  reply.send({ events: cursor.fetch() })
})

api.get('/explore', (req, reply) => {
    const cursor = db.calendars.find({})
    reply.send({ calendars: cursor.fetch().filter(c => !c.personal) })
})

api.get('/calendars', (req, reply) => {
    const cursor = db.calendars.find({})
    reply.send(cursor.fetch())
})

api.get('/calendar/:id', (req, reply) => {
  const { id } = req.params;
  return db.calendars.findOne({ id })
})

// Run the server!
api.listen({ port: 3000 }, function (err, address) {
  if (err) {
    api.log.error(err)
    process.exit(1)
  }
  // Server is now listening on ${address}
})
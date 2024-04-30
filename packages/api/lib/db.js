import { wrap } from '@mikro-orm/core'
import { createId as cuid2 } from '@paralleldrive/cuid2'
import ShortUniqueId from 'short-unique-id'
const { MikroORM, RequestContext } = await import(`@mikro-orm/libsql`)

export async function initDatabase (api, conf) {

  const orm = await MikroORM.init({
    dbName: conf.name,
    entities: [api.paths.entities],
    debug: api.env === 'development',
    logger: msg => api.logger.trace(msg)
  })

  await orm.schema.refreshDatabase()

  api.logger.info({ storage: 'sqlite' }, 'Database initialized')
  return {
    orm,
    em: orm.em,
    getContext () {
      const em = orm.em.fork()
      return {
        em,
        calendars: em.getRepository('Calendar'),
        concepts: em.getRepository('Concept'),
        events: em.getRepository('Event'),
        users: em.getRepository('User'),
        sessions: em.getRepository('Session'),
        blobs: em.getRepository('Blob'),
        messages: em.getRepository('Message'),
        rooms: em.getRepository('Room'),
        wrap
      }
    }
  }
}

export function createId () {
  return cuid2()
}

export function createShortId () {
  return (new ShortUniqueId({ length: 5 }).rnd())
}

export function ObjectId () {
  const timestamp = (new Date().getTime() / 1000 | 0).toString(16)
  return timestamp + 'xxxxxxxxxxxxxxxx'.replace(/[x]/g, function () {
    return (Math.random() * 16 | 0).toString(16)
  }).toLowerCase()
}

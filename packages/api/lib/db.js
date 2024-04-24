import { wrap } from '@mikro-orm/core'
import { createId as cuid2 } from '@paralleldrive/cuid2'

export async function initDatabase (api, conf) {
  console.log(`[Database] Loadind storage: ${conf.storage}`)

  const { MikroORM, RequestContext } = await import(`@mikro-orm/${conf.storage}`)

  const orm = await MikroORM.init({
    dbName: conf.name,
    entities: [api.paths.entities],
    debug: api.env === 'development'
  })

  await orm.schema.refreshDatabase()

  console.log('[Database] Storage initialized')
  return {
    orm,
    em: orm.em,
    getContext () {
      const em = orm.em.fork()
      return {
        em,
        calendars: em.getRepository('Calendar'),
        events: em.getRepository('Event'),
        users: em.getRepository('User'),
        sessions: em.getRepository('Session'),
        blobs: em.getRepository('Blob'),
        wrap
      }
    }
  }
}

export function createId () {
  return cuid2()
}

export function ObjectId () {
  const timestamp = (new Date().getTime() / 1000 | 0).toString(16)
  return timestamp + 'xxxxxxxxxxxxxxxx'.replace(/[x]/g, function () {
    return (Math.random() * 16 | 0).toString(16)
  }).toLowerCase()
}

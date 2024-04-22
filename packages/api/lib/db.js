import { wrap } from '@mikro-orm/core'
import { loadYaml, join } from './utils.js'

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
        wrap
      }
    }
  }
}

export async function loadMockData (api) {
  const db = api.db

  const map = [
    ['calendars.yaml', 'Calendar'],
    ['events.yaml', 'Event'],
    ['users.yaml', 'User'],
    ['sessions.yaml', 'Session']
  ]

  const em = db.em.fork()

  for (const [fn, entityName] of map) {
    const items = await loadYaml(join(api.paths.mockData, fn))
    const repo = em.getRepository(entityName)
    for (const item of items) {
      if (item.id) {
        item._id = item.id
        delete item.id
      }
      const x = repo.create(item)
      em.persist(x)
    }
  }
  await em.flush()
}

export function ObjectId () {
  const timestamp = (new Date().getTime() / 1000 | 0).toString(16)
  return timestamp + 'xxxxxxxxxxxxxxxx'.replace(/[x]/g, function () {
    return (Math.random() * 16 | 0).toString(16)
  }).toLowerCase()
};

import { wrap } from '@mikro-orm/core';
import { loadYaml, join } from './utils.js'

export async function initDatabase(api, conf) {
    console.log(`[Database] Loadind storage: ${conf.storage}`)

    const { MikroORM, RequestContext } = await import(`@mikro-orm/${conf.storage}`)

    const orm = await MikroORM.init({
        dbName: conf.name,
        entities: ['./entities'],
        debug: api.env === 'development',
    });

    await orm.schema.refreshDatabase();

    console.log('[Database] Storage initialized')
    return {
        orm,
        em: orm.em,
        getContext () {
            const em = orm.em.fork()
            return {
                em,
                calendars: em.getRepository('Calendar'),
                wrap,
            }
        }
    }
}

export async function loadMockData (api) {
    const db = api.db

    const mockDir = './mock-data'
    const map = [
        ['calendars.yaml', 'Calendar'],
    ]

    const em = db.em.fork()

    for (const [ fn, entityName ] of map) {
        const items = await loadYaml(join(mockDir, fn))
        const repo = em.getRepository(entityName)
        for (const item of items) {
            item._id = item.id
            delete item.id
            const x = repo.create(item)
            em.persist(x)
        }
    }
    await em.flush()
}

export function ObjectId () {
    var timestamp = (new Date().getTime() / 1000 | 0).toString(16);
    return timestamp + 'xxxxxxxxxxxxxxxx'.replace(/[x]/g, function() {
        return (Math.random() * 16 | 0).toString(16);
    }).toLowerCase();
};
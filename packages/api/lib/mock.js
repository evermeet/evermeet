import { loadYaml, loadFile, loadFileWithInfo, join, detect } from './utils.js'

export async function loadMockData (api) {
  const db = api.db

  const map = [
    ['calendars.yaml', 'Calendar'],
    ['events.yaml', 'Event'],
    ['users.yaml', 'User'],
    ['sessions.yaml', 'Session'],
    ['blobs.yaml', 'Blob']
  ]

  const loadBlob = (cid) => {
    return loadFileWithInfo(join(api.paths.mockData, 'blobs', cid), 'buffer')
  }

  const em = db.em.fork()

  for (const [fn, entityName] of map) {
    const items = await loadYaml(join(api.paths.mockData, fn))
    const repo = em.getRepository(entityName)
    for (const item of items) {
      if (entityName === 'Blob') {
        const b = loadBlob(item.cid)
        item.data = b.data
        item.size = b.size
        item.mimeType = detect(b.data)
      }
      const x = repo.create(item)
      em.persist(x)
    }
  }
  await em.flush()
}

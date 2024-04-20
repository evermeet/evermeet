import { createRxDatabase, addRxPlugin } from 'rxdb'
import { getRxStorageMongoDB } from 'rxdb/plugins/storage-mongodb'
import { getRxStorageFoundationDB } from 'rxdb/plugins/storage-foundationdb'
import { wrappedKeyEncryptionCryptoJsStorage } from 'rxdb/plugins/encryption-crypto-js'
import { RxDBMigrationPlugin } from 'rxdb/plugins/migration-schema'
import { RxDBDevModePlugin } from 'rxdb/plugins/dev-mode'
import { RxDBUpdatePlugin } from 'rxdb/plugins/update'

import { parse } from 'yaml'
import fs from 'node:fs'

export async function initDatabase (api) {
  const conf = api.config.api.db

  addRxPlugin(RxDBMigrationPlugin)
  if (api.env === 'development') {
    addRxPlugin(RxDBDevModePlugin)
  }
  addRxPlugin(RxDBUpdatePlugin)

  let storage = null
  if (conf.storage === 'mongodb') {
    storage = getRxStorageMongoDB({
      /**
             * MongoDB connection string
             * @link https://www.mongodb.com/docs/manual/reference/connection-string/
             */
      connection: conf.connection
    })
    console.log(`MongoDB storage inicialized: ${conf.connection}`)
  } else if (conf.storage === 'foundationdb') {
    storage = getRxStorageFoundationDB({
      /**
             * Version of the API of the FoundationDB cluster..
             * FoundationDB is backwards compatible across a wide range of versions,
             * so you have to specify the api version.
             * If in doubt, set it to 620.
             */
      apiVersion: 620,
      /**
             * Path to the FoundationDB cluster file.
             * (optional)
             * If in doubt, leave this empty to use the default location.
             */
      clusterFile: '/path/to/fdb.cluster',
      /**
             * Amount of documents to be fetched in batch requests.
             * You can change this to improve performance depending on
             * your database access patterns.
             * (optional)
             * [default=50]
             */
      batchSize: 50
    })
  }

  const encryptedStorage = wrappedKeyEncryptionCryptoJsStorage({
    name: conf.name,
    storage
  })

  const db = await createRxDatabase({
    name: conf.name,
    storage: encryptedStorage,
    password: conf.password
  })

  // console.log(mongoStorage, db)
  // await db.waitForLeadership()
  // await (new Promise((resolve) => setTimeout(resolve, 5000)))

  console.log('database initialized')

  return db
}

async function loadMock (fn) {
  const res = fs.readFileSync('./mock-data/' + fn + '.yaml', 'utf-8')
  return parse(res)
}

function ObjectId () {
  const timestamp = (new Date().getTime() / 1000 | 0).toString(16)
  return timestamp + 'xxxxxxxxxxxxxxxx'.replace(/[x]/g, function () {
    return (Math.random() * 16 | 0).toString(16)
  }).toLowerCase()
};

export async function loadMockData (api) {
  // setup mock data
  const db = api.db

  const cols = [
    'events',
    'calendars',
    'users'
  ]

  for (const c of cols) {
    await api.cols[c].find({}).remove()
    const items = await loadMock(c)
    items.map(async (x) => await api.cols[c].insert(Object.assign({ id: ObjectId() }, x)))
  }
}

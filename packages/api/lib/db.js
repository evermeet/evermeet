import { createRxDatabase, addRxPlugin } from 'rxdb';
import { getRxStorageMongoDB } from 'rxdb/plugins/storage-mongodb';
import { wrappedKeyEncryptionCryptoJsStorage } from 'rxdb/plugins/encryption-crypto-js';
//import { RxDBDevModePlugin } from 'rxdb/plugins/dev-mode';
//addRxPlugin(RxDBDevModePlugin);

import { parse } from 'yaml'
import fs from 'node:fs'

export async function initDatabase(api) {

    const encryptedStorage = wrappedKeyEncryptionCryptoJsStorage({
        name: api.config.db.name,
        storage: getRxStorageMongoDB({
            /**
             * MongoDB connection string
             * @link https://www.mongodb.com/docs/manual/reference/connection-string/
             */
            connection: api.config.db.connection
        })
    })

    const db = await createRxDatabase({
        name: api.config.db.name,
        storage: encryptedStorage,
        password: api.config.db.password
    });

    console.log('database initialized')

    return db
}

async function loadMock(fn) {
    const res = fs.readFileSync('./mock-data/' + fn + '.yaml', 'utf-8')
    return parse(res)
}

function ObjectId () {
    var timestamp = (new Date().getTime() / 1000 | 0).toString(16);
    return timestamp + 'xxxxxxxxxxxxxxxx'.replace(/[x]/g, function() {
        return (Math.random() * 16 | 0).toString(16);
    }).toLowerCase();
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
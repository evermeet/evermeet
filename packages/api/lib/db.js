import { createRxDatabase, addRxPlugin } from 'rxdb';
import { getRxStorageMongoDB } from 'rxdb/plugins/storage-mongodb';
import { wrappedKeyEncryptionCryptoJsStorage } from 'rxdb/plugins/encryption-crypto-js';
//import { RxDBDevModePlugin } from 'rxdb/plugins/dev-mode';
//addRxPlugin(RxDBDevModePlugin);

import { parse } from 'yaml'
import fs from 'node:fs'

export async function initDatabase() {

    const encryptedStorage = wrappedKeyEncryptionCryptoJsStorage({
        name: 'deluma-test',
        storage: getRxStorageMongoDB({
            /**
             * MongoDB connection string
             * @link https://www.mongodb.com/docs/manual/reference/connection-string/
             */
            connection: 'mongodb://localhost:27017'
        })
    })

    const db = await createRxDatabase({
        name: 'deluma-test',
        storage: encryptedStorage,
        password: 'yeibae8ceX4japhukuyaegeiZ8pha6uR'
    });

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
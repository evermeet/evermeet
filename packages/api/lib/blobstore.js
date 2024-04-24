import { CID } from 'multiformats/cid'
import * as raw from 'multiformats/codecs/raw'
import { sha256 } from 'multiformats/hashes/sha2'
import { detect } from './utils.js'

export class BlobStore {
  constructor (api) {
    this.api = api
  }

  async add (input, { db }) {
    const data = new Uint8Array(input)
    if (data.length === 0) {
      throw new Error('CannoBeEmpty')
    }
    const hash = await sha256.digest(data)
    const cid = CID.create(1, raw.code, hash)
    // check if exists
    const found = await db.blobs.findOne({ cid: cid.toString() })
    if (found) {
      return found.view()
    }
    const mimeType = detect(data)
    const blob = db.blobs.create({
      cid: cid.toString(),
      mimeType,
      size: data.length,
      data: new Buffer(data.buffer)
    })
    await db.em.persist(blob).flush()
    return blob.view()
  }

  async get (cid, { db }) {
    return db.blobs.findOne({ cid }, { populate: ['data'] })
  }
}

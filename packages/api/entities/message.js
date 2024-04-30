import { EntitySchema, wrap } from '@mikro-orm/core'
import { encrypt, decrypt, PrivateKey } from 'eciesjs'
import { bytesToBase64 } from '../lib/utils.js'

export class Message {
  async view (ctx) {
    const m = wrap(this).toJSON()

    /* const key = JSON.parse(ctx.user.signingKey)
    //const sk = new PrivateKey()
    //const pk = sk.publicKey.toHex() //Uint8Array.from(key.publicKey)
    const pk = new Uint8Array(Object.values(key.publicKey))
    const sk = { secret: new Uint8Array(Object.values(key.privateKey)) }

    const enc = encrypt(pk, Buffer.from(m.msg))

    m.ciphertext = bytesToBase64(enc)
    //m.decrypted = decrypt(sk.secret, enc).toString()
*/
    // load author
    const authorQuery = await ctx.api.objectGet(ctx, m.authorDid)
    if (authorQuery) {
      const a = authorQuery.item
      m.author = {
        type: authorQuery.type,
        did: a.did,
        handle: a.handle,
        name: a.name,
        avatarBlob: a.avatarBlob,
        baseUrl: a.baseUrl
      }
    }
    return m
  }
}

export const schema = new EntitySchema({
  name: 'Message',
  extends: 'BaseEntity',
  class: Message,
  properties: {
    room: {
      type: 'string'
    },
    authorDid: {
      type: 'string'
    },
    msg: {
      type: 'string'
    }
  }
})

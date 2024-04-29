import { EntitySchema, wrap } from '@mikro-orm/core'

export class Message {
  async view (ctx) {
    const m = wrap(this).toJSON()
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

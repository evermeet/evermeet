import { EntitySchema, wrap } from '@mikro-orm/core'
import { ObjectId } from '../lib/db.js'

export const CalendarManager = new EntitySchema({
  name: 'CalendarManager',
  embeddable: true,
  properties: {
    ref: {
      type: 'string'
    },
    t: {
      type: 'string',
      nullable: true
    }
  }
})

export class Calendar {
  async view (opts = {}, ctx) {
    const json = wrap(this).toJSON()
    if (opts.events !== false) {
      json.events = []
      const events = await ctx.db.events.find({ calendarId: json._id })
      for (const e of events) {
        json.events.push(await e.view(Object.assign(opts, { calendar: false }), ctx))
      }
    }
    json.baseUrl = `/${json.handle?.replace('.' + ctx.api.config.domain, '') || json._id}`
    json.url = `https://${ctx.api.config.domain}${json.baseUrl}`
    json.handleUrl = json.handle
    return json
  }
}

export const schema = new EntitySchema({
  class: Calendar,
  name: 'Calendar',
  // extends: 'CustomBaseEntity',
  properties: {
    _id: {
      type: 'string',
      maxLength: 32,
      primary: true,
      onCreate: () => ObjectId()
    },
    name: {
      type: 'string'
    },
    handle: {
      type: 'string',
      format: 'handle',
      nullable: true,
      unique: true
    },
    visibility: {
      type: 'string',
      enum: ['public', 'unlisted', 'private'],
      nullable: true
    },
    did: {
      type: 'string',
      unique: true,
      nullable: true
    },
    slug: {
      type: 'string',
      nullable: true
    },
    personal: {
      type: 'boolean',
      nullable: true,
      onCreate: () => false
    },
    subs: {
      type: 'number',
      nullable: true
    },
    img: {
      type: 'string',
      nullable: true
    },
    backdropImg: {
      type: 'string',
      nullable: true
    },
    description: {
      type: 'string',
      nullable: true
    },
    managersArray: {
      type: 'array',
      nullable: true,
      onCreate: obj => (obj.managers || []).map(m => m.ref),
      onUpdate: obj => (obj.managers || []).map(m => m.ref)
    },
    managers: {
      kind: 'embedded',
      entity: 'CalendarManager',
      onCreate: () => [],
      array: true
    },
    signingKey: {
      type: 'string',
      nullable: true,
      lazy: true
    },
    rotationKey: {
      type: 'string',
      nullable: true,
      lazy: true
    }
  }
})

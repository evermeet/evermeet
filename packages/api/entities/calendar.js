import { EntitySchema, wrap } from '@mikro-orm/core'

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

export const CalendarConfigEntity = new EntitySchema({
  name: 'CalendarConfig',
  embeddable: true,
  properties: {
    name: {
      type: 'string'
    },
    slug: {
      type: 'string',
      nullable: true
    },
    subs: {
      type: 'number',
      nullable: true
    },
    avatarBlob: {
      type: 'string',
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
    headerBlob: {
      type: 'string',
      nullable: true
    },
    description: {
      type: 'string',
      nullable: true
    },
    refs: {
      type: 'object',
      nullable: true
    }
  }
})

export class Calendar {
  async view (ctx, opts = {}) {
    const c = wrap(this).toJSON()

    let events
    if (opts.events !== false) {
      events = []
      for (const e of await ctx.db.events.find({ calendarId: c.id })) {
        events.push(await e.view(ctx, Object.assign(opts, { calendar: this })))
      }
    }
    let concepts
    if (opts.concepts !== false || opts.concepts !== false) {
      if (ctx.cache?.calendar[this.did]?.concepts) {
        concepts = ctx.cache?.calendar[this.did]?.concepts
      } else {
        const conceptsArr = await ctx.db.concepts.find({ calendarId: c.id })
        if (conceptsArr) {
          concepts = await Promise.all(conceptsArr.map(i => i.view(ctx, Object.assign(opts, { calendar: this }))))
        }
      }
    }

    let rooms
    if (ctx.cache?.calendar[this.did]?.rooms) {
      rooms = ctx.cache?.calendar[this.did]?.rooms
    } else {
      const roomsQuery = await ctx.db.rooms.find({ repo: this.did })
      if (roomsQuery) {
        rooms = await Promise.all(roomsQuery.map(r => r.view(ctx)))
        if (!ctx.cache) {
          ctx.cache = {}
        }
        if (!ctx.cache.calendar) {
          ctx.cache.calendar = {}
        }
        ctx.cache.calendar[this.did] = rooms
      }
    }

    const baseUrl = `/${c.handle?.replace('.' + ctx.api.config.domain, '') || c.id}`
    const url = `https://${ctx.api.config.domain}${baseUrl}`
    const handleUrl = c.handle

    let userContext, managers
    if (ctx.user) {
      const isManager = c.managersArray.includes(ctx.user.did)
      userContext = {
        isManager
      }
      if (isManager) {
        managers = c.managers
      }
    }

    return {
      id: c.id,
      did: c.did,
      handle: c.handle,
      visibility: c.visibility,
      personal: c.personal,
      ...c.config,
      events,
      concepts,
      rooms,
      baseUrl,
      url,
      handleUrl,
      $userContext: userContext,
      managers
    }
  }
}

export const CalendarSchema = new EntitySchema({
  class: Calendar,
  name: 'Calendar',
  extends: 'BaseDidEntity',
  properties: {
    visibility: {
      type: 'string',
      enum: ['public', 'unlisted', 'private'],
      onCreate: (obj) => obj.visibility || 'public'
    },
    personal: {
      type: 'boolean',
      nullable: true,
      onCreate: () => false
    },
    config: {
      kind: 'embedded',
      entity: 'CalendarConfig',
      onCreate: () => new CalendarConfig()
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
    }
  }
})

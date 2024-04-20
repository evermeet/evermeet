import { EntitySchema, wrap } from '@mikro-orm/core'
import { ObjectId } from '../lib/db.js'

export const EventConfig = new EntitySchema({
  name: 'EventConfig',
  embeddable: true,
  properties: {
    slug: {
      type: 'string'
    },
    name: {
      type: 'string'
    },
    dateStart: {
      type: 'string',
      format: 'datetime'
    },
    dateEnd: {
      type: 'strnig',
      format: 'datetime'
    },
    img: {
      type: 'string',
      nullable: true
    },
    placeName: {
      type: 'string',
      nullable: true
    },
    placeCountry: {
      type: 'string',
      nullable: true
    },
    placeCity: {
      type: 'string',
      nullable: true
    },
    description: {
      type: 'string',
      nullable: true
    }
  }
})

export class Event {
  async view (opts = {}, ctx) {
    const json = {
      id: this._id,
      calendarId: this.calendarId,
      ...wrap(this.config).toJSON()
    }

    const calendar = await ctx.db.calendars.findOne({ _id: this.calendarId })
    json.calendar = await calendar.view({ events: false }, ctx)
  
    json.baseUrl = json.calendar.baseUrl + '/' + json.slug
    json.url = json.calendar.url + '/' + json.slug
    json.handleUrl = json.calendar.handle + '/' + json.slug
    // json.guestCountNative = (json.guestsNative || []).length
    // json.guestCountTotal = json.guestCountNative + (json.guestCount || 0)
    return json
  }
}

export const schema = new EntitySchema({
  name: 'Event',
  class: Event,
  properties: {
    _id: {
      type: 'string',
      maxLength: 32,
      primary: true,
      onCreate: () => ObjectId()
    },
    calendarId: {
      type: 'string'
    },
    slug: {
      type: 'string',
      unique: true,
      onCreate: obj => obj.config.slug,
      onUpdate: obj => obj.config.slug
    },
    config: {
      kind: 'embedded',
      entity: 'EventConfig'
    }
  }
})

/* dateStart: 2024-04-11T14:00
dateEnd: 2024-04-11T20:00
img: https://images.lumacdn.com/cdn-cgi/image/format=auto,fit=cover,dpr=2,quality=75,width=400,height=400/event-covers/7h/c460b46a-ff30-4192-a230-f51942673eab
placeName: Vrij Paleis
placeCountry: nl
placeCity: Amsterdam
guestCount: 54
guests:
  - name: jensei
  - name: Enrico
guestsNative:
  - ref: did:plc:524tuhdhh3m7li5gycdn6boe
    rel: joined
    time: 2024-01-01T01:00
hosts:
  - name: PD_CDG
    img: https://images.lumacdn.com/cdn-cgi/image/format=auto,fit=cover,dpr=2,background=white,quality=75,width=24,height=24/avatars/4w/53843e16-874b-4666-92a2-2341fb5de057
description: | */

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
      id: this.id,
      calendarId: this.calendarId,
      ...wrap(this.config).toJSON()
    }

    const calendar = await ctx.db.calendars.findOne({ id: this.calendarId })
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
  extends: 'BaseEntity',
  properties: {
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

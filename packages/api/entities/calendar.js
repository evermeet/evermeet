import { EntitySchema, wrap } from '@mikro-orm/core'
import { ObjectId } from '../lib/db.js'

export const CalendarManager = new EntitySchema({
  name: 'CalendarManager',
  embeddable: true,
  properties: {
    ref: {
      type: 'string'
    },
    time: {
        type: 'string',
        nullable: true,
    }
  }
})

export class Calendar {
  async view (opts = {}, { db = {} } = {}) {
    const json = wrap(this).toJSON()
    /*if (opts.events !== false) {
        json.events = []
        const events = await api.cols.events.find({ selector: { calendarId: json.id }}).exec();
        for (const e of events) {
            json.events.push(await e.view(Object.assign(opts, { calendar: false })))
        }
    }*/
    return json;
  }
}

export const schema = new EntitySchema({
  class: Calendar,
  name: 'Calendar',
  //extends: 'CustomBaseEntity',
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
    slug: {
        type: 'string'
    },  
    personal: {
        type: 'boolean',
        nullable: true,
        onCreate: () => false,
    },
    subs: {
      type: 'number',
      nullable: true,
    },
    img: {
      type: 'string',
      nullable: true,
    },
    backdropImg: {
      type: 'string',
      nullable: true,
    },
    description: {
      type: 'string',
      nullable: true,
    },
    managers: {
      kind: 'embedded',
      entity: 'CalendarManager',
      onCreate: () => [],
      array: true
    },
  }
});
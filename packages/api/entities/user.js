import { EntitySchema, wrap } from '@mikro-orm/core'
import { ObjectId } from '../lib/db.js'

class User {
  async view () {
    const json = wrap(this).toJSON()
    return {
      handle: json.handle,
      did: json.did,
      createdOn: json.createdOn
    }
  }
}

export const EventConfig = new EntitySchema({
  name: 'UserCalendarSubscription',
  embeddable: true,
  properties: {
    ref: {
      type: 'string'
    },
    t: {
      type: 'string',
      format: 'date-time'
    }
  }
})

export const schema = new EntitySchema({
  name: 'User',
  class: User,
  properties: {
    _id: {
      type: 'string',
      maxLength: 32,
      primary: true,
      onCreate: () => ObjectId()
    },
    handle: {
      type: 'string',
      unique: true
    },
    password: {
      type: 'string'
    },
    did: {
      type: 'string'
    },
    name: {
      type: 'string',
      nullable: true
    },
    email: {
      type: 'string',
      nullable: true
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
    },
    calendarSubscriptions: {
      kind: 'embedded',
      entity: 'UserCalendarSubscription',
      onCreate: () => [],
      array: true
    },
    createdOn: {
      type: 'string',
      format: 'date-time',
      onCreate: () => (new Date()).toISOString()
    }
  }
})

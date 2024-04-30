import { EntitySchema, wrap } from '@mikro-orm/core'
import { ObjectId } from '../lib/db.js'

export const UserDatePreferences = new EntitySchema({
  name: 'UserDatePreferences',
  embeddable: true,
  properties: {
    hoursFormat: {
      type: 'string',
      enum: [
        '24-hour',
        '12-hour'
      ],
      nullable: true
    }
  }
})

export const UserPreferences = new EntitySchema({
  name: 'UserPreferences',
  embeddable: true,
  properties: {
    date: {
      kind: 'embedded',
      entity: 'UserDatePreferences',
      onCreate: () => ({})
    }
  }
})

class User {
  async view (ctx, opts = {}) {
    const u = wrap(this).toJSON()
    return {
      id: u.id,
      did: u.did,
      handle: u.handle,
      name: u.name,
      description: u.description,
      avatarBlob: u.avatarBlob,
      preferences: u.preferences,
      calendarSubscriptions: u.calendarSubscriptions,
      createdOn: u.createdOn
    }
  }
}

export const UserCalendarSubscription = new EntitySchema({
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
  extends: 'BaseEntity',
  properties: {
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
    description: {
      type: 'string',
      nullable: true
    },
    preferences: {
      kind: 'embedded',
      entity: 'UserPreferences',
      onCreate: () => ({})
    },
    avatarBlob: {
      type: 'string',
      nullable: true
    },
    signingKey: {
      type: 'string',
      nullable: true
      // lazy: true
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

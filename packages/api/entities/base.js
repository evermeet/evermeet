import { EntitySchema, wrap } from '@mikro-orm/core'
import { createId } from '../lib/db.js'

export const BaseEntity = new EntitySchema({
  name: 'BaseEntity',
  properties: {
    id: {
      type: 'string',
      primary: true,
      onCreate: e => e.id || createId()
    },
    createdOn: {
      type: 'string',
      format: 'date-time',
      onCreate: () => (new Date()).toISOString()
    }
  }
})

export const BaseDidEntity = new EntitySchema({
  name: 'BaseDidEntity',
  extends: 'BaseEntity',
  properties: {
    did: {
      type: 'string',
      unique: true
    },
    handle: {
      type: 'string',
      format: 'handle',
      nullable: true,
      unique: true
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
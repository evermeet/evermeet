import { EntitySchema, wrap } from '@mikro-orm/core'

export class Room {
  async view (ctx) {
    const m = wrap(this).toJSON()
    return m
  }
}

export const schema = new EntitySchema({
  name: 'Room',
  class: Room,
  extends: 'BaseEntity',
  properties: {
    repo: {
      type: 'string'
    },
    slug: {
      type: 'string',
      nullable: true
    },
    name: {
      type: 'string',
      nullable: true
    }
  }
})

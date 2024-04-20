import { EntitySchema, wrap } from '@mikro-orm/core'
import { ObjectId } from '../lib/db.js'


export const schema = new EntitySchema({
    name: 'User',
    properties: {
        _id: {
            type: 'string',
            maxLength: 32,
            primary: true,
            onCreate: () => ObjectId()
        },
        username: {
            type: 'string',
            unique: true,
        },
        password: {
            type: 'string',
        },
        did: {
            type: 'string',
        },
        name: {
            type: 'string',
        },
        email: {
            type: 'string',
        },
    }
})
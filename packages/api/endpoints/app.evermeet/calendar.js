
import { createDid } from '../../lib/did.js'
import { createId } from '../../lib/db.js'

export function createCalendar (server, { api: { authVerifier } }) {
  server.endpoint({
    auth: authVerifier.accessUser,
    handler: async (ctx) => {
      const { input, db, user } = ctx
      // check if handle exists if its not private
      if (input.visibility !== 'private') {
        const handleFound = await ctx.api.objectGet(ctx, input.handle)
        if (handleFound) {
          return { error: 'HandleNotAvailable' }
        }
        const domain = input.handle.split('.').slice(1).join('.').toLowerCase()
        // check if its available domain on this instance
        if (!ctx.api.config.availableUserDomains.includes('.' + domain)) {
          return { error: 'UnsupportedDomain' }
        }
      } else if (input.handle) {
        return { error: 'PrivateCannotHaveHandle' }
      }

      // update blobs
      let avatarBlob, headerBlob
      if (input.avatarBlob) {
        const blob = await db.blobs.findOne({ cid: input.avatarBlob.$cid })
        if (!blob) {
          return { error: 'InvalidAvatarBlob' }
        }
        avatarBlob = blob.cid
      }
      if (input.headerBlob) {
        const blob = await db.blobs.findOne({ cid: input.headerBlob.$cid })
        if (!blob) {
          return { error: 'InvalidAvatarBlob' }
        }
        headerBlob = blob.cid
      }

      // setup initial managers
      const managers = [{ ref: user.did, t: new Date() }]

      // get Id
      const id = createId()

      // get DID
      const didData = await createDid(input.handle || id, ctx)

      // construct calendar
      const calendar = db.calendars.create({
        id,
        ...didData,
        handle: input.handle,
        visibility: input.visibility,
        config: {
          name: input.name,
          description: input.description,
          avatarBlob,
          headerBlob
        },
        managers
      })
      await db.em.persist(calendar).flush()

      return {
        body: await calendar.view(ctx)
      }
    }
  })
}

export function getUserCalendars (server, { api: { authVerifier } }) {
  server.endpoint({
    auth: authVerifier.accessUser,
    handler: async (ctx) => {
      // subscribed calendars
      const subs = ctx.user.calendarSubscriptions.map(cs => cs.ref)
      const localSubscribed = await ctx.db.calendars.find({ did: { $in: subs } })

      const subscribed = []
      await Promise.all(subs.map(async (did) => {
        let cal
        const local = localSubscribed.find(c => c.did === did)
        if (local) {
          cal = await local.view(ctx, {})
        } else {
          // implement remote fetching
        }
        subscribed.push(cal)
      }))

      // owned calendars
      const ownLocal = await ctx.db.calendars.find({ managersArray: { $in: [ctx.user.did] } })
      const owned = await Promise.all(ownLocal.map((c) => {
        return c.view(ctx, { events: false })
      }))

      return {
        body: {
          subscribed,
          owned
        }
      }
    }
  })
}

export function getConcepts (server) {
  server.endpoint(async (ctx) => {
    const { did } = ctx.input
    const calendar = await ctx.db.calendars.findOne({ did })
    const concepts = await ctx.db.concepts.find({ calendarId: calendar.id })
    return {
      body: { concepts: await Promise.all(concepts.map(i => i.view(ctx, { calendar }))) }
    }
  })
}

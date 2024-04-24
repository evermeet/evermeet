
import { createDid } from '../../lib/did.js'
import { createId } from '../../lib/db.js'

export function createCalendar (server, ctx) {
  server.endpoint({
    auth: ctx.api.authVerifier.accessUser,
    handler: async ({ input, db, user }) => {
      // check if handle exists if its not private
      if (input.visibility !== 'private') {
        const handleFound = await ctx.api.objectGet(db, input.handle)
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
      let avatarBlob
      if (input.avatarBlob) {
        console.log(input)
        const blob = await db.blobs.findOne({ cid: input.avatarBlob.$cid })
        if (!blob) {
          return { error: 'InvalidAvatarBlob' }
        }
        avatarBlob = blob.cid
      }

      // setup initial managers
      const managers = [{ ref: user.did, t: new Date() }]

      // get DID
      const didData = await createDid(input.handle || id, ctx)

      // construct calendar
      const calendar = db.calendars.create({
        ...input,
        ...didData,
        avatarBlob,
        managers
      })
      await db.em.persist(calendar).flush()

      return {
        body: await calendar.view({}, { db, api: ctx.api })
      }
    }
  })
}

export function getUserCalendars (server, ctx) {
  server.endpoint({
    auth: ctx.api.authVerifier.accessUser,
    handler: async ({ db, user }) => {
      // subscribed calendars
      const subs = user.calendarSubscriptions.map(cs => cs.ref)
      const localSubscribed = await db.calendars.find({ did: { $in: subs } })

      const subscribed = []
      await Promise.all(subs.map(async (did) => {
        let cal
        const local = localSubscribed.find(c => c.did === did)
        if (local) {
          cal = await local.view({}, { db, api: ctx.api })
        } else {
          // implement remote fetching
        }
        subscribed.push(cal)
      }))

      // owned calendars
      const ownLocal = await db.calendars.find({ managersArray: { $in: [user.did] } })
      const owned = await Promise.all(ownLocal.map((c) => {
        return c.view({ events: false }, { db, api: ctx.api })
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

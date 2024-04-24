
export function getExplore (server, ctx) {
  server.endpoint(async ({ db }) => {
    const calendars = await db.calendars.find({
      personal: { $ne: true },
      visibility: 'public'
    })
    const featured = []

    return {
      body: {
        calendars: await Promise.all(calendars.map(async (cal) => {
          return await cal.view({ events: false }, { db, api: ctx.api })
        })),
        featured
      }
    }
  })
}

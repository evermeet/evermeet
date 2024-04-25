
export function getExplore (server) {
  server.endpoint(async (ctx) => {
    const calendars = await ctx.db.calendars.find({
      personal: { $ne: true },
      visibility: 'public'
    })
    const featured = []

    return {
      body: {
        calendars: await Promise.all(calendars.map(async (cal) => {
          return await cal.view(ctx, { events: false })
        })),
        featured
      }
    }
  })
}

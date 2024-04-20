
export function getExplore (server, { api }) {
  server.endpoint(async ({ db }) => {
    const calendars = await db.calendars.find({ personal: { $ne: true } })
    const featured = []

    return {
      body: {
        calendars,
        featured
      }
    }
  })
}

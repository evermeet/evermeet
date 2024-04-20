
export function getProfile (server, ctx) {
  server.endpoint(async ({ input: { id }, db }) => {
    const obj = await ctx.api.objectGet(db, id)
    if (!obj) {
      return { error: 'NotFound' }
    }
    return {
      encoding: 'application/json',
      body: obj
    }
  })
}

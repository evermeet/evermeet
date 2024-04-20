export function getCollection (server, ctx) {
  server.endpoint({
    auth: ctx.api.authVerifier.accessAdmin,
    handler: async ({ input: { col } }) => {
      const items = await ctx.api.cols[col].find({}).exec()
      return {
        encoding: 'application/json',
        body: {
          items
        }
      }
    }
  })
}

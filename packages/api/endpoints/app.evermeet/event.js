
export function importEvents (server, { api: { authVerifier } }) {
  server.endpoint({
    auth: authVerifier.accessUser,
    handler: async (ctx) => {
      const { url } = ctx.input

      const connector = ctx.api.findConnector(url)
      if (!connector) {
        return { error: 'ConnectorNotFound' }
      }

      const inspect = await connector.inspect(ctx, url)
      return {
        body: {
          url,
          connector: connector?.id,
          inspect
        }
      }
    }
  })
}

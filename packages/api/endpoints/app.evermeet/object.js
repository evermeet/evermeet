
export function getProfile (server) {
  server.endpoint(async (ctx) => {
    const obj = await ctx.api.objectGet(ctx, ctx.input.id)
    if (!obj) {
      return { error: 'NotFound' }
    }
    return {
      encoding: 'application/json',
      body: obj
    }
  })
}

export async function getBlob (server, ctx) {
  server.endpoint(async ({ input: { cid }, db }) => {
    const blob = await ctx.api.blobStore.get(cid, { db })
    if (!blob) {
      return {
        error: 'NotFound'
      }
    }
    return blob.toResponse()
  })
}

export async function uploadBlob (server, ctx) {
  server.endpoint({
    auth: ctx.api.authVerifier.accessUser,
    handler: async ({ db, user, input, encoding }) => {
      let blob
      try {
        blob = await ctx.api.blobStore.add(input, { db })
      } catch (e) {
        return {
          error: e.message
        }
      }
      return {
        encoding: 'application/json',
        body: {
          blob
        }
      }
    }
  })
}

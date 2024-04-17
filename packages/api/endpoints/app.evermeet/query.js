
export function resolveObject (server, ctx) {
    server.endpoint(async ({ input: { id } }) => {
        const obj = await ctx.api.objectGet(id)
        if (!obj) {
            return { error: 'NotFound' }
        }
        return {
            encoding: 'application/json',
            body: obj
        }
    })
}
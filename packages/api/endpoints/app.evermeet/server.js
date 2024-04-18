
export function describeServer (server, ctx) {
    server.endpoint(async ({ user }) => {
        const runtime = ctx.api.runtime
        return {
            encoding: 'application/json',
            body: {
                did: `did:web:${ctx.api.config.domain}`,
                app: ctx.api.pkg.name,
                version: ctx.api.pkg.version,
                env: {
                    adapter: ctx.api.config.api.adapter,
                    runtime: runtime.name,
                    runtimeVersion: runtime.version,
                    arch: runtime.arch,
                },
            }
        }
    })
}
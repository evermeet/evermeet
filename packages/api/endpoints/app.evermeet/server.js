import os from 'node:os' 

export function describeServer (server, ctx) {
    server.endpoint(async ({ user }) => {
        return {
            encoding: 'application/json',
            body: {
                did: `did:web:${ctx.api.config.domain}`,
                app: ctx.api.pkg.name,
                version: ctx.api.pkg.version,
                env: {
                    adapter: ctx.api.config.api.adapter,
                    runtime: ctx.api.config.api.runtime,
                    runtimeVersion: ctx.api.config.api.runtime == 'bun' ? Bun.version : process.version,
                    arch: os.arch(),
                },
            }
        }
    })
}
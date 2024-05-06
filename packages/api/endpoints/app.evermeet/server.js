export function describeServer(server, { api }) {
  server.endpoint(async ({ user }) => {
    const runtime = api.runtime;
    return {
      encoding: "application/json",
      body: {
        did: `did:web:${api.config.domain}`,
        app: api.pkg.name,
        version: api.pkg.version,
        env: {
          adapter: api.config.api.adapter,
          runtime: runtime.name,
          runtimeVersion: runtime.version,
          arch: runtime.arch,
        },
      },
    };
  });
}

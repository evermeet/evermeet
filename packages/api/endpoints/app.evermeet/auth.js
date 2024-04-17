
export function createSession (server, ctx) {
    server.endpoint({
        rateLimit: {},
        handler: async ({ input, req }) => {

            return {
                encoding: 'application/json',
                body: {
                    test: 42
                }
            }
        }
    })
}
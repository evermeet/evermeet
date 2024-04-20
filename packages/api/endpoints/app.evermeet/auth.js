import argon2 from 'argon2'

export function createSession (server, ctx) {
    server.endpoint(async ({ input, db }) => {
        if (!input) {
            return { error: 'InputNotSpecified' }
        }
        let ident = input.identifier
        const password = input.password
        
        if (!ident || !password) {
            return { error: 'InputNotSpecified' }
        }
        // normalize input
        ident = ident.toLowerCase()

        // find user
        let user;
        if (ident.match(/@/)) {
            user = await db.users.findOne({ email: ident })
        } else {
            user = await db.users.findOne({ username: ident })
        }
        if (!user) {
            return { error: 'BadCredentials' }
        }

        // verify password
        const verified = await argon2.verify(user.password, password)
        if (!verified) {
            return { error: 'BadCredentials' }
        }

        return {
            encoding: 'application/json',
            body: {
                accessJwt: 'xxx',
                //refreshJwt: 'xxx',
                username: user.username,
                did: user.did
            }
        }
    })
}
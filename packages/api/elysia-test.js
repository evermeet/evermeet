import { Elysia, t } from 'elysia'
import { swagger } from '@elysiajs/swagger'

import pkg from '../../package.json' with {type: "json"};

new Elysia()
    .use(swagger())
    .get(
        '/api',
        () => {
            return {
                app: pkg.name,
                version: pkg.version
            }
        }
    )
    .listen(3002)
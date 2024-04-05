import { sha256 } from './crypto.js';
import { v4 as uuidv4 } from 'uuid';

export function makeRoutes(app, api) {
    const db = api.db;

    app.get('/api', function (request, reply) {
        reply.send({
            app: api.pkg.name,
            version: api.pkg.version
        })
    })

    app.get('/api/query/:id', async (req, reply) => {
        const { id } = req.params;

        const calendar = await api.cols.calendars.findOne({ selector: { slug: id } }).exec()
        if (calendar) {
            return reply.send({
                type: 'calendar',
                item: await calendar.view()
            })
        }

        const event = await api.cols.events.findOne({ selector: { slug: id } }).exec()
        if (event) {
            return reply.send({
                type: 'event',
                item: await event.view()
            })
        }

        return reply.code(404).send({ error: 'not-found' })
    })

    app.get('/api/event/:id', async (req, reply) => {
        const item = await db.events.findOne(req.params.id).exec()
        if (!item) {
            return reply.code(404).send({ error: 'not-found' })
        }
        reply.send(await item.view())
    })

    app.get('/api/event/:id/:subview', async (req, reply) => {
        const item = await api.cols.events.findOne(req.params.id).exec()
        if (!item) {
            return reply.code(404).send({ error: 'not-found' })
        }
        reply.send(await item['view_'+req.params.subview]())
    })

    app.get('/api/events', async (req, reply) => {
        const cursor = await api.cols.events.find({}).exec()
        reply.send({ events: cursor })
    })

    app.get('/api/explore', async (req, reply) => {
        const cursor = await api.cols.calendars.find({
            selector: {
                personal: { $ne: true }
            }
        }).exec()
        reply.send({ calendars: cursor })
    })

    app.get('/api/calendars', async (req, reply) => {
        const cursor = await api.cols.calendars.find({}).exec()
        reply.send(cursor)
    })

    app.get('/api/calendar/:id', async (req, reply) => {
        const { id } = req.params;
        return db.calendars.findOne({ id }).exec()
    })

    app.get('/api/admin', async (req, reply) => {
        const out = {
            collections: []
        }
        for (const colName of Object.keys(api.cols)) {
            out.collections.push({
                name: colName,
                size: await api.cols[colName].count().exec()
            })
        } 
        reply.send(out)
    })

    app.post('/api/login', async (req, reply) => {
        const { email } = req.body

        const hash = sha256(email)
        const user = await api.cols.users.findOne({ selector: { emailShas: { $in: [hash] }}}).exec()

        if (!user) {
            return reply.code(401).send({ error: 'unauthorized' })
        }

        const session = {
            id: uuidv4(),
            user: user.id,
            time: new Date(),
        }
        await api.cols.sessions.insert(session)

        reply.send({
            ok: true,
            sessionId: session.id,
            user
        })
    })

    app.get('/api/me', async (req, reply) => {
        const sessionId = req.headers['evermeet-session-id'];
        if (!sessionId) {
            return reply.code(401).send({ error: 'no sessionId ("evermeet-session-id" header)'})
        }
        const session = await api.cols.sessions.findOne({ selector: { id: sessionId }}).exec()
        if (!session) {
           return reply.code(401).send({ error: 'session not found'}) 
        }
        const user = await api.cols.users.findOne(session.user).exec()
        if (!user) {
            return reply.code(401).send({ error: 'user not found '})
        }
        
        reply.send({
            user
        })
    })

}
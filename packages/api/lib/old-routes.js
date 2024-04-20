import { sha256 } from './crypto.js';
import { v4 as uuidv4 } from 'uuid';

export function makeRoutes(app, api) {
    const db = api.db;

    app.addHook('onRequest', async (req, reply) => {
        let authorized, user, error;
        try {
            const cookie = api.adapterCtl.getCookie(req, api.config.api.sessionName)
            const output = await api.authorizeSession(cookie)
            req.authorized = output[0]
            req.user = output[1]
            req.error = output[2]
        } catch (e) {
            console.error(e, error)
        }
    })

    app.get('/api', function (request, reply) {
        reply.send({
            app: api.pkg.name,
            version: api.pkg.version
        })
    })

    app.get('/api/config', async (req, reply) => {
        
        reply.send({
            domain: api.config.domain,
            sitename: api.config.sitename,
            options: api.config.options,
        })
    })

    app.get('/api/query/:id', async (req, reply) => {
        const { id } = req.params;

        const res = await api.objectGet(id)
        if (res) {
            return reply.send(res)
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
        if (!req.user) {
            return reply.code(401).send({ error: req.authError })
        }
        const user = req.user
        const events = []
        if (!user.events) {
            return reply.send({ events })
        }
        for (const ev of user.events) {
            const res = await api.objectGet(ev.ref, { events: false })
            events.push(res.item)
        }
        reply.send({ events })
    })

    app.get('/api/explore', async (req, reply) => {
        const calendarsQuery = await api.cols.calendars.find({
            selector: {
                personal: { $ne: true }
            }
        }).exec()
        const calendars = []
        if (calendarsQuery) {
            for (const ci of calendarsQuery) {
                calendars.push(await ci.view({ events: false }))
            }
        }

        const featured = []
        for (const fc of api.config.featuredCalendars) {
            const found = await api.objectGet(fc);
            if (found) {
                featured.push(await found.item)
            }
        }

        reply.send({ calendars, featured })
    })

    app.get('/api/calendars', async (req, reply) => {
        if (!req.user) {
            return reply.code(401).send({ error: req.authError })
        }
        const user = req.user

        const subscribed = []
        if (user.subscribedCalendars) {
            for (const c of user.subscribedCalendars) {
                const obj = await api.objectGet(c.ref)
                if (obj) {
                    subscribed.push(obj.item)
                }
            }
        }
        const owned = []
        if (user.personalCalendar) {
            const personal = await api.objectGet(user.personalCalendar);
            if (personal) {
                owned.push(personal.item)
            }
        }
        if (user.calendarsManage) {
            for (const mi of user.calendarsManage) {
                const obj = await api.objectGet(mi.ref)
                owned.push(obj.item)
            }
        }

        reply.send({ subscribed, owned })
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
    app.get('/api/admin/collection/:collectionName', async (req, reply) => {
        const items = await api.cols[req.params.collectionName].find({}).exec()
        reply.send({ items })
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

        reply
            .setCookie('evermeet-session-id', session.id, {
                path: '/',
                httpOnly: true,
                maxAge: 31_556_926,
                secure: true,
                sameSite: 'Lax'
            })
            .send({
                ok: true,
                //sessionId: session.id,
                user
            })
    })

    app.get('/api/me', async (req, reply) => {
        if (!req.user) {
            return reply.code(401).send({ error: req.authError })
        }

        reply.send({
            user: req.user
        })
    })

    app.post('/api/me/calendarSubscribe', async (req, reply) => {
        if (!req.user) {
            return reply.code(401).send({ error: req.authError })
        }
        const user = req.user
        // TODO check if calendar exists
        const calendarId = req.body.id
        if (!calendarId) {
            return reply.code(401).send()
        }
        const calendar = await api.objectGet(calendarId)
        if (!calendar) {
            return reply.code(401).send({ error: 'calendar not exists' })
        }
        
        const subscribedCalendars = (user.subscribedCalendars || []).concat([ { ref: calendarId, time: (new Date()).toISOString() } ])
        await user.update({
            $set: {
                subscribedCalendars
            }
        })

        reply.send({
            done: true,
            subscribedCalendars
        })
    })

    app.post('/api/me/calendarUnsubscribe', async (req, reply) => {
        if (!req.user) {
            return reply.code(401).send({ error: req.authError })
        }
        const user = req.user

        const calendarId = req.body.id
        if (!calendarId) {
            return reply.code(401).send()
        }

        let subscribedCalendars = []
        for (const sc of user.subscribedCalendars) {
            if (sc.ref !== calendarId) {
                subscribedCalendars.push(sc)
            }
        }        
        await user.update({
            $set: {
                subscribedCalendars
            }
        })

        reply.send({
            done: true,
            subscribedCalendars
        })
    })

    app.post('/api/me/register', async (req, reply) => {
        if (!req.user) {
            return reply.code(401).send({ error: req.authError })
        }
        if (!req.body.id) {
            return reply.code(400).send({ error: 'wrong argument' })
        }
        const user = req.user
        const item = await api.cols.events.findOne({ selector: { id: req.body.id }}).exec()
        if (!item) {
            return reply.code(400).send({ error: 'wrong argument'})
        }
        if (user.events && user.events.find(e => e.ref === item.id)) {
            return reply.code(400).send({ error: 'already registered' })
        }

        const time = (new Date()).toISOString()
        const outEvent = await item.update({
            $push: {
                guestsNative: { ref: user.did, rel: 'joined', time }
            }
        })
        const out = await user.update({
            $push: {
                events: { ref: item.id, rel: 'joined', time }
            }
        })
              
        return reply.send({
            done: true,
            event: await outEvent.view({ calendar: false }),
            events: out.events
        })
    })

    app.post('/api/me/unregister', async (req, reply) => {
        if (!req.user) {
            return reply.code(401).send({ error: req.authError })
        }
        if (!req.body.id) {
            return reply.code(400).send({ error: 'wrong argument' })
        }
        const user = req.user
        const item = await api.cols.events.findOne({ selector: { id: req.body.id }}).exec()
        if (!item) {
            return reply.code(400).send({ error: 'wrong argument'})
        }

        let outEvent;
        const eventRegisterItem = item.guestsNative && item.guestsNative.find(e => e.ref === user.did)
        if (eventRegisterItem) {
            outEvent = await item.update({
                $pull: {
                    guestsNative: eventRegisterItem
                }
            })
        }

        const registerItem = user.events && user.events.find(e => e.ref === item.id)
        if (!user.events || !registerItem) {
            return reply.code(400).send({ error: 'not registered' })
        }
        const out = await user.update({
            $pull: {
                events: registerItem
            }
        })
        return reply.send({
            done: true,
            event: await outEvent?.view({ calendar: false, events: false }),
            events: out.events
        })
    })

}
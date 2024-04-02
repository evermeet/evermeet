import { createFilesystemAdapter, Collection } from 'signaldb'
import { parse } from 'yaml'
import fs from 'node:fs'


async function loadMock(fn) {
    const res = fs.readFileSync('./mock-data/' + fn + '.yaml', 'utf-8')
    return parse(res)
}

export async function initDatabase() {

    const calendars = new Collection({
        //persistence: createFilesystemAdapter('./data/calendars.json'),
    })
    const events = new Collection({
        //persistence: createFilesystemAdapter('./data/events.json'),
    })

    // setup mock data
    const mockCalendars = await loadMock('calendars')
    const mockEvents = await loadMock('events')
    //console.log({ mockCalendars, mockEvents })

    mockCalendars.map((x) => calendars.insert(x))
    mockEvents.map((x) => events.insert(x))
    
    console.log('database ready')

    return {
        calendars,
        events
    };
}

import { Connector, fetchNextPage } from '../lib/connector'

function processEvent (e) {
  return {
    remoteId: e.id,
    name: e.title,
    dateStart: e.dateTime,
    dateEnd: e.endTime,
    placeName: e.venue?.name,
    placeCity: e.group?.city || e.venue?.city,
    placeCountry: e.venue?.country,
    timezone: e.timezone,
    description: e.description,
    img: e.featuredEventPhoto?.source,
    hosts: e.host ? [{ name: e.host.name }] : [],
    url: e.eventUrl,
    _obj: e
  }
}

async function inspect (ctx, url) {
  const pg = await fetchNextPage(url)

  return {
    events: [processEvent(pg.event)]
  }
}

export default new Connector({
  urlPatterns: [
    /^https:\/\/(www\.|)meetup\.com\/([^\/]+)\/events\/([\d]+)\//
  ],
  inspect
})

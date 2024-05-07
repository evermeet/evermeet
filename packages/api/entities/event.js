import { EntitySchema, wrap } from "@mikro-orm/core";
import { ObjectId } from "../lib/db.js";
import { URI } from "yaml-language-server";

export const EventConfig = new EntitySchema({
  name: "EventConfig",
  embeddable: true,
  properties: {
    slug: {
      type: "string",
      nullable: true,
    },
    name: {
      type: "string",
    },
    mode: {
      type: "string",
      nullable: true,
      enum: ["offline", "online", "mixed"],
    },
    joinUrl: {
      type: "string",
      format: "url",
      nullable: true,
    },
    dateStart: {
      type: "date",
      format: "datetime",
    },
    dateEnd: {
      type: "date",
      format: "datetime",
    },
    timezone: {
      type: "string",
      nullable: true,
    },
    img: {
      type: "string",
      nullable: true,
    },
    placeName: {
      type: "string",
      nullable: true,
    },
    placeCountry: {
      type: "string",
      nullable: true,
    },
    placeCity: {
      type: "string",
      nullable: true,
    },
    placeRestrictedToGuests: {
      type: "boolean",
      nullable: true,
    },
    description: {
      type: "string",
      nullable: true,
    },
    people: {
      kind: "embedded",
      entity: "EventPeople",
      array: true,
      onCreate: () => [],
    },
    peopleLists: {
      kind: "embedded",
      entity: "EventPeopleList",
      array: true,
      onCreate: () => [],
    },
    hosts: {
      kind: "embedded",
      entity: "EventHost",
      array: true,
      onCreate: () => [],
    },
  },
});

export const EventHost = new EntitySchema({
  name: "EventHost",
  embeddable: true,
  properties: {
    name: {
      type: "string",
    },
    img: {
      type: "string",
      nullable: true,
    },
  },
});

export const EventPeople = new EntitySchema({
  name: "EventPeople",
  embeddable: true,
  properties: {
    id: {
      type: "string",
    },
    name: {
      type: "string",
    },
    img: {
      type: "string",
      nullable: true,
    },
    caption: {
      type: "string",
      nullable: true,
    },
  },
});

export const EventPeopleList = new EntitySchema({
  name: "EventPeopleList",
  embeddable: true,
  properties: {
    id: {
      type: "string",
    },
    name: {
      type: "string",
    },
    people: {
      type: "array",
    },
  },
});

export class Event {
  async view(ctx, opts = {}) {
    const json = {
      id: this.id,
      calendarDid: this.calendarDid,
      did: this.did,
      ...wrap(this.config).toJSON(),
    };

    if (!json.mode) {
      json.mode = json.placeCountry ? "offline" : "online";
    }

    const calendar =
      typeof opts.calendar === "object"
        ? opts.calendar
        : await ctx.db.calendars.findOne({ did: this.calendarDid });

    if (typeof opts.calendar !== "object") {
      json.calendar = await calendar.view(ctx, { events: false });
    }

    json.handleUrl = calendar.handle + ":" + (json.slug || json.id);
    json.baseUrl = "/" + json.handleUrl;

    // json.guestCountNative = (json.guestsNative || []).length
    // json.guestCountTotal = json.guestCountNative + (json.guestCount || 0)
    return json;
  }
}

export const EventEntity = new EntitySchema({
  name: "Event",
  class: Event,
  extends: "BaseEntity",
  properties: {
    calendarDid: {
      type: "string",
    },
    slug: {
      type: "string",
      onCreate: (obj) => obj.config.slug,
      onUpdate: (obj) => obj.config.slug,
      nullable: true,
    },
    config: {
      kind: "embedded",
      entity: "EventConfig",
    },
    did: {
      type: "string",
      unique: true,
      nullable: true,
    },
    handle: {
      type: "string",
      format: "handle",
      nullable: true,
      unique: true,
    },
  },
});

class Concept extends Event {}

export const ConceptSchema = new EntitySchema({
  class: Concept,
  name: "Concept",
  extends: "Event",
});

import { EntitySchema, wrap } from "@mikro-orm/core";
import { sortBy } from "lodash";

export const CalendarManager = new EntitySchema({
  name: "CalendarManager",
  embeddable: true,
  properties: {
    ref: {
      type: "string",
    },
    t: {
      type: "string",
      nullable: true,
    },
  },
});

export const CalendarConfigEntity = new EntitySchema({
  name: "CalendarConfig",
  embeddable: true,
  properties: {
    name: {
      type: "string",
    },
    slug: {
      type: "string",
      nullable: true,
    },
    subs: {
      type: "number",
      nullable: true,
    },
    avatarBlob: {
      type: "string",
      nullable: true,
    },
    img: {
      type: "string",
      nullable: true,
    },
    backdropImg: {
      type: "string",
      nullable: true,
    },
    headerBlob: {
      type: "string",
      nullable: true,
    },
    description: {
      type: "string",
      nullable: true,
    },
    refs: {
      type: "object",
      nullable: true,
    },
  },
});

export class Calendar {
  async view(ctx, opts = {}) {
    const c = wrap(this).toJSON();

    let parent;
    if (typeof opts.parent === "object") {
      //parent = await opts.parent.view(ctx, Object.assign(opts, { childrens: false, parent: false }))
    } else if (opts.parent !== false && c.parent) {
      const pc = await ctx.db.calendars.findOne({ did: c.parent });
      if (pc) {
        parent = await pc.view(ctx, Object.assign(opts, { childrens: false }));
      }
    }

    let childrens;
    if (opts.childrens !== false) {
      childrens = [];
      for (const e of await ctx.db.calendars.find({ parent: c.did })) {
        childrens.push(
          await e.view(ctx, Object.assign(opts, { parent: this })),
        );
      }
    }

    const cals = [c.did];
    if (childrens?.length > 0) {
      cals.push(childrens.map((c) => c.did));
    }

    let events;
    if (opts.events !== false) {
      events = [];
      for (const e of await ctx.db.events.find(
        {
          calendarDid: { $in: cals },
          config: { dateEnd: { $gt: new Date().toISOString() } },
        },
        { orderBy: { config: { dateStart: 1 } } },
      )) {
        events.push(
          await e.view(
            ctx,
            Object.assign(opts, {
              calendar: childrens?.find((c) => c.did === e.calendarDid) || this,
            }),
          ),
        );
      }
    }
    let pastEvents;
    if (opts.pastEvents !== false) {
      pastEvents = [];
      for (const e of await ctx.db.events.find(
        {
          calendarDid: { $in: cals },
          config: { dateEnd: { $lt: new Date().toISOString() } },
        },
        { orderBy: { config: { dateEnd: -1 } } },
      )) {
        pastEvents.push(
          await e.view(ctx, Object.assign(opts, { calendar: this })),
        );
      }
    }

    let concepts;
    if (opts.concepts !== false || opts.concepts !== false) {
      if (ctx.cache?.calendar[this.did]?.concepts) {
        concepts = ctx.cache?.calendar[this.did]?.concepts;
      } else {
        const conceptsArr = await ctx.db.concepts.find({ calendarDid: c.did });
        if (conceptsArr) {
          concepts = await Promise.all(
            conceptsArr.map((i) =>
              i.view(ctx, Object.assign(opts, { calendar: this })),
            ),
          );
        }
      }
    }

    let rooms;
    if (ctx.cache?.calendar[this.did]?.rooms) {
      rooms = ctx.cache?.calendar[this.did]?.rooms;
    } else {
      const roomsQuery = await ctx.db.rooms.find({ repo: this.did });
      if (roomsQuery) {
        rooms = await Promise.all(roomsQuery.map((r) => r.view(ctx)));
        if (!ctx.cache) {
          ctx.cache = {};
        }
        if (!ctx.cache.calendar) {
          ctx.cache.calendar = {};
        }
        ctx.cache.calendar[this.did] = rooms;
      }
    }

    const baseUrl = `/${c.handle?.replace("." + ctx.api.config.domain, "") || c.id}`;
    const url = `https://${ctx.api.config.domain}${baseUrl}`;
    const handleUrl = c.handle;

    let userContext, managers;
    if (ctx.user) {
      const isManager = c.managersArray.includes(ctx.user.did);
      userContext = {
        isManager,
      };
      if (isManager) {
        managers = c.managers;
      }
    }

    return {
      id: c.id,
      did: c.did,
      handle: c.handle,
      visibility: c.visibility,
      personal: c.personal,
      ...c.config,
      parent,
      childrens,
      events,
      pastEvents,
      concepts,
      rooms,
      baseUrl,
      url,
      handleUrl,
      $userContext: userContext,
      managers,
    };
  }
}

export const CalendarSchema = new EntitySchema({
  class: Calendar,
  name: "Calendar",
  extends: "BaseDidEntity",
  properties: {
    visibility: {
      type: "string",
      enum: ["public", "unlisted", "private"],
      onCreate: (obj) => obj.visibility || "public",
    },
    personal: {
      type: "boolean",
      nullable: true,
      onCreate: () => false,
    },
    parent: {
      type: "string",
      nullable: true,
    },
    config: {
      kind: "embedded",
      entity: "CalendarConfig",
      onCreate: () => new CalendarConfig(),
    },
    managersArray: {
      type: "array",
      nullable: true,
      onCreate: (obj) => (obj.managers || []).map((m) => m.ref),
      onUpdate: (obj) => (obj.managers || []).map((m) => m.ref),
    },
    managers: {
      kind: "embedded",
      entity: "CalendarManager",
      onCreate: () => [],
      array: true,
    },
  },
});

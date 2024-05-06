export function getExplore(server) {
  server.endpoint(async (ctx) => {
    const calendars = await ctx.db.calendars.find({
      personal: { $ne: true },
      visibility: "public",
    });
    const featured = [];

    if (!ctx.cache) {
      ctx.cache = {};
    }
    if (!ctx.cache.calendar) {
      ctx.cache.calendar = {};
    }

    const rooms = await ctx.db.rooms.find({
      repo: { $in: calendars.map((c) => c.did) },
    });
    const concepts = await ctx.db.concepts.find({
      calendarDid: { $in: calendars.map((c) => c.did) },
    });
    for (const c of calendars) {
      ctx.cache.calendar[c.did] = {
        rooms: rooms.filter((r) => r.repo === c.did),
        concepts: concepts.filter((_c) => _c.calendarDid === c.did),
      };
    }

    return {
      body: {
        calendars: await Promise.all(
          calendars.map(async (cal) => {
            return await cal.view(ctx, { events: false });
          }),
        ),
        featured,
      },
    };
  });
}

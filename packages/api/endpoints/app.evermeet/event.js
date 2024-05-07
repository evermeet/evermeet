export function importEvents(server, { api: { authVerifier } }) {
  server.endpoint({
    auth: authVerifier.accessUser,
    handler: async (ctx) => {
      const { url } = ctx.input;

      const connector = ctx.api.findConnector(url);
      if (!connector) {
        return { error: "ConnectorNotFound" };
      }

      const inspect = await connector.inspect(ctx, url);
      return {
        body: {
          url,
          connector: connector?.id,
          inspect,
        },
      };
    },
  });
}

async function toggleWatchEvent(ctx, type = "add") {
  const { db, api } = ctx;

  const obj = await api.objectGet(ctx, ctx.input.eventHandle);
  const event = obj.type === "event" ? obj.item : null;

  if (!event) {
    return { error: "EventNotFound" };
  }

  const watchBase = {
    calendarDid: event.calendarDid,
    eventId: event.id,
    authorDid: ctx.user.did,
  };

  let watch;
  const found = await db.watches.findOne(watchBase);

  if (type === "add") {
    if (found) {
      watch = found;
    } else {
      watch = db.watches.create(watchBase);
      await db.em.persist(watch).flush();
    }
  }
  if (type === "remove") {
    if (!found) {
      return { error: "WatchNotFound" };
    }

    await db.em.remove(found).flush();
  }

  const eventPreview = await db.events.findOne({ id: event.id });
  return {
    body: {
      watch: watch && db.wrap(watch).toJSON(),
      event: await eventPreview.view(ctx, {
        calendar: false,
        concepts: false,
        pastEvents: false,
      }),
    },
  };
}

export function watchEvent(server, { api: { authVerifier } }) {
  server.endpoint({
    auth: authVerifier.accessUser,
    handler: async (ctx) => toggleWatchEvent(ctx, "add"),
  });
}

export function unwatchEvent(server, { api: { authVerifier } }) {
  server.endpoint({
    auth: authVerifier.accessUser,
    handler: async (ctx) => toggleWatchEvent(ctx, "remove"),
  });
}

export function getUserEvents(server, { api: { authVerifier } }) {
  server.endpoint({
    auth: authVerifier.accessUser,
    handler: async (ctx) => {
      const { db, user } = ctx;
      const watched = await db.watches.find({ authorDid: user.did });
      const events = await db.events.find(
        { id: { $in: watched.map((w) => w.eventId) } },
        { orderBy: { config: { dateStart: 1 } } },
      );

      return {
        body: await Promise.all(events.map((e) => e.view(ctx))),
      };
    },
  });
}

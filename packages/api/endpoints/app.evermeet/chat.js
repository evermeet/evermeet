export function listRooms(server) {
  server.endpoint(async (ctx) => {
    const repo = ctx.input.repo || ctx.user?.did;
    if (!repo) {
      return { error: "Need specify repo or be logged" };
    }

    const msgs = await ctx.db.rooms.find();

    if (!msgs) {
      return { error: "NotFound" };
    }
    return {
      encoding: "application/json",
      body: await Promise.all(msgs.map((m) => m.view(ctx))),
    };
  });
}

export function getMessages(server) {
  server.endpoint(async (ctx) => {
    const { repo, room } = ctx.input;

    const msgs = await ctx.db.messages.find(
      { room },
      { orderBy: [{ createdOn: 1 }] },
    );

    if (!msgs) {
      return { error: "NotFound" };
    }
    return {
      encoding: "application/json",
      body: await Promise.all(msgs.map((m) => m.view(ctx))),
    };
  });
}

export async function createMessage(server, { api: { authVerifier } }) {
  server.endpoint({
    auth: authVerifier.accessUser,
    handler: async (ctx) => {
      const {
        input: { room, msg },
        api,
        db,
        user,
      } = ctx;

      const roomObj = await db.rooms.findOne({ id: room });
      if (!roomObj) {
        return { error: "RoomNotFound" };
      }

      const message = await db.messages.create({
        room,
        authorDid: user.did,
        msg,
      });
      await db.em.persist(message).flush();

      return {
        encoding: "application/json",
        body: await message.view(ctx),
      };
    },
  });
}

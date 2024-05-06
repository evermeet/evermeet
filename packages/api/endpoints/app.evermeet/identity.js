export function resolveHandle(server) {
  server.endpoint(async (ctx) => {
    const { handle } = ctx.input;

    let did;
    const services = [];
    const local = await ctx.api.objectGet(ctx, handle);
    if (local && local.item.did) {
      did = local.item.did;
      services.push("evermeet");
    } else {
      const bsky = await fetch(
        `https://bsky.social/xrpc/com.atproto.identity.resolveHandle?handle=${handle}`,
      );
      if (bsky) {
        const bskyData = await bsky.json();
        did = bskyData.did;
        services.push("bluesky");
      }
    }
    if (!did) {
      return { error: "HandleNotFound" };
    }
    return {
      body: {
        did,
        services,
      },
    };
  });
}

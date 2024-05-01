import { xrpcCall } from "../../lib/api.js";

export async function load({ params, fetch, parent, url }) {
  const data = await parent();
  const rooms = await xrpcCall(
    { fetch, user: data.user },
    "app.evermeet.chat.listRooms",
  );

  let room = rooms ? rooms[0] : null;
  const roomQuery = url.searchParams.get("room");
  if (roomQuery) {
    const [slug, repo] = roomQuery.split("@");
    room = rooms.find((r) => r.repo === repo && r.slug === slug);
  }
  const messages = (
    await xrpcCall(
      { fetch, user: data.user },
      "app.evermeet.chat.getMessages",
      { repo: room.repo, room: room.id },
    )
  )?.reverse();

  return {
    rooms,
    room,
    messages,
  };
}

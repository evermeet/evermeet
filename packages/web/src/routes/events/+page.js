import { xrpcCall } from "$lib/api";

export async function load({ params, fetch, parent }) {
  const data = await parent();

  return {
    events: await xrpcCall(
      { fetch, user: data.user },
      "app.evermeet.event.getUserEvents",
    ),
  };
}

import { xrpcCall } from "../lib/api.js";

export async function load({ params, fetch, parent }) {
  const { user } = await parent();

  return xrpcCall({ fetch, user }, "app.evermeet.explore.getExplore");
}

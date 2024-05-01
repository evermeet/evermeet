import { apiCall } from "../../../../../lib/api.js";

export async function load({ params, fetch }) {
  return {
    item: await apiCall(fetch, "event/" + params.id),
    tab: params.tab,
  };
}

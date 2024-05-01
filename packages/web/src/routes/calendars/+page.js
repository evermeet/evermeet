import { xrpcCall } from "../../lib/api.js";

export async function load({ params, fetch, parent }) {
  const data = await parent();

  return {
    calendars: await xrpcCall(
      { fetch, user: data.user },
      "app.evermeet.calendar.getUserCalendars",
    ),
  };
}

import { EntitySchema, wrap } from "@mikro-orm/core";

class Watch {}

export const WatchEntity = new EntitySchema({
  name: "Watch",
  class: Watch,
  extends: "BaseEntity",
  properties: {
    calendarDid: {
      type: "string",
    },
    eventId: {
      type: "string",
    },
    authorDid: {
      type: "string",
    },
  },
});

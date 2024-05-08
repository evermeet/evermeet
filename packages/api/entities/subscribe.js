import { EntitySchema, wrap } from "@mikro-orm/core";

class Subscribe {}

export const SubscribeEntity = new EntitySchema({
  name: "Subscribe",
  class: Subscribe,
  extends: "BaseEntity",
  properties: {
    calendarDid: {
      type: "string",
    },
    authorDid: {
      type: "string",
    },
  },
});

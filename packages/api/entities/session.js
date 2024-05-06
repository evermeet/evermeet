import { EntitySchema, wrap } from "@mikro-orm/core";
import { ObjectId } from "../lib/db.js";

export class Session {}

export const schema = new EntitySchema({
  name: "Session",
  class: Session,
  properties: {
    _id: {
      type: "string",
      maxLength: 32,
      primary: true,
      onCreate: () => ObjectId(),
    },
    userId: {
      type: "string",
    },
    token: {
      type: "string",
    },
    createdOn: {
      type: "string",
      format: "date-time",
      onCreate: () => new Date().toISOString(),
    },
  },
});

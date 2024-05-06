import { EntitySchema, wrap } from "@mikro-orm/core";
import { createId } from "../lib/db.js";

export const BaseEntity = new EntitySchema({
  name: "BaseEntity",
  abstract: true,
  properties: {
    id: {
      type: "string",
      primary: true,
      onCreate: (e) => e.id || createId(),
    },
    createdOn: {
      type: "date",
      onCreate: () => new Date().toISOString(),
    },
  },
});

export const BaseDidEntity = new EntitySchema({
  name: "BaseDidEntity",
  extends: "BaseEntity",
  abstract: true,
  properties: {
    did: {
      type: "string",
      unique: true,
    },
    handle: {
      type: "string",
      format: "handle",
      nullable: true,
      unique: true,
    },
    signingKey: {
      type: "object",
      object: true,
      nullable: true,
      // lazy: true
    },
    rotationKey: {
      type: "object",
      nullable: true,
      lazy: true,
    },
  },
});

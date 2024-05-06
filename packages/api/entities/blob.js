import { EntitySchema, wrap } from "@mikro-orm/core";

class Blob {
  view() {
    return {
      cid: this.cid,
      mimeType: this.mimeType,
      size: this.size,
    };
  }

  raw() {
    return this.data;
  }

  toResponse() {
    return {
      encoding: this.mimeType,
      body: this.data,
    };
  }
}

export const schema = new EntitySchema({
  name: "Blob",
  class: Blob,
  properties: {
    cid: {
      type: "string",
      primary: true,
    },
    dids: {
      type: "array",
      nullable: true,
      onCreate: () => [],
    },
    mimeType: {
      type: "string",
    },
    size: {
      type: "number",
    },
    data: {
      type: "blob",
      lazy: true,
    },
    createdOn: {
      type: "string",
      format: "date-time",
      onCreate: () => new Date().toISOString(),
    },
  },
});

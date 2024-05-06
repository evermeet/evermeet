import { expect, test, bench, describe, beforeAll, beforeEach } from "vitest";
import { Evermeet } from "./evermeet.js";
import { before } from "lodash";

export async function xrpcTestBuilder() {
  const em = new Evermeet({
    configFile: "../../config.test.yaml",
  });

  return {
    describeXrpc(id, cb) {
      return describe(id, () => {
        let xrpc, url;
        beforeAll(async () => {
          await em.start();
          xrpc = em.xrpc.service(url);
          url = `http://${em.config.api.host}:${em.config.api.port}/xrpc/`;
          console.log(url);
        });
        const ctx = {
          id,
          em,
          xrpc,
          test,
          exec: async () => xrpc.call(id),
        };
        return cb(ctx);
      });
    },
  };
}

import { listen } from "listhen";
import { createApp, toNodeListener } from "h3";
import {
  createIPX,
  ipxFSStorage,
  ipxHttpStorage,
  createIPXH3Handler,
} from "ipx";

const ipx = createIPX({
  //storage: ipxFSStorage({ dir: "./static" }),
  httpStorage: ipxHttpStorage({ domains: ["localhost:3001"] }),
});

const app = createApp().use("/", createIPXH3Handler(ipx));

listen(toNodeListener(app));
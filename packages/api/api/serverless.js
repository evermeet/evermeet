import { prepareServer  } from "../lib/api.js";

const app = prepareServer();

export default async function handler(req, res) {
    await app.ready()
    app.server.emit('request', req, res)
}
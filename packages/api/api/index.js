import { prepareServer  } from "../lib/api.js";

const api = prepareServer();

export default async function handler(req, res) {
    await api.ready()
    api.server.emit('request', req, res)
}
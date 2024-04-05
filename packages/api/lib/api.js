
import { initDatabase, loadMockData } from '../lib/db.js'
import { parse } from 'yaml'
import fs from 'node:fs'

import { makeRoutes } from './routes.js';
import { makeCollections } from './collections.js';

export class API {

  constructor () {
    this.models = {}
    this.config = parse(fs.readFileSync('../../config.yaml', 'utf-8'))
    this.pkg = JSON.parse(fs.readFileSync('../../package.json', 'utf-8'))
  }

  async init () {
    this.db = await initDatabase(this)
    this.cols = await this.db.addCollections(makeCollections(this))

    // import mock-data
    await loadMockData(this)
  }

  async makeRoutes(app) {
    return makeRoutes(app, this)
  }
}
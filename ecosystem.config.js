//const { parse } = require('yaml')
//const fs = require('node:fs')
//const config = parse(fs.readFileSync('./config.yaml', 'utf-8'))

module.exports = [{
    script: './index.js',
    cwd: './packages/api',
    name: 'api',
  }, {
    script: './build/index.js',
    cwd: './packages/web',
    env: {
      PORT: 3001,
      VITE_BACKEND_URL: 'http://0.0.0.0:3000',
      //VITE_BACKEND_URL_PUBLIC: `https://${config.domain}`
    },
    name: 'web'
  }]
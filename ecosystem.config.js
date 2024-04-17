const { parse } = require('yaml')
const { defaultsDeep } = require('lodash')
const fs = require('node:fs')
const defaults = parse(fs.readFileSync('./config.defaults.yaml', 'utf-8'))
const local = parse(fs.readFileSync('./config.yaml', 'utf-8'))
const config = defaultsDeep(local, defaults)

const services = []

if (config.api.enabled) {
  services.push({
    name: 'evermeet-api',
    script: './index.js',
    cwd: './packages/api',
    interpreter: config.api.runtime,
    watch: process.env.NODE_ENV === 'development',
    env: {
      PORT: config.api.port,
      HOST: config.api.host,
    }
  })
}

if (config.web.enabled) {
  if (process.env.NODE_ENV === 'development') {
    services.push({
      name: 'evermeet-web-dev',
      script: './node_modules/vite/bin/vite.js',
      args: `--host ${config.web.host} --port ${config.web.port} dev`,
      cwd: './packages/web',
      env: {
        PORT: config.web.port,
        VITE_BACKEND_URL: `http://${config.api.host}:${config.api.port}`,
        //VITE_BACKEND_URL_PUBLIC: `https://${config.domain}`
      },
    })
  } else {
    services.push({
      name: 'evermeet-web-prod',
      script: './build/index.js',
      cwd: './packages/web',
      env: {
        PORT: config.web.port,
        VITE_BACKEND_URL: `http://${config.api.host}:${config.api.port}`,
        //VITE_BACKEND_URL_PUBLIC: `https://${config.domain}`
      },
    })   
  }
}
console.log(services)
module.exports = services
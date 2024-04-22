import { loadYaml, defaultsDeep } from './utils.js'
import Ajv from 'ajv'
import addFormats from 'ajv-formats'

const ajv = new Ajv()
addFormats(ajv)

export function loadConfig (configFile, configDefaultsFile, schema) {
  const defaults = loadYaml(configDefaultsFile)
  const localConfig = loadYaml(configFile)
  const config = defaultsDeep(localConfig, defaults)

  const { valid, errors } = validateConfig(schema, config)
  if (!valid) {
    console.error(errors)
    throw new Error('Config validation error')
  }
  return config
}

export function validateConfig (schema, config) {
  const validate = ajv.compile(schema)
  const valid = validate(config)

  return { valid, errors: validate.errors }
}

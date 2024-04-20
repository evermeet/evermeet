import { parse, stringify } from 'yaml'
import { readFileSync, readdirSync } from 'node:fs'
import { join } from 'node:path'
import _ from 'lodash'

export function loadYaml (fn) {
    return parse(readFileSync(fn, 'utf-8'))
}

export const defaultsDeep = _.defaultsDeep

export { 
    stringify,
    parse,
    readdirSync,
    join
}
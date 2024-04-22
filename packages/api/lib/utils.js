import { parse, stringify } from 'yaml'
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'
import _ from 'lodash'

export function loadYaml (fn) {
  return parse(readFileSync(fn, 'utf-8'))
}

export function loadYamlDir (dir) {
  const data = {}
  for (const fn of readdirSync(dir)) {
    const fp = join(dir, fn)
    if (statSync(fp).isDirectory()) {
      data[fn] = loadYamlDir(fp)
      continue
    }
    const haveExt = fn.match(/^(.+)\.([^\.]+)$/)
    const [n, ext] = haveExt ? haveExt.slice(1) : [fn, null]
    if (ext !== 'yaml') {
      continue
    }
    data[n] = loadYaml(fp)
  }
  return data
}

export function loadYamlDirList (dir, arr = [], p = []) {
  for (const fn of readdirSync(dir)) {
    const fp = join(dir, fn)
    const isDirectory = statSync(fp).isDirectory()
    if (isDirectory) {
      loadYamlDirList(fp, arr, [...p, fn])
      continue
    }
    const haveExt = fn.match(/^(.+)\.([^\.]+)$/)
    const [n, ext] = haveExt ? haveExt.slice(1) : [fn, null]
    if (ext !== 'yaml') {
      continue
    }
    arr.push({ id: [...p, n].join('.'), data: loadYaml(fp) })
  }
  return arr
}

export const defaultsDeep = _.defaultsDeep

export {
  stringify,
  parse,
  readdirSync,
  join
}

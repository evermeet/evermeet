import { parse, stringify } from 'yaml'
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'
import _ from 'lodash'
import { WASMagic } from "wasmagic";
const magic = await WASMagic.create();

export function detect (buffer) {
  return magic.detect(buffer)
}

export function loadFile (fn, enc = 'utf-8') {
  return readFileSync(fn, enc)
}

export function loadYaml (fn) {
  return parse(loadFile(fn, 'utf-8'))
}

export function loadFileWithInfo (fn, enc) {
  const stat = statSync(fn)
  return {
    size: stat.size,
    data: loadFile(fn, enc)
  }
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

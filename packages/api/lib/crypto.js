import crypto from 'node:crypto'

export function sha256 (str) {
  const hash = crypto.createHash('sha256')
  const data = hash.update(str, 'utf-8')
  return data.digest('hex')
}

export function sha512 (str) {
  const hash = crypto.createHash('sha512')
  const data = hash.update(str, 'utf-8')
  return data.digest('hex')
}

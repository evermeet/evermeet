import { Secp256k1Keypair } from '@atproto/crypto'
import * as plc from '@evermeet/did-plc-lib'

function base64ToBytes (base64) {
  const binString = atob(base64)
  return Uint8Array.from(binString, (m) => m.codePointAt(0))
}

function bytesToBase64 (bytes) {
  const binString = Array.from(bytes, (byte) =>
    String.fromCodePoint(byte)
  ).join('')
  return btoa(binString)
}

export async function createDid (handle, ctx) {
  const signingKey = await Secp256k1Keypair.create({ exportable: true })
  const rotationKey = await Secp256k1Keypair.create({ exportable: true })

  const plcClient = new plc.Client(ctx.api.config.plcServer)
  const did = await plcClient.createDid({
    signingKey: signingKey.did(),
    handle,
    pds: `https://${ctx.api.config.domain}`,
    rotationKeys: [rotationKey.did()],
    signer: rotationKey
  })

  return {
    did,
    signingKey,
    rotationKey
  }
}

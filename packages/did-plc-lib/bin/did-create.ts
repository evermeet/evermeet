#!/usr/bin/env ts-node

import { Secp256k1Keypair } from '@atproto/crypto'
import { Client } from '../'

export async function main() {
  const url = process.argv[2]
  const handle = process.argv[3]
  console.log({ url, handle })
  const signingKey = await Secp256k1Keypair.create()
  const rotationKey = await Secp256k1Keypair.create()
  const client = new Client(url)

  const did = await client.createDid({
    signingKey: signingKey.did(),
    handle,
    pds: url,
    rotationKeys: [rotationKey.did()],
    signer: rotationKey,
  })
  console.log(`Created did: ${url}/${did}`)
}

main()
import { expect, test, bench, describe, beforeAll, beforeEach } from 'vitest'
import { xrpcTestBuilder } from '../../lib/test.js'

const { em, describeXrpc } = await xrpcTestBuilder()

describeXrpc('app.evermeet.server.describeServer', ({ exec, id, test }) => {
  let data;
  beforeEach(async () => {
    ({ data } = await exec())
  })
  test('It contains', async () => {
    expect(data).toBeTypeOf('object')
  })
})
import { $ } from 'bun'


const json = JSON.stringify({ hello: 42 })

await $`nats publish chat xxx@@@@@`
//console.log(text)
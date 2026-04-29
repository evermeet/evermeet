

export const index = 3;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/auth/login/_page.svelte.js')).default;
export const imports = ["_app/immutable/nodes/3.8EuRPkJU.js","_app/immutable/chunks/DdRuSZuU.js","_app/immutable/chunks/CFKVnMbq.js","_app/immutable/chunks/CQ_juSQk.js"];
export const stylesheets = ["_app/immutable/assets/3.CLf6Yk9D.css"];
export const fonts = [];

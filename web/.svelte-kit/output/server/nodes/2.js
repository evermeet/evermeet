

export const index = 2;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_page.svelte.js')).default;
export const imports = ["_app/immutable/nodes/2.Ct0VNVmc.js","_app/immutable/chunks/DdRuSZuU.js","_app/immutable/chunks/CFKVnMbq.js","_app/immutable/chunks/CQ_juSQk.js"];
export const stylesheets = ["_app/immutable/assets/2.CdskE0sj.css"];
export const fonts = [];

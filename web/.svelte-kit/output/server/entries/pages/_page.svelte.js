import "../../chunks/index-server.js";
import "../../chunks/dev.js";
import "../../chunks/api.js";
//#region src/routes/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		$$renderer.push(`<main class="svelte-1uha8ag"><h1 class="svelte-1uha8ag">Upcoming events</h1> `);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<p class="muted svelte-1uha8ag">Loading…</p>`);
		$$renderer.push(`<!--]--></main>`);
	});
}
//#endregion
export { _page as default };

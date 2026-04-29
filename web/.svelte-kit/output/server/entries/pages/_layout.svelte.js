import "../../chunks/index-server.js";
import { R as escape_html } from "../../chunks/dev.js";
import { t as auth } from "../../chunks/auth.svelte.js";
//#region src/routes/+layout.svelte
function _layout($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let { children } = $$props;
		$$renderer.push(`<nav class="svelte-12qhfyh"><a href="/" class="svelte-12qhfyh">Evermeet</a> <div class="nav-right svelte-12qhfyh">`);
		if (!auth.loading) {
			$$renderer.push("<!--[0-->");
			if (auth.user) {
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<a href="/events/create" class="svelte-12qhfyh">+ New event</a> <span class="did svelte-12qhfyh">${escape_html(auth.user.display_name || auth.user.did.slice(0, 20) + "…")}</span> <button class="svelte-12qhfyh">Sign out</button>`);
			} else {
				$$renderer.push("<!--[-1-->");
				$$renderer.push(`<a href="/auth/login" class="svelte-12qhfyh">Sign in</a>`);
			}
			$$renderer.push(`<!--]-->`);
		} else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div></nav> `);
		children($$renderer);
		$$renderer.push(`<!---->`);
	});
}
//#endregion
export { _layout as default };

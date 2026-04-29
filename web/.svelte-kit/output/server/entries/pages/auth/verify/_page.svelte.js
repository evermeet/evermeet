import "../../../../chunks/index-server.js";
import { R as escape_html } from "../../../../chunks/dev.js";
import "../../../../chunks/client.js";
//#region src/routes/auth/verify/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let status = "verifying";
		let error = "";
		$$renderer.push(`<main class="svelte-1xy6sat">`);
		if (status === "verifying") {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<p>Signing you in…</p>`);
		} else {
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<p class="error svelte-1xy6sat">${escape_html(error)}</p> <a href="/auth/login">Try again</a>`);
		}
		$$renderer.push(`<!--]--></main>`);
	});
}
//#endregion
export { _page as default };

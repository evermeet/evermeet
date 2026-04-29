import { L as attr, R as escape_html } from "../../../../chunks/dev.js";
//#region src/routes/auth/login/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let email = "";
		let submitting = false;
		$$renderer.push(`<main class="svelte-1i2smtp">`);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<h1 class="svelte-1i2smtp">Sign in</h1> <p class="muted svelte-1i2smtp">We'll email you a sign-in link — no password needed.</p> <form class="svelte-1i2smtp"><label for="email" class="svelte-1i2smtp">Email</label> <input id="email" type="email"${attr("value", email)} placeholder="you@example.com" autocomplete="email" required="" class="svelte-1i2smtp"/> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> <button type="submit"${attr("disabled", submitting, true)} class="svelte-1i2smtp">${escape_html("Send sign-in link")}</button></form>`);
		$$renderer.push(`<!--]--></main>`);
	});
}
//#endregion
export { _page as default };

import { jstsExtractor, svelteExtractor } from "svelte-i18n-lingui/extractor";

export default {
  locales: ["en", "es", "uk", "cs", "pseudo-LOCALE"],
  pseudoLocale: "pseudo-LOCALE",
  format: "po",
  catalogs: [
    {
      path: "../../locales/{locale}/messages",
      include: ["./src"],
    },
  ],
  sourceLocale: "en",
  extractors: [jstsExtractor, svelteExtractor],
};

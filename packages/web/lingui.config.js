
import { jstsExtractor, svelteExtractor } from 'svelte-i18n-lingui/extractor';

export default {
  locales: ["en", "cs"],
  format: "po",
  catalogs: [
    {
      path: "../../locales/{locale}/messages",
      include: ["./src"],
    },
  ],
  sourceLocale: "en",
  extractors: [jstsExtractor, svelteExtractor]
}
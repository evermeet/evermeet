import * as locales from "date-fns/locale";
import countries from "i18n-iso-countries";
import { getContext } from "svelte";

import CountriesLocale_en from "i18n-iso-countries/langs/en.json";
import CountriesLocale_cs from "i18n-iso-countries/langs/cs.json";
import CountriesLocale_uk from "i18n-iso-countries/langs/uk.json";

countries.registerLocale(CountriesLocale_en);
countries.registerLocale(CountriesLocale_cs);
countries.registerLocale(CountriesLocale_uk);

const supportedLanguages = {
  cs: "Česky",
  en: "English",
  es: "Español",
  uk: "Українська",
  zu: "Pseudolocale",
};

export * from "svelte-i18n-lingui";

export function detectLanguage(lang) {
  let langFile = lang;
  if (lang.substring(0, 2) === "en") {
    langFile = "en";
  }
  if (!supportedLanguages[langFile]) {
    langFile = "en";
  }
  return {
    lang,
    langFile,
  };
}

export function makeLocaleContext(locale) {
  const dateLocaleCode = locale?.lang || "en";

  let dateLocale;
  if (locales[dateLocaleCode]) {
    dateLocale = locales[dateLocaleCode];
  }
  if (dateLocaleCode.substr(0, 2) === "en") {
    dateLocale = locales.enUS;
  }
  return {
    ...locale,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    dateLocale,
  };
}

export function getCountryName(code, lang) {
  if (!supportedLanguages[lang]) {
    lang = "en";
  }
  return countries.getName(code, lang);
}

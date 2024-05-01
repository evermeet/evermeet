import { xrpcCall } from "$lib/api.js";
import { loadConfig } from "$lib/config.js";
import { t, locale } from "$lib/i18n";
//import * as dateLocales from "date-fns/locale";

const supportedLanguages = {
  cs: "Česky",
  en: "English",
};

export async function load({ fetch, cookies, request: { headers } }) {
  const config = await loadConfig();
  let user = false;
  const sessionId = cookies.get("evermeet-session-id");
  if (sessionId) {
    try {
      const session = await xrpcCall(
        { fetch },
        "app.evermeet.auth.getSession",
        null,
        null,
        {
          headers: {
            authorization: "Bearer " + sessionId,
          },
        },
      );
      //session.accessJwt = sessionId
      user = session.user;
      user.token = sessionId;
    } catch (e) {
      console.error(e);
    }
  }
  //"pseudo-LOCALE"
  const lang =
    headers.get("accept-language")?.split(";")[0].split(",")[0] || "en";

  let langFile = lang;
  if (lang.substring(0, 2) === "en") {
    langFile = "en";
  }
  if (!supportedLanguages[langFile]) {
    langFile = "en";
  }

  const { messages } = await import(
    `../../../../locales/${langFile}/messages.ts`
  );

  return {
    user,
    config,
    locale: {
      lang,
      messages,
      //date: dateLocales[lang],
    },
  };
}

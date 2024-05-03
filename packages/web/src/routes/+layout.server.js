import { xrpcCall } from "$lib/api.js";
import { loadConfig } from "$lib/config.js";
import { t, locale, detectLanguage } from "$lib/i18n";
//import * as dateLocales from "date-fns/locale";

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
  const { lang, langFile } = detectLanguage(
    headers.get("accept-language")?.split(";")[0].split(",")[0],
  );

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

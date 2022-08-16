import i18n from "i18next";
import I18nextBrowserLanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";
import resourcesToBackend from "i18next-resources-to-backend";

i18n
  .use(
    resourcesToBackend((language, namespace, callback) => {
      import(`./locales/${language}/${namespace}.json`)
        .then((resources) => {
          callback(null, resources);
        })
        .catch((err) => callback(err, null));
    })
  )
  .use(I18nextBrowserLanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: "en-US",

    debug: true,

    interpolation: {
      escapeValue: false,
    },
  });

export default i18n;

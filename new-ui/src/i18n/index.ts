import i18n from 'i18next'
import I18nextBrowserLanguageDetector from 'i18next-browser-languagedetector'
import resourcesToBackend from 'i18next-resources-to-backend'
import moment from 'moment'
import { initReactI18next } from 'react-i18next'

i18n.on('languageChanged', (lng) => {
  if (lng === 'en') {
    return
  }
  import(`./locales/${lng}/moment.js`)
    .then(({ default: locale }) => {
      moment.locale(lng, locale)

      return true
    })
    .catch((err) => {
      throw err
    })
})

i18n
  .use(
    resourcesToBackend((language, namespace, callback) => {
      import(`./locales/${language}/${namespace}.json`)
        .then((resources) => {
          // eslint-disable-next-line promise/no-callback-in-promise
          callback(null, resources)

          return 'ok'
        })
        .catch((error) => {
          // eslint-disable-next-line promise/no-callback-in-promise
          callback(error, null)
        })
    })
  )
  .use(I18nextBrowserLanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: 'en',
    debug: false,

    interpolation: {
      escapeValue: false // not needed for react as it escapes by default
    }
  })

export default i18n

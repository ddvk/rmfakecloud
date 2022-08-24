import { useTranslation } from 'react-i18next'

export default function Login() {
  const { t } = useTranslation()

  return (
    <div className="grid justify-items-center items-center">
      <div className="w-full max-w-md">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-neutral-200 mb-4 font-serif">{t('login.title')}</h1>
          <p className="text-sm text-neutral-600">{t('login.subtitle')}</p>
        </div>
        <form className="shadow-md rounded px-8 pt-6 pb-8 mb-4">
          <div className="mb-4">
            <label className="block text-neutral-400 font-bold mb-2">{t('login.form.username.label')}</label>
            <input
              className="shadow appearance-none border border-neutral-600 rounded w-full py-3 px-3 bg-slate-900 text-neutral-400 leading-tight focus:outline-none focus:shadow-outline focus:border-neutral-500 placeholder:text-neutral-600"
              id="username"
              type="text"
              placeholder={t('login.form.username.placeholder')}
            />
          </div>
          <div className="mb-6">
            <label className="block text-neutral-400 font-bold mb-2">{t('login.form.password.label')}</label>
            <input
              className="shadow appearance-none border border-neutral-600 rounded w-full py-3 px-3 text-neutral-400 mb-3 bg-slate-900 leading-tight focus:outline-none focus:shadow-outline focus:border-neutral-500 placeholder:text-neutral-600"
              id="password"
              type="password"
              placeholder={t('login.form.password.placeholder')}
            />
          </div>
          <div className="flex items-center justify-between">
            <button
              className="bg-blue-700 hover:bg-blue-600 w-full text-neutral-200 font-bold py-3 rounded focus:outline-none focus:shadow-outline"
              type="button"
            >
              {t('login.form.login-btn')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

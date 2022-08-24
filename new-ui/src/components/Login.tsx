import { Formik } from 'formik'
import { useTranslation } from 'react-i18next'
import { PulseLoader } from 'react-spinners'

function inputClassName(error? : boolean | string) : string {
  let borderColor = "border-neutral-600"
  let focusedBorderColor = "focus:border-neutral-500"

  if (error) {
    borderColor = "border-red-600"
    focusedBorderColor = "focus:border-red-600"
  }

  const classNames = `shadow appearance-none border ${borderColor} rounded w-full py-3 px-3 bg-slate-900 text-neutral-400 leading-tight focus:outline-none focus:shadow-outline ${focusedBorderColor} placeholder:text-neutral-600`

  return classNames
}

export default function Login() {
  const { t } = useTranslation()

  return (
    <div className="grid justify-items-center items-center">
      <div className="w-full max-w-md">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-neutral-200 mb-4 font-serif">
            {t('login.title')}
          </h1>
          <p className="text-sm text-neutral-600">{t('login.subtitle')}</p>
        </div>
        <Formik
          initialValues={{ username: '', password: '' }}
          validate={(values) => {
            const errors: { username?: string; password?: string } = {}

            if (!values.username) {
              errors.username = t('login.form.username.required')
            }

            if (!values.password) {
              errors.password = t('login.form.password.required')
            }

            return errors
          }}
          onSubmit={(values, { setSubmitting }) => {
            console.log(values)
            setTimeout(() => setSubmitting(false), 1000)
          }}
        >
          {({ values, errors, touched, handleSubmit, handleChange, isSubmitting }) => (
            <form
              className="shadow-md rounded px-8 pt-6 pb-8 mb-4"
              onSubmit={handleSubmit}
            >
              <div className="mb-4">
                <label className="block text-neutral-400 font-bold mb-2">
                  {t('login.form.username.label')}
                </label>
                <input
                  className={inputClassName(errors.username && touched.username)}
                  id="username"
                  type="text"
                  placeholder={t('login.form.username.placeholder')}
                  onChange={handleChange}
                  value={values.username}
                />
                {errors.username && touched.username ? (
                  <p className="mt-2 text-red-600 text-xs animate-pulse">{errors.username}</p>
                ) : (
                  <></>
                )}
              </div>
              <div className="mb-9">
                <label className="block text-neutral-400 font-bold mb-2">
                  {t('login.form.password.label')}
                </label>
                <input
                  className={inputClassName(errors.password && touched.password)}
                  id="password"
                  type="password"
                  placeholder={t('login.form.password.placeholder')}
                  onChange={handleChange}
                  value={values.password}
                />
                {errors.password && touched.password ? (
                  <p className="mt-2 text-red-600 text-xs">{errors.password}</p>
                ) : (
                  <></>
                )}
              </div>
              <div className="flex items-center justify-between">
                <button
                  className="bg-blue-700 hover:bg-blue-600 w-full text-neutral-200 font-bold py-3 rounded focus:outline-none focus:shadow-outline disabled:bg-blue-500"
                  type="submit"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    <PulseLoader
                      cssOverride={{ lineHeight: 0, padding: '6px 0' }}
                      speedMultiplier={0.8}
                      size={8}
                      color="#e5e5e5"
                    ></PulseLoader>
                  ) : (
                    t('login.form.login-btn')
                  )}
                </button>
              </div>
            </form>
          )}
        </Formik>
      </div>
    </div>
  )
}

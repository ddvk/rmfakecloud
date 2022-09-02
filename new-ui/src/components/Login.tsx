import { Formik } from 'formik'
import { useTranslation } from 'react-i18next'
import { PulseLoader } from 'react-spinners'
import { useLocation, useNavigate } from 'react-router-dom'
import { StatusCodes } from 'http-status-codes'
import { toast, ToastContainer } from 'react-toastify'
import { AxiosError } from 'axios'
import { Helmet } from 'react-helmet-async'

import { fullSiteTitle } from '../utils/site'
import { login } from '../api/auth'

function inputClassName(error?: boolean | string): string {
  let borderColor = 'border-neutral-600'
  let focusedBorderColor = 'focus:border-neutral-500'

  if (error) {
    borderColor = 'border-red-600'
    focusedBorderColor = 'focus:border-red-600'
  }

  return `shadow appearance-none border ${borderColor} rounded w-full py-3 px-3 bg-slate-900 text-neutral-400 leading-tight focus:outline-none focus:shadow-outline ${focusedBorderColor} placeholder:text-neutral-600`
}

export default function Login() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()

  return (
    <>
      <Helmet>
        <title>{fullSiteTitle(t('site.titles.login'))}</title>
      </Helmet>
      <div className="grid items-center justify-items-center">
        <div className="w-full max-w-md">
          <div className="text-center">
            <h1 className="mb-4 font-serif text-2xl font-bold text-neutral-200">
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
              // eslint-disable-next-line promise/catch-or-return
              login(values.username, values.password)
                .then(() => {
                  const { from } = (location.state as { from: { pathname: string } }) || {
                    from: { pathname: '/' }
                  }

                  navigate(from.pathname, { replace: true })

                  return 'ok'
                })
                .catch((err: AxiosError) => {
                  const { response } = err

                  if (response && response.status === StatusCodes.UNAUTHORIZED) {
                    toast.error(t('login.username_or_password_error'), {
                      position: 'top-center',
                      theme: 'dark'
                    })
                  }
                })
                .finally(() => {
                  setSubmitting(false)
                })
            }}
          >
            {({ values, errors, touched, handleSubmit, handleChange, isSubmitting }) => (
              <form
                className="mb-4 rounded px-8 pt-6 pb-8 shadow-md"
                onSubmit={handleSubmit}
              >
                <div className="mb-4">
                  <label className="mb-2 block font-bold text-neutral-400">
                    {t('login.form.username.label')}
                  </label>
                  <input
                    className={inputClassName(errors.username && touched.username)}
                    id="username"
                    placeholder={t('login.form.username.placeholder')}
                    type="text"
                    value={values.username}
                    onChange={handleChange}
                  />
                  {errors.username && touched.username ? (
                    <p className="mt-2 text-xs text-red-600">{errors.username}</p>
                  ) : (
                    <></>
                  )}
                </div>
                <div className="mb-9">
                  <label className="mb-2 block font-bold text-neutral-400">
                    {t('login.form.password.label')}
                  </label>
                  <input
                    className={inputClassName(errors.password && touched.password)}
                    id="password"
                    placeholder={t('login.form.password.placeholder')}
                    type="password"
                    value={values.password}
                    onChange={handleChange}
                  />
                  {errors.password && touched.password ? (
                    <p className="mt-2 text-xs text-red-600">{errors.password}</p>
                  ) : (
                    <></>
                  )}
                </div>
                <div className="flex items-center justify-between">
                  <button
                    className="w-full rounded bg-blue-700 py-3 font-bold text-neutral-200 hover:bg-blue-600 focus:outline-none disabled:bg-blue-500"
                    disabled={isSubmitting}
                    type="submit"
                  >
                    {isSubmitting ? (
                      <PulseLoader
                        color="#e5e5e5"
                        cssOverride={{ lineHeight: 0, padding: '6px 0' }}
                        size={8}
                        speedMultiplier={0.8}
                      />
                    ) : (
                      t('login.form.login-btn')
                    )}
                  </button>
                </div>
              </form>
            )}
          </Formik>
          <ToastContainer />
        </div>
      </div>
    </>
  )
}

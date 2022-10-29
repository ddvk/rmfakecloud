import { ErrorMessage, Field, Form, Formik } from 'formik'
import { Helmet } from 'react-helmet-async'
import { useTranslation } from 'react-i18next'
import { PulseLoader } from 'react-spinners'
import { toast } from 'react-toastify'
import * as Yup from 'yup'

import { fullSiteTitle } from '../utils/site'
import { inputClassName } from '../utils/form'
import { resetPassword } from '../api/auth'

import UserIdField from './UserIdField'

export default function ResetPassword() {
  const { t } = useTranslation()
  const initialValues = {
    userid: '',
    currentPassword: '',
    newPassword: '',
    newPasswordConfirmation: ''
  }
  const validationSchema = Yup.object().shape({
    userid: Yup.string().required(),
    currentPassword: Yup.string().required(t('reset_password.form.current_password.required')),
    newPassword: Yup.string()
      .required(t('reset_password.form.new_password.required'))
      .min(6, t('reset_password.form.new_password.min')),
    newPasswordConfirmation: Yup.string()
      .required(t('reset_password.form.new_password.required'))
      .equals([Yup.ref('newPassword')], t('reset_password.form.new_password_confirmation.equals'))
  })

  return (
    <>
      <Helmet>
        <title>{fullSiteTitle(t('site.titles.reset_password'))}</title>
      </Helmet>
      <div className="min-h-[calc(100vh-63px)] bg-slate-900 text-neutral-400 md:min-h-[calc(100vh-57px)]">
        <div className="mx-auto max-w-4xl">
          <div className="mx-4 py-8">
            <h1 className="mb-8 text-2xl font-semibold text-neutral-200">
              {t('nav.change_password')}
            </h1>
            <Formik
              initialValues={initialValues}
              validationSchema={validationSchema}
              onSubmit={(values, { setSubmitting, resetForm }) => {
                resetPassword(values)
                  .then(() => {
                    resetForm()
                    toast.success(t('reset_password.form.success'))

                    return 'ok'
                  })
                  .catch((err) => {
                    throw err
                  })
                  .finally(() => {
                    setSubmitting(false)
                  })
              }}
            >
              {({ isSubmitting, handleSubmit, errors, touched }) => (
                <Form
                  className="mb-4 pb-8"
                  onSubmit={handleSubmit}
                >
                  <UserIdField
                    name="userid"
                    type="hidden"
                  />
                  <div className="mb-4">
                    <label className="mb-2 block font-bold text-neutral-400">
                      {t('reset_password.form.current_password.label')}
                    </label>
                    <Field
                      className={inputClassName(errors.currentPassword && touched.currentPassword)}
                      name="currentPassword"
                      type="password"
                    />
                    <ErrorMessage
                      className="mt-2 text-xs text-red-600"
                      component="div"
                      name="currentPassword"
                    />
                  </div>
                  <div className="mb-4">
                    <label className="mb-2 block font-bold text-neutral-400">
                      {t('reset_password.form.new_password.label')}
                    </label>
                    <Field
                      className={inputClassName(errors.newPassword && touched.newPassword)}
                      name="newPassword"
                      type="password"
                    />
                    <ErrorMessage
                      className="mt-2 text-xs text-red-600"
                      component="div"
                      name="newPassword"
                    />
                  </div>
                  <div className="mb-8">
                    <label className="mb-2 block font-bold text-neutral-400">
                      {t('reset_password.form.new_password_confirmation.label')}
                    </label>
                    <Field
                      className={inputClassName(
                        errors.newPasswordConfirmation && touched.newPasswordConfirmation
                      )}
                      name="newPasswordConfirmation"
                      type="password"
                    />
                    <ErrorMessage
                      className="mt-2 text-xs text-red-600"
                      component="div"
                      name="newPasswordConfirmation"
                    />
                  </div>
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
                      t('reset_password.form.submit-btn')
                    )}
                  </button>
                </Form>
              )}
            </Formik>
          </div>
        </div>
      </div>
    </>
  )
}

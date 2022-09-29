import { ErrorMessage, Field, Form, Formik } from 'formik'
import { useEffect, useState } from 'react'
import { Helmet } from 'react-helmet-async'
import { useTranslation } from 'react-i18next'
import { Navigate, useParams } from 'react-router-dom'
import { PulseLoader } from 'react-spinners'
import { toast } from 'react-toastify'
import * as Yup from 'yup'

import { getUser, updateUser } from '../api/users'
import { inputClassName } from '../utils/form'
import { fullSiteTitle } from '../utils/site'

export default function EditUser() {
  const { userId } = useParams()
  const { t } = useTranslation()
  const [navigateTo, setNavigateTo] = useState('')
  const [user, setUser] = useState<{ userid?: string; email?: string } | null>()

  useEffect(() => {
    if (!userId) {
      return
    }
    getUser(userId as string)
      .then((response) => {
        setUser(response.data)

        return true
      })
      .catch((err) => {
        throw err
      })
  }, [userId])

  const validationSchema = Yup.object().shape({
    user_id: Yup.string().required(t('new_user.form.user_id.required')),
    email: Yup.string()
      .required(t('new_user.form.email.required'))
      .email(t('new_user.form.email.invalid')),
    password: Yup.string().min(6, t('new_user.form.password.min')),
    password_confirmation: Yup.string().equals(
      [Yup.ref('password')],
      t('new_user.form.password_confirmation.equals')
    )
  })

  return (
    <>
      <Helmet>
        <title>{fullSiteTitle(t('site.titles.edit_user'))}</title>
      </Helmet>
      {navigateTo ? (
        <Navigate
          replace={true}
          to={navigateTo}
        />
      ) : (
        <></>
      )}
      <div className="min-h-[calc(100vh-63px)] bg-slate-900 text-neutral-400 md:min-h-[calc(100vh-57px)]">
        <div className="mx-auto max-w-4xl">
          <div className="mx-4 py-8">
            <h1 className="mb-8 text-2xl font-semibold text-neutral-200">
              {t('site.titles.edit_user')}
            </h1>
            {user ? (
              <Formik
                initialValues={{
                  user_id: user.userid || '',
                  email: user.email || '',
                  password: '',
                  password_confirmation: ''
                }}
                validationSchema={validationSchema}
                onSubmit={(values, { setSubmitting }) => {
                  setSubmitting(true)

                  updateUser(values)
                    .then(() => {
                      setNavigateTo('/users')
                      toast.success(t('notifications.user_created'))

                      return true
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
                    <div className="mb-4">
                      <label className="mb-2 block font-bold text-neutral-400">
                        {t('new_user.form.user_id.label')}
                      </label>
                      <Field
                        readOnly
                        className={`${inputClassName(
                          errors.user_id && touched.user_id
                        )} bg-slate-200/10`}
                        name="user_id"
                        type="text"
                      />
                      <ErrorMessage
                        className="mt-2 text-xs text-red-600"
                        component="div"
                        name="user_id"
                      />
                    </div>
                    <div className="mb-4">
                      <label className="mb-2 block font-bold text-neutral-400">
                        {t('new_user.form.email.label')}
                      </label>
                      <Field
                        className={inputClassName(errors.email && touched.email)}
                        name="email"
                        type="text"
                      />
                      <ErrorMessage
                        className="mt-2 text-xs text-red-600"
                        component="div"
                        name="email"
                      />
                    </div>
                    <div className="mb-8">
                      <label className="mb-2 block font-bold text-neutral-400">
                        {t('new_user.form.password.label')}
                      </label>
                      <Field
                        className={inputClassName(errors.password && touched.password)}
                        name="password"
                        type="password"
                      />
                      <ErrorMessage
                        className="mt-2 text-xs text-red-600"
                        component="div"
                        name="password"
                      />
                    </div>
                    <div className="mb-8">
                      <label className="mb-2 block font-bold text-neutral-400">
                        {t('new_user.form.password_confirmation.label')}
                      </label>
                      <Field
                        className={inputClassName(
                          errors.password_confirmation && touched.password_confirmation
                        )}
                        name="password_confirmation"
                        type="password"
                      />
                      <ErrorMessage
                        className="mt-2 text-xs text-red-600"
                        component="div"
                        name="password_confirmation"
                      />
                    </div>
                    <button
                      className="w-full rounded bg-blue-700 py-3 font-bold text-neutral-200 hover:bg-blue-600 focus:outline-none disabled:bg-blue-500"
                      disabled={isSubmitting || !user}
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
            ) : (
              <></>
            )}
          </div>
        </div>
      </div>
    </>
  )
}

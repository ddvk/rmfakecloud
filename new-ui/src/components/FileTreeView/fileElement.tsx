/* eslint-disable tailwindcss/no-custom-classname */

import { DocumentTextIcon } from '@heroicons/react/outline'
import { ErrorMessage, Field, Form, Formik } from 'formik'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'react-toastify'
import * as Yup from 'yup'
import { PulseLoader } from 'react-spinners'

import { renameDocument } from '../../api/document'
import { inputClassName } from '../../utils/form'
import { EpubIcon, PDFIcon } from '../../utils/icons'

import { HashDocElementProp } from './props'

export default function FileElement(params: HashDocElementProp) {
  const { doc, onClickDoc, onDocEditingDiscard, onDocRenamed, className, ...remainParams } = params
  const [unmountForm, setUnmountForm] = useState(false)
  const { preMode } = doc
  let { mode } = doc

  if (!mode) {
    mode = 'display'
  }
  const { t } = useTranslation()

  const formFadeout = () => {
    setTimeout(() => {
      setUnmountForm(true)
    }, 500)
  }

  const editingToDisplay = mode === 'display' && preMode === 'editing'

  if (editingToDisplay) {
    formFadeout()
  }

  const validationSchema = Yup.object().shape({
    name: Yup.string().required(t('documents.rename_form.name.required'))
  })

  const formDom = (
    <Formik
      initialValues={{ name: doc.name }}
      validationSchema={validationSchema}
      onSubmit={(values, { setSubmitting }) => {
        setSubmitting(true)

        renameDocument(doc.id, values.name)
          .then(() => {
            toast.success(t('notifications.document_renamed'))
            onDocRenamed && onDocRenamed(doc, values.name)

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
          className={`w-full overflow-hidden ${
            editingToDisplay ? 'animate-roll-up' : 'animate-roll-down'
          }`}
          onSubmit={handleSubmit}
        >
          <div className="mb-4">
            <label className="mb-2 block font-bold text-neutral-400">
              {t('documents.rename_form.name.label')}
            </label>
            <Field
              autoFocus={true}
              className={inputClassName(errors.name && touched.name)}
              name="name"
              type="text"
            />
            <ErrorMessage
              className="mt-2 text-xs text-red-600"
              component="div"
              name="name"
            />
          </div>
          <div className="flex">
            <button
              className="mr-2 w-full basis-1/2 rounded border border-slate-600 py-3 font-bold text-neutral-200 focus:outline-none"
              type="button"
              onClick={(e) => {
                e.stopPropagation()
                setUnmountForm(false)
                onDocEditingDiscard && onDocEditingDiscard(doc)
              }}
            >
              {t('documents.rename_form.cancel-btn')}
            </button>
            <button
              className="w-full basis-1/2 rounded bg-blue-700 py-3 font-bold text-neutral-200 hover:bg-blue-600 focus:outline-none disabled:bg-blue-500"
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
                t('documents.rename_form.submit-btn')
              )}
            </button>
          </div>
        </Form>
      )}
    </Formik>
  )
  let innerDom: JSX.Element

  if (mode === 'editing') {
    innerDom = formDom
  } else {
    innerDom = (
      <>
        {doc.extension === '.epub' ? (
          <EpubIcon className="top-[-1px] mr-2 h-6 w-6 shrink-0" />
        ) : doc.extension === '.pdf' ? (
          <PDFIcon className="top-[-1px] mr-2 h-6 w-6 shrink-0" />
        ) : (
          <DocumentTextIcon className="top-[-1px] mr-2 h-6 w-6 shrink-0" />
        )}
        <p className="max-w-[calc(100%-28px)] overflow-hidden text-ellipsis whitespace-nowrap leading-6">
          {doc.name}
        </p>
      </>
    )
  }

  return (
    <>
      <div
        className={`flex cursor-pointer select-none py-6 ${
          editingToDisplay ? 'animate-fadein' : ''
        } ${className || ''}`}
        {...remainParams}
        onClick={() => {
          onClickDoc && onClickDoc(doc)
        }}
      >
        {innerDom}
      </div>
      {editingToDisplay && !unmountForm ? formDom : <></>}
    </>
  )
}

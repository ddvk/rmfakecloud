import { useEffect, useState } from 'react'
import { Helmet } from 'react-helmet-async'
import { useTranslation } from 'react-i18next'
import moment from 'moment'
import { PulseLoader } from 'react-spinners'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'

import { listUsers, removeUser } from '../api/users'
import { fullSiteTitle } from '../utils/site'
import { AppUser } from '../utils/models'

import { ConfirmationDialog, ConfirmationDialogProps } from './Dialog'

export default function Users() {
  const { t } = useTranslation()
  const [isLoading, setIsLoading] = useState(false)
  const [users, setUsers] = useState<AppUser[]>([])
  const [dialogProps, setDialogProps] = useState<ConfirmationDialogProps>({
    isOpen: false,
    isLoading: false,
    onClose: () => {
      setDialogProps((props) => {
        return {
          ...props,
          isLoading: false,
          isOpen: false,
          onConfirm: () => {
            throw 'not implement'
          }
        }
      })
    },
    onConfirm: () => {
      throw 'not implement'
    },
    title: ''
  })

  useEffect(() => {
    setIsLoading(true)

    listUsers()
      .then((response) => {
        const data = response.data as AppUser[]

        setUsers(data)

        return true
      })
      .catch((err) => {
        throw err
      })
      .finally(() => {
        setIsLoading(false)
      })
  }, [])

  return (
    <>
      <Helmet>
        <title>{fullSiteTitle(t('site.titles.users'))}</title>
      </Helmet>
      <div className="min-h-[calc(100vh-63px)] bg-slate-900 text-neutral-400 md:min-h-[calc(100vh-57px)]">
        <div className="mx-auto max-w-4xl">
          <div className="relative mx-4 py-8">
            <h1 className="mb-8 text-2xl font-semibold text-neutral-200">{t('nav.users')}</h1>
            <div className="mb-4 text-sm font-semibold text-sky-600 hover:text-sky-500">
              <Link to="/users/new">{t('users.list_table.columns.actions.new_user')}</Link>
            </div>
            <div className="relative overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="bg-slate-800 text-xs font-bold uppercase">
                  <tr>
                    <th className="whitespace-nowrap py-3 px-6">
                      {t('users.list_table.columns.user_id.title')}
                    </th>
                    <th className="whitespace-nowrap py-3 px-6">
                      {t('users.list_table.columns.email.title')}
                    </th>
                    <th className="whitespace-nowrap py-3 px-6">
                      {t('users.list_table.columns.name.title')}
                    </th>
                    <th className="whitespace-nowrap py-3 px-6">
                      {t('users.list_table.columns.created_at.title')}
                    </th>
                    <th className="whitespace-nowrap py-3 px-6">
                      {t('users.list_table.columns.actions.title')}
                    </th>
                  </tr>
                </thead>
                {isLoading ? (
                  <Loader />
                ) : (
                  <tbody>
                    {users.map((user) => (
                      <tr
                        key={`row-${user.userid}`}
                        className="border-b border-slate-100/10 font-semibold last:border-none"
                      >
                        <td className="whitespace-nowrap py-4 px-6">{user.userid}</td>
                        <td className="whitespace-nowrap py-4 px-6">{user.email || '-'}</td>
                        <td className="whitespace-nowrap py-4 px-6">{user.name || '-'}</td>
                        <td className="whitespace-nowrap py-4 px-6">
                          {moment(user.CreatedAt).fromNow()}
                        </td>
                        <td className="w-36 whitespace-nowrap py-4 px-6">
                          <div className="flex text-sky-600">
                            <Link
                              className="mr-4 whitespace-nowrap hover:text-sky-500"
                              to={`/users/${user.userid}/edit`}
                            >
                              {t('users.list_table.columns.actions.edit')}
                            </Link>
                            <button
                              className="whitespace-nowrap hover:text-sky-500"
                              onClick={() => {
                                setDialogProps({
                                  ...dialogProps,
                                  isOpen: true,
                                  title: t('users.list_table.columns.actions.remove_confirmation', {
                                    user: user.userid
                                  }),
                                  onConfirm: () => {
                                    setDialogProps((props) => ({ ...props, isLoading: true }))

                                    removeUser(user.userid)
                                      .then(() => {
                                        setUsers((users) => {
                                          return [...users].filter(
                                            (entity) => user.userid !== entity.userid
                                          )
                                        })
                                        dialogProps.onClose()
                                        toast.success(t('notifications.user_removed'))

                                        return true
                                      })
                                      .catch((err) => {
                                        toast.error(t('notifications.failed_user_remove'))
                                        throw err
                                      })
                                      .finally(() => {
                                        dialogProps.onClose()
                                      })
                                  }
                                })
                              }}
                            >
                              {t('users.list_table.columns.actions.remove')}
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                )}
              </table>
            </div>
          </div>
        </div>
      </div>
      <ConfirmationDialog {...dialogProps} />
    </>
  )
}

function Loader() {
  return (
    <tbody>
      <tr>
        <td colSpan={5}>
          <div className="relative mt-12 text-center">
            <PulseLoader
              color="#e5e5e5"
              cssOverride={{ lineHeight: 0, padding: '6px 0' }}
              size={8}
              speedMultiplier={0.8}
            />
          </div>
        </td>
      </tr>
    </tbody>
  )
}

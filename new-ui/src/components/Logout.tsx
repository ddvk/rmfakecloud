import { useEffect, useState } from 'react'
import { Navigate } from 'react-router-dom'
import { PulseLoader } from 'react-spinners'
import { useTranslation } from 'react-i18next'

import { logout } from '../api/auth'

export default function Logout() {
  const { t } = useTranslation()
  const [isLogout, setIsLogout] = useState(false)

  useEffect(() => {
    logout()
      .then(() => {
        setIsLogout(true)

        return 'ok'
      })
      .catch((e) => {
        throw e
      })
  }, [])

  const loadingDom = (
    <div className="grid items-center justify-items-center">
      <div className="text-center text-neutral-400">
        <p>
          <PulseLoader
            color="#a3a3a3"
            cssOverride={{ lineHeight: 0, padding: '6px 0' }}
            size={8}
            speedMultiplier={0.8}
          />
        </p>
        <p className="text-xs font-semibold">{t('logout.loading')}</p>
      </div>
    </div>
  )

  return isLogout ? (
    <Navigate
      replace={true}
      to="/login"
    />
  ) : (
    loadingDom
  )
}

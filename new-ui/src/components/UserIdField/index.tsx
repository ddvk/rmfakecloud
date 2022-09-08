import { FieldAttributes, useField, useFormikContext } from 'formik'
import { useEffect } from 'react'

import { User } from '../../utils/models'
import { getUserProfile } from '../../api/auth'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export default function UserIdField(props: FieldAttributes<any>) {
  const { setFieldValue } = useFormikContext()
  const { name } = props
  const [field] = useField(props)

  useEffect(() => {
    getUserProfile()
      .then((response) => {
        const data = response.data as User

        setFieldValue(name, data.userid)

        return 'ok'
      })
      .catch((error) => {
        throw error
      })
  }, [setFieldValue, name])

  return (
    <input
      {...props}
      {...field}
    />
  )
}

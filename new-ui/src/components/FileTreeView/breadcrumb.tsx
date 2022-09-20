import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'

import { HashDoc } from '../../utils/models'

export interface BreakcrumbItem {
  title: string
  docs: HashDoc[]
}

export default function Breadcrumbs(params: {
  items: BreakcrumbItem[]
  className?: string
  onClickBreadcrumb?: (item: BreakcrumbItem, index: number) => void
  onClickNewFolder?: () => void
}) {
  const { items, className, onClickBreadcrumb, onClickNewFolder } = params
  const { t } = useTranslation()
  const [isShowCreateFolder, setIsShowCreateFolder] = useState(true)

  useEffect(() => {
    setIsShowCreateFolder(items.length <= 1)
  }, [items])

  const innerDom = items.map((item, i) => {
    return (
      <li
        key={`breakcrumb-item-${i}`}
        className="cursor-pointer after:mx-2 after:text-neutral-400 after:content-['>'] last:after:hidden"
        onClick={(e) => {
          e.preventDefault()
          if (onClickBreadcrumb) {
            onClickBreadcrumb(item, i)
          }
        }}
      >
        {item.title}
      </li>
    )
  })

  return (
    <div className={className}>
      <div className="flex">
        <ul className="flex text-sm font-semibold text-sky-600">{innerDom}</ul>
        <div className="ml-auto flex">
          {isShowCreateFolder ? (
            <Link
              title={t('documents.breadcrumbs.new_folder')}
              to="#"
              onClick={(e) => {
                e.preventDefault()
                onClickNewFolder && onClickNewFolder()
              }}
            >
              <svg
                className="h-6 w-6"
                fill="none"
                stroke="currentColor"
                strokeWidth={1.5}
                viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M12 10.5v6m3-3H9m4.06-7.19l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </Link>
          ) : (
            <></>
          )}
        </div>
      </div>
    </div>
  )
}

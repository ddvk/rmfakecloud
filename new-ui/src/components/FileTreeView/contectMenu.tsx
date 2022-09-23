import { Transition } from '@headlessui/react'
import { PencilAltIcon } from '@heroicons/react/outline'
import { MouseEventHandler, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import useWindowDimensions from '../../utils/hooks'
import { HashDoc } from '../../utils/models'

interface ContextMenuProps {
  doc?: HashDoc
  index?: number
  x?: number
  y?: number
  onClickMenuItem?: (target: { doc: HashDoc; index: number; menuItem: MenuItemType }) => void
  onDismissMenu?: () => void
}

type MenuItemType = 'rename'

export default function ContextMenu({
  doc,
  x,
  y,
  index,
  onClickMenuItem,
  onDismissMenu
}: ContextMenuProps) {
  const [isShow, setIsShow] = useState(false)
  const [t] = useTranslation()
  const menuRef = useRef<HTMLDivElement>(null)
  const [translateX, setTranslateX] = useState('')
  const [translateY, setTranslateY] = useState('')
  const { width: windowWidth, height: windowHeight } = useWindowDimensions()

  useEffect(() => {
    if (isShow) {
      setIsShow(false)
      if (doc) {
        setTimeout(() => setIsShow(true))
      }

      return
    }
    setIsShow(Boolean(doc))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [doc])

  useEffect(() => {
    if (!(isShow && menuRef.current && x && y)) {
      return
    }

    const menuWidth = 192

    if (windowWidth - x < menuWidth) {
      setTranslateX('-translate-x-full')
    } else {
      setTranslateX('')
    }

    const menuHeight = 56

    if (windowHeight - y < menuHeight) {
      setTranslateY('-translate-y-full')
    } else {
      setTranslateY('')
    }
  }, [isShow, windowWidth, windowHeight, x, y])

  useEffect(() => {
    function handler() {
      setIsShow(false)
      onDismissMenu && onDismissMenu()
    }

    window.addEventListener('scroll', handler)
    window.addEventListener('mouseup', handler)

    return () => {
      window.removeEventListener('scroll', handler)
      window.removeEventListener('mouseup', handler)
    }
  }, [onDismissMenu])

  const menuItems: { icon: JSX.Element; title: string; menuItem: MenuItemType }[] = [
    {
      icon: <PencilAltIcon className="mr-2 h-6 w-6 shrink-0" />,
      title: t('documents.file_tree_view.menu.rename'),
      menuItem: 'rename'
    }
  ]

  return (
    <Transition
      ref={menuRef}
      className={`fixed z-10 flex w-48 flex-col divide-y divide-slate-100/10 overflow-hidden rounded bg-slate-800 shadow-lg ring-1 ring-slate-100/10 ${translateX} ${translateY}`}
      enter="transition-[width height] duration-300 ease-out"
      enterFrom="w-0 h-0"
      enterTo="w-48 h-fit-content"
      show={isShow}
      style={{ top: y, left: x }}
      unmount={false}
    >
      {menuItems.map(({ icon, title, menuItem }, i) => (
        <ContextMenuItem
          key={`menu-item-${i}`}
          icon={icon}
          title={title}
          onClick={(e) => {
            e.stopPropagation()
            if (doc && index !== undefined) {
              onClickMenuItem && onClickMenuItem({ doc, index, menuItem })
            }
          }}
        />
      ))}
    </Transition>
  )
}

function ContextMenuItem({
  icon,
  title,
  onClick
}: {
  icon: JSX.Element
  title: string
  onClick?: MouseEventHandler
}) {
  return (
    <div className="w-full p-2">
      <button
        className="flex w-full items-center rounded p-2 transition-colors hover:bg-slate-900 hover:text-sky-600"
        onClick={onClick}
        onMouseUp={(e) => {
          e.stopPropagation()
        }}
      >
        {icon}
        <p className="text-sm font-bold">{title}</p>
      </button>
    </div>
  )
}

import { Menu, Transition } from '@headlessui/react'
import { useTranslation } from 'react-i18next'
import { Link, matchPath, useLocation } from 'react-router-dom'
import { ChevronDownIcon, MenuIcon } from '@heroicons/react/solid'
import { Fragment } from 'react'

import Logo from './logo'

interface RouteItem {
  title: string
  path?: string
  children?: RouteItem[]
}

function Nav(props: { items: RouteItem[] }) {
  const { items } = props
  const { pathname } = useLocation()
  const navDom = items.map((route, i) => {
    const classNames = ['text-sm', 'hover:text-neutral-100', 'transition-colors', 'duration-500']

    if (route.path && matchPath(route.path, pathname)) {
      classNames.push('text-neutral-100')
    } else {
      classNames.push('text-neutral-400')
    }

    function InnerDom() {
      if (!route.children) {
        return <>{route.title}</>
      }

      const menuItems = route.children.map((route, j) => {
        return (
          <div
            key={`nav-item-${i}-menuitem-${j}`}
            className="p-3"
          >
            <Menu.Item>
              <Link
                className="transition-colors duration-300 hover:text-neutral-100"
                to={route.path ? route.path : '#'}
              >
                {route.title}
              </Link>
            </Menu.Item>
          </div>
        )
      })

      return (
        <Menu
          as="div"
          className="relative inline-block text-left"
        >
          <Menu.Button>
            {route.title}
            <ChevronDownIcon
              aria-hidden={true}
              className="ml-1 -mr-1 inline h-5 w-5"
            />
          </Menu.Button>
          <Transition
            as={Fragment}
            enter="transition ease-out duration-100"
            enterFrom="transform opacity-0 scale-95"
            enterTo="transform opacity-100 scale-100"
            leave="transition ease-in duration-75"
            leaveFrom="transform opacity-100 scale-100"
            leaveTo="transform opacity-0 scale-95"
          >
            <Menu.Items className="absolute right-0 mt-3 w-56 origin-top-right divide-y divide-gray-700 rounded-md bg-slate-900/95 text-neutral-400 shadow-lg ring-1 ring-black/5 focus:outline-none">
              {menuItems}
            </Menu.Items>
          </Transition>
        </Menu>
      )
    }

    return route.path ? (
      <li key={`nav-item-${i}`}>
        <Link
          className={classNames.join(' ')}
          to={route.path}
        >
          <InnerDom />
        </Link>
      </li>
    ) : (
      <li key={`nav-item-${i}`}>
        <span className={classNames.join(' ')}>
          <InnerDom />
        </span>
      </li>
    )
  })

  return (
    <nav className="font-semibold leading-6">
      <ul className="flex space-x-8">{navDom}</ul>
    </nav>
  )
}

function MobileNav(props: { items: RouteItem[] }) {
  const { items } = props

  const menuItems = items.map((route, i) => {
    const subMenuItems = (route.children || []).map((subRoute, j) => {
      return (
        <li key={`nav-item-${i}-sub-item-${j}`}>
          <Link to={subRoute.path || '#'}>
            <p className="p-3">{subRoute.title}</p>
          </Link>
        </li>
      )
    })

    return (
      <Menu.Item key={`nav-item-${i}`}>
        {route.children ? (
          <div className="bg-slate-800 mx-3 mb-6 rounded shadow-inner">
            <p className="pt-3 text-xs text-neutral-400 font-normal">{route.title}</p>
            <ul>{subMenuItems}</ul>
          </div>
        ) : (
          <Link to={route.path || '#'}>
            <p className="p-3 relative">{route.title}</p>
          </Link>
        )}
      </Menu.Item>
    )
  })

  return (
    <Menu>
      <Menu.Button>
        <MenuIcon className="w-6 h-6 text-neutral-200" />
      </Menu.Button>
      <Transition
        as={Fragment}
        enter="transition-max-h ease-in duration-300"
        enterFrom="max-h-0"
        enterTo="max-h-screen"
        leave="transition-max-h ease-out duration-300"
        leaveFrom="max-h-screen"
        leaveTo="max-h-0"
      >
        <Menu.Items
          as="div"
          className="absolute w-screen -translate-x-[calc(100%-40px)] top-[46px] bg-slate-900 text-neutral-200 text-center z-10 overflow-hidden font-semibold"
        >
          {menuItems}
        </Menu.Items>
      </Transition>
    </Menu>
  )
}

export default function Navbar() {
  const { t } = useTranslation()

  const routes: RouteItem[] = [
    {
      title: t('nav.documents'),
      path: '/'
    },
    {
      title: t('nav.users'),
      path: '/users'
    },
    {
      title: t('nav.devices'),
      path: '/devices'
    },
    {
      title: t('nav.profile'),
      children: [
        {
          title: t('nav.change_password'),
          path: '/profile/change_password'
        },
        {
          title: t('nav.logout'),
          path: '/logout'
        }
      ]
    }
  ]

  return (
    <>
      <div className="relative top-0 w-full flex-none border-slate-900/10 bg-slate-900 backdrop-blur transition-colors duration-500">
        <div className="mx-auto max-w-4xl">
          <div className="mx-4 border-slate-300/10 py-4">
            <div className="relative flex items-center">
              <Link to="/">
                <Logo className="h-5 w-auto fill-gray-100" />
              </Link>
              <div className="relative ml-auto hidden md:flex">
                <Nav items={routes} />
              </div>
              <div className="relative ml-auto md:hidden">
                <MobileNav items={routes} />
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

import { HashDoc } from '../../utils/models'

export interface BreakcrumbItem {
  title: string
  docs: HashDoc[]
}

export default function Breadcrumbs(params: {
  items: BreakcrumbItem[]
  className?: string
  onClickBreadcrumb?: (item: BreakcrumbItem, index: number) => void
}) {
  const { items, className, onClickBreadcrumb } = params

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
      <ul className="flex text-sm font-semibold text-sky-600">{innerDom}</ul>
    </div>
  )
}

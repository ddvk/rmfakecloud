export function inputClassName(error?: boolean | string): string {
  let borderColor = 'border-neutral-600'
  let focusedBorderColor = 'focus:border-neutral-500'

  if (error) {
    borderColor = 'border-red-600'
    focusedBorderColor = 'focus:border-red-600'
  }

  return `shadow appearance-none border ${borderColor} rounded w-full py-3 px-3 bg-slate-900 text-neutral-400 leading-tight focus:outline-none focus:shadow-outline ${focusedBorderColor} placeholder:text-neutral-600`
}

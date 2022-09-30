export function isWindows(): boolean {
  const UA = window.navigator.userAgent

  if (UA) {
    return /windows/i.test(UA)
  }

  return false
}

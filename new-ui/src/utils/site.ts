const SITE_TITLE = 'rmfakecloud'

export function fullSiteTitle(title?: string): string {
  title = title?.trim()

  if (!title) {
    return SITE_TITLE
  }

  return `${title} | ${SITE_TITLE}`
}

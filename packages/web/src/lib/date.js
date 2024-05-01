import { cs, enGB } from 'date-fns/locale';

export const getLocale = (lang) => {
  if (!lang) {
    return null
  }
  if (lang === 'cs') {
    return cs
  }
  if (lang.substr(0,2) === 'en') {
    return enGB
  }
  return null
}

import { format as _format } from 'date-fns'
import { getContext } from 'svelte'

export function format(date, formatStr) {
    const locale = getContext("locale")
    return _format(date, formatStr, { locale: getLocale(locale.lang) })
}
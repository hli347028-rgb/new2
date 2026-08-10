import i18n from '@/language'

const lang = (text: string, variables: any = {}) => {
  let words = i18n.global.t(text, variables)
  if (words === text) words = text
  return words as string
}

export default lang

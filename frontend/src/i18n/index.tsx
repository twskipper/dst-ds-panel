import i18n from "i18next"
import { initReactI18next } from "react-i18next"
import en from "./en.json"
import zh from "./zh.json"

const saved = localStorage.getItem("dst_locale")
const browserLang = navigator.language.startsWith("zh") ? "zh" : "en"

i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
    zh: { translation: zh },
  },
  lng: saved || browserLang,
  fallbackLng: "en",
  interpolation: {
    escapeValue: false,
  },
})

export default i18n

// Helper to change language and persist
export function changeLanguage(lang: string) {
  i18n.changeLanguage(lang)
  localStorage.setItem("dst_locale", lang)
}

// Re-export useI18n as a convenience wrapper for backward compatibility
export { useTranslation as useI18n } from "react-i18next"

import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import ru from './locales/ru.json'

export const DEFAULT_LOCALE = 'ru'
export const FALLBACK_LOCALE = 'en'
export const LOCALE_STORAGE_KEY = 'clicksafe.locale'
export const supportedLocales = ['ru', 'en']

const LOCALE_TAGS = {
  ru: 'ru-RU',
  en: 'en-US',
}

const CAMPAIGN_STATUS_KEYS = {
  DRAFT: 'campaign.status.draft',
  ACTIVE: 'campaign.status.active',
  PAUSED: 'campaign.status.paused',
  COMPLETED: 'campaign.status.completed',
  CANCELED: 'campaign.status.canceled',
}

const TARGET_STATUS_KEYS = {
  PENDING: 'target.status.pending',
  SENT: 'target.status.sent',
  FAILED: 'target.status.failed',
  OPENED: 'target.status.opened',
  CLICKED: 'target.status.clicked',
  SUBMITTED: 'target.status.submitted',
  REPLIED: 'target.status.replied',
}

const EVENT_TYPE_KEYS = {
  MESSAGE_SENT: 'events.messageSent',
  DELIVERY_FAILED: 'events.deliveryFailed',
  EMAIL_OPENED: 'events.emailOpened',
  LINK_OPENED: 'events.linkOpened',
  DATA_SENT: 'events.dataSent',
  MESSAGE_READ: 'events.messageRead',
  MESSAGE_REPLIED: 'events.messageReplied',
}

const KNOWN_ERROR_KEYS = new Map([
  ['Invalid login or password.', 'errors.auth.invalidCredentials'],
  ['Too many sign-in attempts. Wait a moment and try again.', 'errors.auth.tooManyAttempts'],
  ['The request could not be completed.', 'errors.requestFailed'],
  ['Dashboard could not be loaded.', 'errors.dashboardLoad'],
  ['Reports could not be loaded.', 'errors.reportsLoad'],
  ['Campaign could not be loaded.', 'errors.campaignLoad'],
  ['Campaign could not be loaded for editing.', 'errors.campaignEditLoad'],
  ['Monitoring data could not be loaded.', 'errors.monitoringLoad'],
  ['Reference data could not be loaded.', 'errors.referenceLoad'],
  ['Campaign targets could not be loaded.', 'errors.targetsLoad'],
  ['Schedule could not be saved.', 'errors.scheduleSave'],
])

function isValidationError(error) {
  const payload = error?.payload
  return payload?.code === 'invalid_argument'
    && payload?.message === 'validation failed'
    && Array.isArray(payload.details)
}

function translateLoginValidation(message) {
  if (message.includes('cannot be empty')) {
    return globalT('auth.enterCredentials')
  }

  const minMatch = message.match(/at least (\d+) characters/)
  if (minMatch) {
    return globalT('errors.auth.loginTooShort', { count: minMatch[1] })
  }

  const maxMatch = message.match(/(\d+) characters or fewer/)
  if (maxMatch) {
    return globalT('errors.auth.loginTooLong', { count: maxMatch[1] })
  }

  if (message.includes('too common')) {
    return globalT('errors.auth.loginTooCommon')
  }

  return globalT('errors.auth.loginInvalid')
}

function translatePasswordValidation(message) {
  if (message.includes('cannot be empty')) {
    return globalT('auth.enterCredentials')
  }

  if (message.includes('must not start or end with spaces')) {
    return globalT('errors.auth.passwordEdgeSpaces')
  }

  const minMatch = message.match(/at least (\d+) characters/)
  if (minMatch) {
    return globalT('errors.auth.passwordTooShort', { count: minMatch[1] })
  }

  const maxMatch = message.match(/(\d+) characters or fewer/)
  if (maxMatch) {
    return globalT('errors.auth.passwordTooLong', { count: maxMatch[1] })
  }

  if (message.includes('control character')) {
    return globalT('errors.auth.passwordControlCharacter')
  }

  if (message.includes('different characters')) {
    return globalT('errors.auth.passwordDifferentCharacters')
  }

  if (message.includes('three character classes')) {
    return globalT('errors.auth.passwordCharacterClasses')
  }

  if (message.includes('too common')) {
    return globalT('errors.auth.passwordTooCommon')
  }

  return globalT('errors.auth.passwordInvalid')
}

function translateAuthValidationDetail(detail) {
  const message = String(detail?.error || '')

  switch (detail?.field) {
    case 'login':
      return translateLoginValidation(message)
    case 'password':
      return translatePasswordValidation(message)
    default:
      return globalT('errors.auth.validationFailed')
  }
}

export function translateAuthValidationError(error) {
  if (!isValidationError(error)) {
    return ''
  }

  const messages = error.payload.details
    .map(translateAuthValidationDetail)
    .filter(Boolean)

  return [...new Set(messages)].join(' ') || globalT('errors.auth.validationFailed')
}

function normalizeLocale(locale) {
  return supportedLocales.includes(locale) ? locale : DEFAULT_LOCALE
}

function readStoredLocale() {
  if (typeof window === 'undefined') {
    return DEFAULT_LOCALE
  }

  try {
    return normalizeLocale(window.localStorage.getItem(LOCALE_STORAGE_KEY))
  } catch {
    return DEFAULT_LOCALE
  }
}

function updateDocumentLocale(locale) {
  if (typeof document === 'undefined') {
    return
  }

  document.documentElement.lang = normalizeLocale(locale)
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: readStoredLocale(),
  fallbackLocale: FALLBACK_LOCALE,
  messages: {
    ru,
    en,
  },
})

updateDocumentLocale(i18n.global.locale.value)

export function getCurrentLocale() {
  return normalizeLocale(i18n.global.locale.value)
}

export function setLocale(locale) {
  const nextLocale = normalizeLocale(locale)
  i18n.global.locale.value = nextLocale
  updateDocumentLocale(nextLocale)

  if (typeof window !== 'undefined') {
    try {
      window.localStorage.setItem(LOCALE_STORAGE_KEY, nextLocale)
    } catch {
      // Locale switching should still work if storage is unavailable.
    }
  }

  return nextLocale
}

export function globalT(key, params = {}) {
  return i18n.global.t(key, params)
}

export function hasTranslation(key) {
  return i18n.global.te(key)
}

export function resolveLocaleTag(locale = getCurrentLocale()) {
  return LOCALE_TAGS[normalizeLocale(locale)] || LOCALE_TAGS[DEFAULT_LOCALE]
}

function translateEnum(value, keyMap) {
  if (!value) {
    return globalT('common.placeholder')
  }

  const key = keyMap[value]
  return key ? globalT(key) : String(value)
}

export function translateCampaignStatus(status) {
  return translateEnum(status, CAMPAIGN_STATUS_KEYS)
}

export function translateTargetStatus(status) {
  return translateEnum(status, TARGET_STATUS_KEYS)
}

export function translateEventType(type) {
  return translateEnum(type, EVENT_TYPE_KEYS)
}

export function translateAttachmentVisibility(value) {
  return value
    ? globalT('attachments.visibility.public')
    : globalT('attachments.visibility.private')
}

export function translateKnownError(error, fallbackKey = 'errors.requestFailed') {
  const status = typeof error?.status === 'number' ? error.status : null
  const message = typeof error === 'string' ? error : error?.message

  if (status === 401 || status === 403) {
    return globalT('errors.auth.invalidCredentials')
  }

  if (status === 429) {
    return globalT('errors.auth.tooManyAttempts')
  }

  if (message && KNOWN_ERROR_KEYS.has(message)) {
    return globalT(KNOWN_ERROR_KEYS.get(message))
  }

  if (fallbackKey && hasTranslation(fallbackKey)) {
    return globalT(fallbackKey)
  }

  return fallbackKey || globalT('errors.requestFailed')
}

export default i18n

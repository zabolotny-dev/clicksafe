import { readonly, ref } from 'vue'
import { globalT } from '../i18n'

const notifications = ref([])
const timers = new Map()
let notificationSequence = 0

function dismissNotification(id) {
  if (timers.has(id)) {
    clearTimeout(timers.get(id))
    timers.delete(id)
  }

  notifications.value = notifications.value.filter((notification) => notification.id !== id)
}

function notify(notification) {
  const id = `notification-${notificationSequence + 1}`
  notificationSequence += 1

  const nextNotification = {
    id,
    variant: notification.variant || 'info',
    title: notification.title || globalT('common.notifications.defaultTitle'),
    message: notification.message || '',
    autoDismiss: notification.autoDismiss !== false,
    duration: Number.isFinite(notification.duration) ? notification.duration : 5200,
  }

  const nextNotifications = [nextNotification, ...notifications.value]
  nextNotifications.slice(5).forEach((droppedNotification) => {
    if (timers.has(droppedNotification.id)) {
      clearTimeout(timers.get(droppedNotification.id))
      timers.delete(droppedNotification.id)
    }
  })
  notifications.value = nextNotifications.slice(0, 5)

  if (nextNotification.autoDismiss && nextNotification.duration > 0) {
    timers.set(id, setTimeout(() => dismissNotification(id), nextNotification.duration))
  }

  return id
}

function clearNotifications() {
  timers.forEach((timer) => clearTimeout(timer))
  timers.clear()
  notifications.value = []
}

export function useNotifications() {
  return {
    notifications: readonly(notifications),
    notify,
    notifySuccess: (title, message, options = {}) => notify({ ...options, title, message, variant: 'success' }),
    notifyError: (title, message, options = {}) => notify({ ...options, title, message, variant: 'error', autoDismiss: options.autoDismiss ?? false }),
    notifyWarning: (title, message, options = {}) => notify({ ...options, title, message, variant: 'warning' }),
    notifyInfo: (title, message, options = {}) => notify({ ...options, title, message, variant: 'info' }),
    dismissNotification,
    clearNotifications,
  }
}

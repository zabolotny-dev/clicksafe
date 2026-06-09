import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../i18n', () => ({
  globalT: (key) => `[${key}]`,
}))

import { useNotifications } from './useNotifications'

describe('useNotifications', () => {
  let notifications, notify, notifySuccess, notifyError, notifyWarning, notifyInfo, dismissNotification, clearNotifications

  beforeEach(() => {
    const n = useNotifications()
    notifications = n.notifications
    notify = n.notify
    notifySuccess = n.notifySuccess
    notifyError = n.notifyError
    notifyWarning = n.notifyWarning
    notifyInfo = n.notifyInfo
    dismissNotification = n.dismissNotification
    clearNotifications = n.clearNotifications
    clearNotifications()
  })

  it('notify adds a notification', () => {
    notify({ title: 'Hello', message: 'World', variant: 'info' })
    expect(notifications.value).toHaveLength(1)
    expect(notifications.value[0].title).toBe('Hello')
    expect(notifications.value[0].message).toBe('World')
    expect(notifications.value[0].variant).toBe('info')
  })

  it('notify assigns unique incrementing ids', () => {
    const id1 = notify({ title: 'First' })
    const id2 = notify({ title: 'Second' })
    expect(id1).not.toBe(id2)
    expect(typeof id1).toBe('string')
  })

  it('notifications are prepended (newest first)', () => {
    notify({ title: 'First' })
    notify({ title: 'Second' })
    expect(notifications.value[0].title).toBe('Second')
    expect(notifications.value[1].title).toBe('First')
  })

  it('caps the notification list at 5', () => {
    for (let i = 0; i < 7; i++) {
      notify({ title: `Notification ${i}` })
    }
    expect(notifications.value).toHaveLength(5)
    expect(notifications.value[0].title).toBe('Notification 6')
  })

  it('uses "info" as default variant', () => {
    notify({ title: 'Default' })
    expect(notifications.value[0].variant).toBe('info')
  })

  it('uses default title from i18n when title is omitted', () => {
    notify({ message: 'No title' })
    expect(notifications.value[0].title).toBe('[common.notifications.defaultTitle]')
  })

  it('autoDismiss defaults to true with 5200ms duration', () => {
    notify({ title: 'Auto' })
    expect(notifications.value[0].autoDismiss).toBe(true)
    expect(notifications.value[0].duration).toBe(5200)
  })

  it('autoDismiss can be disabled', () => {
    notify({ title: 'Sticky', autoDismiss: false })
    expect(notifications.value[0].autoDismiss).toBe(false)
  })

  it('dismissNotification removes the notification by id', () => {
    const id = notify({ title: 'To remove' })
    notify({ title: 'Keep' })
    dismissNotification(id)
    expect(notifications.value).toHaveLength(1)
    expect(notifications.value[0].title).toBe('Keep')
  })

  it('dismissNotification is a no-op for unknown id', () => {
    notify({ title: 'Existing' })
    dismissNotification('nonexistent-id')
    expect(notifications.value).toHaveLength(1)
  })

  it('clearNotifications removes all', () => {
    notify({ title: 'A' })
    notify({ title: 'B' })
    clearNotifications()
    expect(notifications.value).toHaveLength(0)
  })

  it('notifySuccess sets success variant', () => {
    notifySuccess('Title', 'Body')
    expect(notifications.value[0].variant).toBe('success')
    expect(notifications.value[0].title).toBe('Title')
    expect(notifications.value[0].message).toBe('Body')
  })

  it('notifyError sets error variant and disables autoDismiss by default', () => {
    notifyError('Error', 'Something went wrong')
    expect(notifications.value[0].variant).toBe('error')
    expect(notifications.value[0].autoDismiss).toBe(false)
  })

  it('notifyWarning sets warning variant', () => {
    notifyWarning('Warn', 'Be careful')
    expect(notifications.value[0].variant).toBe('warning')
  })

  it('notifyInfo sets info variant', () => {
    notifyInfo('Info', 'Just so you know')
    expect(notifications.value[0].variant).toBe('info')
  })

  it('auto-dismisses notification after duration with fake timers', () => {
    vi.useFakeTimers()
    notify({ title: 'Timed', autoDismiss: true, duration: 1000 })
    expect(notifications.value).toHaveLength(1)
    vi.advanceTimersByTime(1000)
    expect(notifications.value).toHaveLength(0)
    vi.useRealTimers()
  })

  it('does not auto-dismiss when duration is 0', () => {
    vi.useFakeTimers()
    notify({ title: 'No timer', autoDismiss: true, duration: 0 })
    vi.advanceTimersByTime(10000)
    expect(notifications.value).toHaveLength(1)
    vi.useRealTimers()
  })
})

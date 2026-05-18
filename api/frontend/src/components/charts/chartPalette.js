const fallbackTokens = {
  primary: '#422082',
  primarySoft: '#f1effb',
  accentViolet: '#6a5fc1',
  accentLime: '#c2ef4e',
  accentPink: '#fa7faa',
  accentIndigo: '#533afd',
  surface: '#ffffff',
  ink: '#1f1633',
  inkSecondary: '#3b3348',
  inkMuted: '#6b6475',
  inkSubtle: '#8a8495',
  hairline: '#e5e7eb',
  hairlineCool: '#d7d4de',
  hairlineStrong: '#cfc8db',
  success: '#0f8a3b',
  warning: '#a86800',
  danger: '#c92a3a',
  info: '#254fad',
  canvasMuted: '#f8fafc',
}

const tokenNames = {
  primary: '--color-primary',
  primarySoft: '--color-primary-soft',
  accentViolet: '--color-accent-violet',
  accentLime: '--color-accent-lime',
  accentPink: '--color-accent-pink',
  accentIndigo: '--color-accent-indigo',
  surface: '--color-surface',
  ink: '--color-ink',
  inkSecondary: '--color-ink-secondary',
  inkMuted: '--color-ink-muted',
  inkSubtle: '--color-ink-subtle',
  hairline: '--color-hairline',
  hairlineCool: '--color-hairline-cool',
  hairlineStrong: '--color-hairline-strong',
  success: '--color-success',
  warning: '--color-warning',
  danger: '--color-danger',
  info: '--color-info',
  canvasMuted: '--color-canvas-muted',
}

function readCssToken(tokenName, fallback) {
  if (typeof window === 'undefined' || typeof document === 'undefined') {
    return fallback
  }

  return window
    .getComputedStyle(document.documentElement)
    .getPropertyValue(tokenName)
    .trim() || fallback
}

export function chartTokens() {
  return Object.fromEntries(
    Object.entries(tokenNames).map(([key, tokenName]) => [
      key,
      readCssToken(tokenName, fallbackTokens[key]),
    ]),
  )
}

export function eventChartColor(eventType, tokens = chartTokens()) {
  const colors = {
    MESSAGE_SENT: tokens.primary,
    DELIVERY_FAILED: tokens.accentPink,
    EMAIL_OPENED: tokens.accentViolet,
    LINK_OPENED: tokens.warning,
    DATA_SENT: tokens.danger,
  }

  return colors[eventType] || tokens.inkMuted
}

export function riskValueColor(value, tokens = chartTokens()) {
  if (value >= 75) {
    return tokens.danger
  }

  if (value >= 45) {
    return tokens.warning
  }

  return tokens.primary
}

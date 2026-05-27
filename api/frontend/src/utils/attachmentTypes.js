export const AUDIO_ATTACHMENT_TYPES = ['.aac', '.flac', '.m4a', '.mp3', '.ogg', '.opus', '.wav', '.weba']
export const SUPPORTED_ATTACHMENT_TYPES = ['.aac', '.docx', '.flac', '.gif', '.html', '.jpeg', '.jpg', '.m4a', '.mp3', '.ogg', '.opus', '.png', '.pptx', '.txt', '.wav', '.weba', '.webp', '.xlsx']
export const IMAGE_ATTACHMENT_TYPES = ['.gif', '.jpeg', '.jpg', '.png', '.webp']
export const TEXT_ATTACHMENT_TYPES = ['.html', '.txt']
export const DOWNLOAD_ONLY_ATTACHMENT_TYPES = ['.docx', '.pptx', '.xlsx']

const AUDIO_TYPE_SET = new Set([
  ...AUDIO_ATTACHMENT_TYPES,
  'audio/aac',
  'audio/flac',
  'audio/mp4',
  'audio/mpeg',
  'audio/ogg',
  'audio/ogg; codecs=opus',
  'audio/wav',
  'audio/webm',
  'audio/x-m4a',
  'audio/x-wav',
])
const IMAGE_TYPE_SET = new Set([
  ...IMAGE_ATTACHMENT_TYPES,
  'image/gif',
  'image/jpeg',
  'image/png',
  'image/webp',
])
const HTML_TYPE_SET = new Set(['.html', 'html', 'text/html'])
const TEXT_TYPE_SET = new Set(['.txt', 'txt', 'text/plain'])
const DOWNLOAD_ONLY_TYPE_SET = new Set(DOWNLOAD_ONLY_ATTACHMENT_TYPES)

export function normalizeAttachmentType(type) {
  return String(type || '').trim().toLowerCase()
}

export function isImageAttachmentType(type) {
  return IMAGE_TYPE_SET.has(normalizeAttachmentType(type))
}

export function isAudioAttachmentType(type) {
  return AUDIO_TYPE_SET.has(normalizeAttachmentType(type))
}

export function isHtmlAttachmentType(type) {
  return HTML_TYPE_SET.has(normalizeAttachmentType(type))
}

export function isTextAttachmentType(type) {
  return TEXT_TYPE_SET.has(normalizeAttachmentType(type))
}

export function isTextPreviewAttachmentType(type) {
  return isHtmlAttachmentType(type) || isTextAttachmentType(type)
}

export function isDownloadOnlyAttachmentType(type) {
  return DOWNLOAD_ONLY_TYPE_SET.has(normalizeAttachmentType(type))
}

export function isPreviewableAttachmentType(type) {
  return isImageAttachmentType(type) || isAudioAttachmentType(type) || isTextPreviewAttachmentType(type)
}

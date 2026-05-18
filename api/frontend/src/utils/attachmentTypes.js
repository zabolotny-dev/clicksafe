export const SUPPORTED_ATTACHMENT_TYPES = ['.docx', '.gif', '.html', '.jpeg', '.jpg', '.png', '.pptx', '.txt', '.webp', '.xlsx']
export const IMAGE_ATTACHMENT_TYPES = ['.gif', '.jpeg', '.jpg', '.png', '.webp']
export const TEXT_ATTACHMENT_TYPES = ['.html', '.txt']
export const DOWNLOAD_ONLY_ATTACHMENT_TYPES = ['.docx', '.pptx', '.xlsx']

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
  return isImageAttachmentType(type) || isTextPreviewAttachmentType(type)
}

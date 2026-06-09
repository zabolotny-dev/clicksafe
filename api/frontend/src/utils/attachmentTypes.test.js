import { describe, it, expect } from 'vitest'
import {
  normalizeAttachmentType,
  isImageAttachmentType,
  isAudioAttachmentType,
  isHtmlAttachmentType,
  isTextAttachmentType,
  isTextPreviewAttachmentType,
  isDownloadOnlyAttachmentType,
  isPreviewableAttachmentType,
} from './attachmentTypes'

describe('normalizeAttachmentType', () => {
  it('lowercases and trims', () => {
    expect(normalizeAttachmentType('  .PNG  ')).toBe('.png')
    expect(normalizeAttachmentType('IMAGE/JPEG')).toBe('image/jpeg')
  })

  it('returns empty string for falsy input', () => {
    expect(normalizeAttachmentType(null)).toBe('')
    expect(normalizeAttachmentType(undefined)).toBe('')
    expect(normalizeAttachmentType('')).toBe('')
  })
})

describe('isImageAttachmentType', () => {
  it('returns true for image extensions', () => {
    expect(isImageAttachmentType('.jpg')).toBe(true)
    expect(isImageAttachmentType('.jpeg')).toBe(true)
    expect(isImageAttachmentType('.png')).toBe(true)
    expect(isImageAttachmentType('.gif')).toBe(true)
    expect(isImageAttachmentType('.webp')).toBe(true)
  })

  it('returns true for image mime types', () => {
    expect(isImageAttachmentType('image/jpeg')).toBe(true)
    expect(isImageAttachmentType('image/png')).toBe(true)
  })

  it('returns true case-insensitively', () => {
    expect(isImageAttachmentType('.JPG')).toBe(true)
    expect(isImageAttachmentType('IMAGE/PNG')).toBe(true)
  })

  it('returns false for non-image types', () => {
    expect(isImageAttachmentType('.mp3')).toBe(false)
    expect(isImageAttachmentType('.html')).toBe(false)
    expect(isImageAttachmentType('')).toBe(false)
  })
})

describe('isAudioAttachmentType', () => {
  it('returns true for audio extensions', () => {
    expect(isAudioAttachmentType('.mp3')).toBe(true)
    expect(isAudioAttachmentType('.wav')).toBe(true)
    expect(isAudioAttachmentType('.ogg')).toBe(true)
    expect(isAudioAttachmentType('.opus')).toBe(true)
    expect(isAudioAttachmentType('.aac')).toBe(true)
    expect(isAudioAttachmentType('.flac')).toBe(true)
    expect(isAudioAttachmentType('.m4a')).toBe(true)
  })

  it('returns true for audio mime types', () => {
    expect(isAudioAttachmentType('audio/mpeg')).toBe(true)
    expect(isAudioAttachmentType('audio/wav')).toBe(true)
  })

  it('returns false for non-audio types', () => {
    expect(isAudioAttachmentType('.jpg')).toBe(false)
    expect(isAudioAttachmentType('.html')).toBe(false)
  })
})

describe('isHtmlAttachmentType', () => {
  it('returns true for html extension and mime', () => {
    expect(isHtmlAttachmentType('.html')).toBe(true)
    expect(isHtmlAttachmentType('html')).toBe(true)
    expect(isHtmlAttachmentType('text/html')).toBe(true)
  })

  it('returns false for other types', () => {
    expect(isHtmlAttachmentType('.txt')).toBe(false)
    expect(isHtmlAttachmentType('.png')).toBe(false)
  })
})

describe('isTextAttachmentType', () => {
  it('returns true for txt extension and mime', () => {
    expect(isTextAttachmentType('.txt')).toBe(true)
    expect(isTextAttachmentType('txt')).toBe(true)
    expect(isTextAttachmentType('text/plain')).toBe(true)
  })

  it('returns false for other types', () => {
    expect(isTextAttachmentType('.html')).toBe(false)
    expect(isTextAttachmentType('.mp3')).toBe(false)
  })
})

describe('isTextPreviewAttachmentType', () => {
  it('returns true for html and txt', () => {
    expect(isTextPreviewAttachmentType('.html')).toBe(true)
    expect(isTextPreviewAttachmentType('.txt')).toBe(true)
    expect(isTextPreviewAttachmentType('text/html')).toBe(true)
    expect(isTextPreviewAttachmentType('text/plain')).toBe(true)
  })

  it('returns false for images and audio', () => {
    expect(isTextPreviewAttachmentType('.jpg')).toBe(false)
    expect(isTextPreviewAttachmentType('.mp3')).toBe(false)
  })
})

describe('isDownloadOnlyAttachmentType', () => {
  it('returns true for office formats', () => {
    expect(isDownloadOnlyAttachmentType('.docx')).toBe(true)
    expect(isDownloadOnlyAttachmentType('.pptx')).toBe(true)
    expect(isDownloadOnlyAttachmentType('.xlsx')).toBe(true)
  })

  it('returns false for previewable types', () => {
    expect(isDownloadOnlyAttachmentType('.jpg')).toBe(false)
    expect(isDownloadOnlyAttachmentType('.html')).toBe(false)
    expect(isDownloadOnlyAttachmentType('.mp3')).toBe(false)
  })
})

describe('isPreviewableAttachmentType', () => {
  it('returns true for image, audio, and text types', () => {
    expect(isPreviewableAttachmentType('.jpg')).toBe(true)
    expect(isPreviewableAttachmentType('.mp3')).toBe(true)
    expect(isPreviewableAttachmentType('.html')).toBe(true)
    expect(isPreviewableAttachmentType('.txt')).toBe(true)
  })

  it('returns false for download-only types', () => {
    expect(isPreviewableAttachmentType('.docx')).toBe(false)
    expect(isPreviewableAttachmentType('.pptx')).toBe(false)
    expect(isPreviewableAttachmentType('.xlsx')).toBe(false)
  })

  it('returns false for unknown types', () => {
    expect(isPreviewableAttachmentType('.xyz')).toBe(false)
    expect(isPreviewableAttachmentType('')).toBe(false)
  })
})

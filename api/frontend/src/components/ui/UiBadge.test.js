import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import UiBadge from './UiBadge.vue'

describe('UiBadge', () => {
  it('renders slot content', () => {
    const wrapper = mount(UiBadge, { slots: { default: 'Active' } })
    expect(wrapper.text()).toBe('Active')
  })

  it('applies default variant class "neutral"', () => {
    const wrapper = mount(UiBadge)
    expect(wrapper.classes()).toContain('ui-badge--neutral')
  })

  it.each(['neutral', 'info', 'success', 'warning', 'danger', 'primary'])(
    'applies class for variant "%s"',
    (variant) => {
      const wrapper = mount(UiBadge, { props: { variant } })
      expect(wrapper.classes()).toContain(`ui-badge--${variant}`)
    },
  )

  it('always has base class "ui-badge"', () => {
    const wrapper = mount(UiBadge, { props: { variant: 'success' } })
    expect(wrapper.classes()).toContain('ui-badge')
  })
})

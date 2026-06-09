import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import UiButton from './UiButton.vue'

describe('UiButton', () => {
  it('renders slot content', () => {
    const wrapper = mount(UiButton, { slots: { default: 'Click me' } })
    expect(wrapper.text()).toContain('Click me')
  })

  it('uses "button" as default type', () => {
    const wrapper = mount(UiButton)
    expect(wrapper.attributes('type')).toBe('button')
  })

  it('applies primary variant class by default', () => {
    const wrapper = mount(UiButton)
    expect(wrapper.classes()).toContain('ui-button--primary')
  })

  it.each(['primary', 'secondary', 'ghost', 'danger'])('applies class for variant "%s"', (variant) => {
    const wrapper = mount(UiButton, { props: { variant } })
    expect(wrapper.classes()).toContain(`ui-button--${variant}`)
  })

  it('is not disabled by default', () => {
    const wrapper = mount(UiButton)
    expect(wrapper.attributes('disabled')).toBeUndefined()
  })

  it('is disabled when disabled prop is true', () => {
    const wrapper = mount(UiButton, { props: { disabled: true } })
    expect(wrapper.element.disabled).toBe(true)
  })

  it('shows spinner and disables button when loading', () => {
    const wrapper = mount(UiButton, { props: { loading: true } })
    expect(wrapper.find('.button-spinner').exists()).toBe(true)
    expect(wrapper.element.disabled).toBe(true)
    expect(wrapper.classes()).toContain('is-loading')
    expect(wrapper.attributes('aria-busy')).toBe('true')
  })

  it('has no aria-busy when not loading', () => {
    const wrapper = mount(UiButton)
    expect(wrapper.attributes('aria-busy')).toBeUndefined()
  })

  it('does not show spinner when not loading', () => {
    const wrapper = mount(UiButton)
    expect(wrapper.find('.button-spinner').exists()).toBe(false)
  })

  it('is disabled when both disabled and loading are true', () => {
    const wrapper = mount(UiButton, { props: { disabled: true, loading: true } })
    expect(wrapper.element.disabled).toBe(true)
  })

  it('passes type prop to the button element', () => {
    const wrapper = mount(UiButton, { props: { type: 'submit' } })
    expect(wrapper.attributes('type')).toBe('submit')
  })
})

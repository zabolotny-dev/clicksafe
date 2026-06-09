import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import EmptyState from './EmptyState.vue'

describe('EmptyState', () => {
  it('renders title', () => {
    const wrapper = mount(EmptyState, { props: { title: 'No data found' } })
    expect(wrapper.find('h2').text()).toBe('No data found')
  })

  it('renders description when provided', () => {
    const wrapper = mount(EmptyState, {
      props: { title: 'Empty', description: 'Try adding some items' },
    })
    expect(wrapper.find('p').text()).toBe('Try adding some items')
  })

  it('does not render description element when not provided', () => {
    const wrapper = mount(EmptyState, {
      props: { title: 'Empty' },
    })
    expect(wrapper.find('p').exists()).toBe(false)
  })

  it('renders default slot content', () => {
    const wrapper = mount(EmptyState, {
      props: { title: 'Nothing here' },
      slots: { default: '<button>Add item</button>' },
    })
    expect(wrapper.find('.empty-state-actions').exists()).toBe(true)
    expect(wrapper.find('button').text()).toBe('Add item')
  })

  it('does not render actions container when no slot provided', () => {
    const wrapper = mount(EmptyState, { props: { title: 'Nothing' } })
    expect(wrapper.find('.empty-state-actions').exists()).toBe(false)
  })

  it('renders the decorative mark element', () => {
    const wrapper = mount(EmptyState, { props: { title: 'Nothing' } })
    expect(wrapper.find('.empty-state-mark').exists()).toBe(true)
  })
})

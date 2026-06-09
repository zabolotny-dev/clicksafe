import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import MetricCard from './MetricCard.vue'

describe('MetricCard', () => {
  it('renders label and value', () => {
    const wrapper = mount(MetricCard, {
      props: { label: 'Open Rate', value: '42%' },
    })
    expect(wrapper.find('.metric-label').text()).toBe('Open Rate')
    expect(wrapper.find('.metric-value').text()).toBe('42%')
  })

  it('shows default value "—" when value is not passed', () => {
    const wrapper = mount(MetricCard, {
      props: { label: 'Metric' },
    })
    expect(wrapper.find('.metric-value').text()).toBe('—')
  })

  it('renders description when provided', () => {
    const wrapper = mount(MetricCard, {
      props: { label: 'Clicks', value: '10', description: 'Last 30 days' },
    })
    expect(wrapper.find('.metric-description').text()).toBe('Last 30 days')
  })

  it('does not render description element when not provided', () => {
    const wrapper = mount(MetricCard, {
      props: { label: 'Clicks', value: '10' },
    })
    expect(wrapper.find('.metric-description').exists()).toBe(false)
  })

  it('renders icon when provided', () => {
    const FakeIcon = { template: '<svg data-testid="icon" />' }
    const wrapper = mount(MetricCard, {
      props: { label: 'Risk', value: '55', icon: FakeIcon },
    })
    expect(wrapper.find('.metric-icon').exists()).toBe(true)
  })

  it('does not render icon element when icon is null', () => {
    const wrapper = mount(MetricCard, {
      props: { label: 'Risk', value: '55', icon: null },
    })
    expect(wrapper.find('.metric-icon').exists()).toBe(false)
  })

  it('renders label-actions slot', () => {
    const wrapper = mount(MetricCard, {
      props: { label: 'Metric' },
      slots: { 'label-actions': '<button>Filter</button>' },
    })
    expect(wrapper.find('button').text()).toBe('Filter')
  })
})

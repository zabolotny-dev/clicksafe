import {
  BarChart3,
  Building2,
  Flag,
  Files,
  LayoutDashboard,
  Mail,
  Monitor,
  Settings,
  Users,
} from 'lucide-vue-next'

export const navigationGroups = [
  {
    labelKey: '',
    items: [
      {
        labelKey: 'nav.dashboard',
        to: '/dashboard',
        icon: LayoutDashboard,
      },
      {
        labelKey: 'nav.campaigns',
        to: '/campaigns',
        icon: Flag,
      },
    ],
  },
  {
    labelKey: 'nav.templates',
    items: [
      {
        labelKey: 'nav.emailTemplates',
        to: '/templates/email',
        icon: Mail,
      },
      {
        labelKey: 'nav.landingPages',
        to: '/templates/landing',
        icon: Monitor,
      },
      {
        labelKey: 'nav.attachments',
        to: '/templates/attachments',
        icon: Files,
      },
    ],
  },
  {
    labelKey: 'nav.people',
    items: [
      {
        labelKey: 'nav.employees',
        to: '/people/employees',
        icon: Users,
      },
      {
        labelKey: 'nav.departments',
        to: '/people/departments',
        icon: Building2,
      },
    ],
  },
  {
    labelKey: '',
    items: [
      {
        labelKey: 'nav.reports',
        to: '/reports',
        icon: BarChart3,
      },
      {
        labelKey: 'nav.organizationSettings',
        to: '/organization-settings',
        icon: Settings,
      },
    ],
  },
]

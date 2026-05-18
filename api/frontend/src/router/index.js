import { createRouter, createWebHistory } from 'vue-router'
import { useSession } from '../composables/useSession'
import AdminShell from '../layouts/AdminShell.vue'
import AttachmentsPage from '../pages/AttachmentsPage.vue'
import CampaignDetailPage from '../pages/CampaignDetailPage.vue'
import CampaignWizardPage from '../pages/CampaignWizardPage.vue'
import CampaignsPage from '../pages/CampaignsPage.vue'
import DashboardPage from '../pages/DashboardPage.vue'
import DepartmentsPage from '../pages/DepartmentsPage.vue'
import EmailTemplatesPage from '../pages/EmailTemplatesPage.vue'
import EmployeesPage from '../pages/EmployeesPage.vue'
import LandingPagesPage from '../pages/LandingPagesPage.vue'
import LoginPage from '../pages/LoginPage.vue'
import OrganizationSettingsPage from '../pages/OrganizationSettingsPage.vue'
import PlaceholderPage from '../pages/PlaceholderPage.vue'
import ReportsPage from '../pages/ReportsPage.vue'
import TargetsEventsPage from '../pages/TargetsEventsPage.vue'

const adminRoutes = [
  {
    path: '/dashboard',
    name: 'dashboard',
    meta: {
      titleKey: 'routes.dashboard.title',
      descriptionKey: 'routes.dashboard.description',
      sectionKey: 'sections.operations',
      requiresAuth: true,
    },
  },
  {
    path: '/campaigns',
    name: 'campaigns',
    meta: {
      titleKey: 'routes.campaigns.title',
      descriptionKey: 'routes.campaigns.description',
      sectionKey: 'sections.operations',
      requiresAuth: true,
    },
  },
  {
    path: '/campaigns/new',
    name: 'campaigns-new',
    meta: {
      titleKey: 'routes.campaignCreate.title',
      descriptionKey: 'routes.campaignCreate.description',
      sectionKey: 'sections.campaigns',
      requiresAuth: true,
    },
  },
  {
    path: '/campaigns/:id/edit',
    name: 'campaign-edit',
    meta: {
      titleKey: 'routes.campaignEdit.title',
      descriptionKey: 'routes.campaignEdit.description',
      sectionKey: 'sections.campaigns',
      requiresAuth: true,
    },
  },
  {
    path: '/campaigns/:id',
    name: 'campaign-detail',
    meta: {
      titleKey: 'routes.campaignDetail.title',
      descriptionKey: 'routes.campaignDetail.description',
      sectionKey: 'sections.campaigns',
      requiresAuth: true,
    },
  },
  {
    path: '/targets-events',
    name: 'targets-events',
    meta: {
      titleKey: 'routes.targetsEvents.title',
      descriptionKey: 'routes.targetsEvents.description',
      sectionKey: 'sections.operations',
      requiresAuth: true,
    },
  },
  {
    path: '/templates/email',
    name: 'email-templates',
    meta: {
      titleKey: 'routes.emailTemplates.title',
      descriptionKey: 'routes.emailTemplates.description',
      sectionKey: 'sections.templates',
      requiresAuth: true,
    },
  },
  {
    path: '/templates/email/new',
    name: 'email-template-new',
    meta: {
      titleKey: 'routes.emailTemplateCreate.title',
      descriptionKey: 'routes.emailTemplateCreate.description',
      sectionKey: 'sections.templates',
      requiresAuth: true,
    },
  },
  {
    path: '/templates/email/:id/edit',
    name: 'email-template-edit',
    meta: {
      titleKey: 'routes.emailTemplateEdit.title',
      descriptionKey: 'routes.emailTemplateEdit.description',
      sectionKey: 'sections.templates',
      requiresAuth: true,
    },
  },
  {
    path: '/templates/landing',
    name: 'landing-pages',
    meta: {
      titleKey: 'routes.landingPages.title',
      descriptionKey: 'routes.landingPages.description',
      sectionKey: 'sections.templates',
      requiresAuth: true,
    },
  },
  {
    path: '/templates/landing/new',
    name: 'landing-page-new',
    meta: {
      titleKey: 'routes.landingPageCreate.title',
      descriptionKey: 'routes.landingPageCreate.description',
      sectionKey: 'sections.templates',
      requiresAuth: true,
    },
  },
  {
    path: '/templates/landing/:id/edit',
    name: 'landing-page-edit',
    meta: {
      titleKey: 'routes.landingPageEdit.title',
      descriptionKey: 'routes.landingPageEdit.description',
      sectionKey: 'sections.templates',
      requiresAuth: true,
    },
  },
  {
    path: '/templates/attachments',
    name: 'attachments',
    meta: {
      titleKey: 'routes.attachments.title',
      descriptionKey: 'routes.attachments.description',
      sectionKey: 'sections.templates',
      requiresAuth: true,
    },
  },
  {
    path: '/people/employees',
    name: 'employees',
    meta: {
      titleKey: 'routes.employees.title',
      descriptionKey: 'routes.employees.description',
      sectionKey: 'sections.people',
      requiresAuth: true,
    },
  },
  {
    path: '/people/employees/new',
    name: 'employees-new',
    meta: {
      titleKey: 'routes.employeeCreate.title',
      descriptionKey: 'routes.employeeCreate.description',
      sectionKey: 'sections.people',
      requiresAuth: true,
    },
  },
  {
    path: '/people/departments',
    name: 'departments',
    meta: {
      titleKey: 'routes.departments.title',
      descriptionKey: 'routes.departments.description',
      sectionKey: 'sections.people',
      requiresAuth: true,
    },
  },
  {
    path: '/people/departments/new',
    name: 'departments-new',
    meta: {
      titleKey: 'routes.departmentCreate.title',
      descriptionKey: 'routes.departmentCreate.description',
      sectionKey: 'sections.people',
      requiresAuth: true,
    },
  },
  {
    path: '/reports',
    name: 'reports',
    meta: {
      titleKey: 'routes.reports.title',
      descriptionKey: 'routes.reports.description',
      sectionKey: 'sections.analytics',
      requiresAuth: true,
    },
  },
  {
    path: '/organization-settings',
    name: 'organization-settings',
    meta: {
      titleKey: 'routes.organizationSettings.title',
      descriptionKey: 'routes.organizationSettings.description',
      sectionKey: 'sections.settings',
      requiresAuth: true,
    },
  },
]

const routes = [
  {
    path: '/login',
    name: 'login',
    component: LoginPage,
    meta: {
      guestOnly: true,
    },
  },
  {
    path: '/',
    component: AdminShell,
    children: [
      {
        path: '',
        redirect: '/dashboard',
      },
      ...adminRoutes.map((route) => ({
        ...route,
        component: route.name === 'dashboard'
          ? DashboardPage
          : route.name === 'campaigns'
            ? CampaignsPage
            : route.name === 'campaign-detail'
              ? CampaignDetailPage
              : ['campaigns-new', 'campaign-edit'].includes(route.name)
                ? CampaignWizardPage
                : route.name === 'targets-events'
                  ? TargetsEventsPage
                  : ['email-templates', 'email-template-new', 'email-template-edit'].includes(route.name)
                    ? EmailTemplatesPage
                    : ['landing-pages', 'landing-page-new', 'landing-page-edit'].includes(route.name)
                      ? LandingPagesPage
                      : route.name === 'attachments'
                        ? AttachmentsPage
                        : ['employees', 'employees-new'].includes(route.name)
                          ? EmployeesPage
                          : ['departments', 'departments-new'].includes(route.name)
                            ? DepartmentsPage
                            : route.name === 'reports'
                              ? ReportsPage
                              : route.name === 'organization-settings'
                                ? OrganizationSettingsPage
                                : PlaceholderPage,
        props: {
          titleKey: route.meta.titleKey,
          descriptionKey: route.meta.descriptionKey,
          sectionKey: route.meta.sectionKey,
        },
      })),
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const session = useSession()

  if (!session.isReady.value) {
    await session.loadCurrentSession()
  }

  if (to.meta.guestOnly && session.isAuthenticated.value) {
    return { name: 'dashboard' }
  }

  if (!to.meta.requiresAuth) {
    return true
  }

  if (!session.isAuthenticated.value) {
    return {
      name: 'login',
      query: {
        redirect: to.fullPath,
      },
    }
  }

  return true
})

export default router

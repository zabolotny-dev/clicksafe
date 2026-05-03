import {
  Connection,
  Files,
  Message as MessageIcon,
  OfficeBuilding,
  Promotion,
  User,
} from '@element-plus/icons-vue'

export const endpointGroups = [
  {
    name: 'Organization',
    icon: OfficeBuilding,
    routes: [
      ['POST', '/organization', 'создать организацию'],
      ['GET', '/organization', 'получить организацию'],
      ['PUT', '/organization', 'обновить организацию'],
      ['PUT', '/organization/logo', 'загрузить логотип multipart file'],
      ['DELETE', '/organization/logo', 'удалить логотип'],
    ],
  },
  {
    name: 'Department',
    icon: Files,
    routes: [
      ['GET', '/department', 'список с page, rows, label, orderBy'],
      ['POST', '/department', 'создать отдел'],
      ['GET', '/department/:id', 'получить отдел'],
      ['PUT', '/department/:id', 'обновить отдел'],
      ['DELETE', '/department/:id', 'удалить отдел'],
    ],
  },
  {
    name: 'Employee',
    icon: User,
    routes: [
      ['GET', '/employee', 'список с фильтрами full_name, email, department_id'],
      ['POST', '/employee', 'создать сотрудника'],
      ['GET', '/employee/:id', 'получить сотрудника'],
      ['PUT', '/employee/:id', 'обновить сотрудника'],
      ['DELETE', '/employee/:id', 'удалить сотрудника'],
    ],
  },
  {
    name: 'Messageapp',
    icon: MessageIcon,
    routes: [
      ['GET', '/message', 'список шаблонов'],
      ['POST', '/message', 'создать метаданные шаблона'],
      ['PUT', '/message/:id/content', 'загрузить HTML multipart file'],
      ['GET', '/message/:id/content', 'получить исходный HTML'],
      ['POST', '/message/:id/render', 'отрендерить HTML по employee_id'],
      ['GET', '/message/:id', 'получить шаблон'],
      ['PUT', '/message/:id', 'обновить метаданные'],
      ['DELETE', '/message/:id', 'удалить шаблон'],
    ],
  },
  {
    name: 'Events',
    icon: Promotion,
    routes: [['POST', '/events', 'зафиксировать событие campaign_id, employee_id, type']],
  },
]

export const navItems = [
  { key: 'messages', label: 'Messageapp', icon: MessageIcon },
  { key: 'directory', label: 'Данные', icon: User },
  { key: 'organization', label: 'Организация', icon: OfficeBuilding },
  { key: 'events', label: 'События', icon: Promotion },
  { key: 'routes', label: 'Роуты', icon: Connection },
]

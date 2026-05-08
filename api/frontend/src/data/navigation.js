import {
  Connection,
  Files,
  Flag,
  Message as MessageIcon,
  Monitor,
  OfficeBuilding,
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
      ['GET', '/employee', 'список с full_name, email, department_id'],
      ['POST', '/employee', 'создать сотрудника'],
      ['GET', '/employee/:id', 'получить сотрудника'],
      ['PUT', '/employee/:id', 'обновить сотрудника'],
      ['DELETE', '/employee/:id', 'удалить сотрудника'],
    ],
  },
  {
    name: 'Message',
    icon: MessageIcon,
    routes: [
      ['GET', '/message', 'список писем'],
      ['POST', '/message', 'создать письмо'],
      ['PUT', '/message/:id/content', 'загрузить HTML multipart file'],
      ['GET', '/message/:id/content', 'получить исходный HTML'],
      ['POST', '/message/:id/render', 'отрендерить HTML по target_id'],
      ['GET', '/message/:id', 'получить письмо'],
      ['PUT', '/message/:id', 'обновить письмо'],
      ['DELETE', '/message/:id', 'удалить письмо'],
    ],
  },
  {
    name: 'Landing',
    icon: Monitor,
    routes: [
      ['GET', '/landing', 'список лендингов'],
      ['POST', '/landing', 'создать лендинг'],
      ['PUT', '/landing/:id/content', 'загрузить HTML multipart file'],
      ['GET', '/landing/:id/content', 'получить исходный HTML'],
      ['POST', '/landing/:id/render', 'отрендерить HTML по target_id'],
      ['GET', '/landing/:id', 'получить лендинг'],
      ['PUT', '/landing/:id', 'обновить лендинг'],
      ['DELETE', '/landing/:id', 'удалить лендинг'],
    ],
  },
  {
    name: 'Campaign / Target / Visit',
    icon: Flag,
    routes: [
      ['GET', '/campaign', 'список кампаний'],
      ['POST', '/campaign', 'создать кампанию'],
      ['PUT', '/campaign/:id', 'обновить кампанию'],
      ['PUT', '/campaign/:id/start', 'запустить кампанию'],
      ['PUT', '/campaign/:id/pause', 'поставить на паузу'],
      ['PUT', '/campaign/:id/cancel', 'отменить кампанию'],
      ['DELETE', '/campaign/:id', 'удалить кампанию'],
      ['POST', '/target', 'добавить сотрудника в кампанию'],
      ['PUT', '/target/:id/schedule', 'назначить расписание target'],
      ['PUT', '/target/campaign/:campaign_id/distribute', 'автораспределить target-ы'],
      ['DELETE', '/target/:id', 'удалить target'],
      ['DELETE', '/target/campaign/:campaign_id', 'удалить target-ы кампании'],
      ['GET', '/vtarget', 'витрина target-ов с сотрудником и событиями'],
      ['GET', '/:token', 'публичный visitapp для landing'],
    ],
  },
]

export const navItems = [
  { key: 'campaigns', label: 'Кампании', icon: Flag },
  { key: 'messages', label: 'Письма', icon: MessageIcon },
  { key: 'landings', label: 'Лендинги', icon: Monitor },
  { key: 'directory', label: 'Сотрудники', icon: User },
  { key: 'organization', label: 'Организация', icon: OfficeBuilding },
  { key: 'routes', label: 'Роуты', icon: Connection },
]

import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import { createApiClient, getStoredApiBase, storeApiBase } from '../api'
import { defaultLandingTemplate, defaultMessageTemplate } from '../data/defaultTemplate'

const campaignStatuses = ['DRAFT', 'ACTIVE', 'PAUSED', 'COMPLETED', 'CANCELED']
const targetStatuses = ['PENDING', 'SENT', 'FAILED', 'OPENED', 'CLICKED', 'SUBMITTED']

function addHours(hours) {
  return new Date(Date.now() + hours * 60 * 60 * 1000).toISOString()
}

function defaultDomain() {
  const { protocol, hostname } = window.location
  return `${protocol}//${hostname || 'localhost'}`
}

function emptyResult(rows = 25) {
  return reactive({ items: [], total: 0, page: 1, rowsPerPage: rows })
}

export function useApiConsole() {
  const apiBaseUrl = ref(getStoredApiBase())
  const activeSection = ref('campaigns')
  const traces = ref([])
  const lastError = ref(null)

  watch(apiBaseUrl, (value) => storeApiBase(value))

  const api = createApiClient(() => apiBaseUrl.value, addTrace)

  const loading = reactive({
    bootstrap: false,
    organization: false,
    logo: false,
    departments: false,
    employees: false,
    messages: false,
    landings: false,
    campaigns: false,
    targets: false,
    content: false,
    landingContent: false,
    render: false,
  })

  const organization = ref(null)
  const organizationForm = reactive({
    label: 'ClickSafe Demo',
    attributes: '{\n  "Domain": "clicksafe.test",\n  "Industry": "Education"\n}',
  })
  const selectedLogoFile = ref(null)

  const departmentQuery = reactive({ page: 1, rows: 25, label: '' })
  const departmentResult = emptyResult()
  const departmentForm = reactive({
    id: '',
    label: 'Security Operations',
    attributes: '{\n  "City": "Moscow",\n  "Level": "Blue Team"\n}',
  })

  const employeeQuery = reactive({ page: 1, rows: 25, full_name: '', email: '', department_id: '' })
  const employeeResult = emptyResult()
  const employeeForm = reactive({
    id: '',
    department_id: '',
    first_name: 'Иван',
    last_name: 'Петров',
    email: 'ivan.petrov@clicksafe.test',
    phone: '+79991234567',
    attributes: '{\n  "Role": "Security Analyst",\n  "Seniority": "Middle"\n}',
  })

  const messageQuery = reactive({ page: 1, rows: 25, label: '', subject: '' })
  const messageResult = emptyResult()
  const messageForm = reactive({
    id: '',
    label: 'Security Training Invite',
    from_email: 'training@clicksafe.test',
    from_name: 'ClickSafe Training',
    subject: 'Security awareness training',
  })
  const selectedMessageId = ref('')
  const selectedTemplateFile = ref(null)
  const templateText = ref(defaultMessageTemplate)
  const rawTemplateContent = ref('')
  const renderedContent = ref('')

  const landingQuery = reactive({ page: 1, rows: 25, label: '' })
  const landingResult = emptyResult()
  const landingForm = reactive({
    id: '',
    label: 'Training Portal Landing',
  })
  const selectedLandingId = ref('')
  const selectedLandingFile = ref(null)
  const landingTemplateText = ref(defaultLandingTemplate)
  const rawLandingContent = ref('')
  const renderedLandingContent = ref('')

  const campaignQuery = reactive({ page: 1, rows: 25, label: '', status: '' })
  const campaignResult = emptyResult()
  const campaignForm = reactive({
    id: '',
    message_id: '',
    landing_id: '',
    label: 'ClickSafe Demo Campaign',
    domain: defaultDomain(),
    date_from: addHours(-1),
    date_to: addHours(24),
    attributes: '{\n  "Scenario": "Training",\n  "Risk": "Medium"\n}',
  })
  const selectedCampaignId = ref('')

  const vtargetQuery = reactive({ page: 1, rows: 50, campaign_id: '', full_name: '', status: '' })
  const vtargetResult = emptyResult(50)
  const targetForm = reactive({
    campaign_id: '',
    employee_id: '',
    scheduled_at: addHours(1),
  })
  const selectedTargetId = ref('')
  const targetScheduleForm = reactive({
    scheduled_at: addHours(1),
  })

  const selectedEmployeeId = ref('')

  const selectedMessage = computed(() =>
    messageResult.items.find((message) => message.id === selectedMessageId.value),
  )

  const selectedLanding = computed(() =>
    landingResult.items.find((landing) => landing.id === selectedLandingId.value),
  )

  const selectedEmployee = computed(() =>
    employeeResult.items.find((employee) => employee.id === selectedEmployeeId.value),
  )

  const selectedCampaign = computed(() =>
    campaignResult.items.find((campaign) => campaign.id === selectedCampaignId.value),
  )

  const selectedTarget = computed(() =>
    vtargetResult.items.find((target) => target.id === selectedTargetId.value),
  )

  const apiBaseLabel = computed(() => apiBaseUrl.value || 'same origin / Vite proxy')

  watch(selectedCampaignId, (id) => {
    targetForm.campaign_id = id
    vtargetQuery.campaign_id = id

    if (id) {
      void loadTargets()
    }
  })

  watch(selectedTargetId, () => {
    if (selectedTarget.value?.scheduled_at) {
      targetScheduleForm.scheduled_at = selectedTarget.value.scheduled_at
    }
  })

  function addTrace(trace) {
    const id = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`
    traces.value = [
      {
        ...trace,
        id,
        at: new Date().toLocaleTimeString('ru-RU'),
      },
      ...traces.value,
    ].slice(0, 10)
  }

  function buildQuery(params) {
    const query = new URLSearchParams()

    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined && value !== null && value !== '') {
        query.set(key, value)
      }
    }

    const suffix = query.toString()
    return suffix ? `?${suffix}` : ''
  }

  function parseAttributes(value) {
    const trimmed = value.trim()

    if (!trimmed) {
      return {}
    }

    const parsed = JSON.parse(trimmed)

    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      throw new Error('attributes должен быть JSON-объектом')
    }

    return Object.fromEntries(Object.entries(parsed).map(([key, attrValue]) => [key, String(attrValue)]))
  }

  function pretty(value) {
    return JSON.stringify(value, null, 2)
  }

  function shortId(value) {
    return value ? `${value.slice(0, 8)}...${value.slice(-4)}` : 'не выбран'
  }

  function formatDate(value) {
    if (!value) {
      return 'не задано'
    }

    return new Date(value).toLocaleString('ru-RU')
  }

  function toApiDate(value, field) {
    const date = new Date(value)

    if (!value || Number.isNaN(date.getTime())) {
      throw new Error(`${field} должен быть валидной датой`)
    }

    return date.toISOString()
  }

  function setQueryResult(target, payload) {
    target.items = payload?.items ?? []
    target.total = payload?.total ?? 0
    target.page = payload?.page ?? 1
    target.rowsPerPage = payload?.rowsPerPage ?? target.rowsPerPage
  }

  function keepSelectedId(items, selectedRef) {
    if (items.some((item) => item.id === selectedRef.value)) {
      return
    }

    selectedRef.value = items[0]?.id ?? ''
  }

  function showError(error, title = 'Ошибка API') {
    lastError.value = {
      title,
      message: error.message,
      payload: error.payload,
    }
    ElNotification.error({
      title,
      message: error.message,
      duration: 6000,
    })
  }

  async function confirmAction(message, confirmButtonText = 'Удалить') {
    try {
      await ElMessageBox.confirm(message, 'Подтверждение', {
        type: 'warning',
        confirmButtonText,
        cancelButtonText: 'Отмена',
      })
      return true
    } catch {
      return false
    }
  }

  async function withLoading(key, action, successMessage) {
    loading[key] = true
    lastError.value = null

    try {
      const result = await action()

      if (successMessage) {
        ElMessage.success(successMessage)
      }

      return result
    } catch (error) {
      showError(error)
      return null
    } finally {
      loading[key] = false
    }
  }

  async function loadOrganization() {
    return withLoading('organization', async () => {
      try {
        organization.value = await api.get('/organization')
        organizationForm.label = organization.value.label
        organizationForm.attributes = pretty(organization.value.attributes ?? {})
      } catch (error) {
        if (error.status !== 404) {
          throw error
        }

        organization.value = null
      }
    })
  }

  async function saveOrganization() {
    return withLoading('organization', async () => {
      const payload = {
        label: organizationForm.label,
        attributes: parseAttributes(organizationForm.attributes),
      }

      if (organization.value) {
        await api.put('/organization', payload)
      } else {
        await api.post('/organization', payload)
      }

      await loadOrganization()
    }, 'Организация сохранена')
  }

  function onLogoFileChange(uploadFile) {
    selectedLogoFile.value = uploadFile.raw
  }

  function onLogoFileRemove() {
    selectedLogoFile.value = null
  }

  async function uploadOrganizationLogo() {
    if (!selectedLogoFile.value) {
      ElMessage.warning('Выбери изображение логотипа')
      return null
    }

    return withLoading('logo', async () => {
      const body = new FormData()
      body.append('file', selectedLogoFile.value)
      await api.put('/organization/logo', body)
      await loadOrganization()
    }, 'Логотип загружен')
  }

  async function deleteOrganizationLogo() {
    if (!(await confirmAction('Удалить логотип организации?'))) {
      return null
    }

    return withLoading('logo', async () => {
      await api.delete('/organization/logo')
      await loadOrganization()
    }, 'Логотип удален')
  }

  async function loadDepartments() {
    return withLoading('departments', async () => {
      const query = buildQuery({
        page: departmentQuery.page,
        rows: departmentQuery.rows,
        label: departmentQuery.label,
      })
      setQueryResult(departmentResult, await api.get(`/department${query}`))

      if (!departmentResult.items.some((department) => department.id === employeeForm.department_id)) {
        employeeForm.department_id = departmentResult.items[0]?.id ?? ''
      }
    })
  }

  function editDepartment(row) {
    departmentForm.id = row.id
    departmentForm.label = row.label
    departmentForm.attributes = pretty(row.attributes ?? {})
  }

  function resetDepartmentForm() {
    departmentForm.id = ''
    departmentForm.label = 'Security Operations'
    departmentForm.attributes = '{\n  "City": "Moscow",\n  "Level": "Blue Team"\n}'
  }

  async function saveDepartment() {
    return withLoading('departments', async () => {
      const payload = {
        label: departmentForm.label,
        attributes: parseAttributes(departmentForm.attributes),
      }

      if (departmentForm.id) {
        await api.put(`/department/${departmentForm.id}`, payload)
      } else {
        await api.post('/department', payload)
      }

      await loadDepartments()
    }, departmentForm.id ? 'Отдел обновлен' : 'Отдел создан')
  }

  async function deleteDepartment(id) {
    if (!(await confirmAction('Удалить отдел? Сотрудники с привязкой могут заблокировать удаление.'))) {
      return null
    }

    return withLoading('departments', async () => {
      await api.delete(`/department/${id}`)
      if (employeeForm.department_id === id) {
        employeeForm.department_id = ''
      }
      if (departmentForm.id === id) {
        resetDepartmentForm()
      }
      await loadDepartments()
    }, 'Отдел удален')
  }

  async function loadEmployees() {
    return withLoading('employees', async () => {
      const query = buildQuery({
        page: employeeQuery.page,
        rows: employeeQuery.rows,
        full_name: employeeQuery.full_name,
        email: employeeQuery.email,
        department_id: employeeQuery.department_id,
      })
      setQueryResult(employeeResult, await api.get(`/employee${query}`))
      keepSelectedId(employeeResult.items, selectedEmployeeId)

      if (!targetForm.employee_id) {
        targetForm.employee_id = employeeResult.items[0]?.id ?? ''
      }
    })
  }

  function editEmployee(row) {
    employeeForm.id = row.id
    employeeForm.department_id = row.department_id ?? ''
    employeeForm.first_name = row.first_name
    employeeForm.last_name = row.last_name
    employeeForm.email = row.email
    employeeForm.phone = row.phone
    employeeForm.attributes = pretty(row.attributes ?? {})
    selectedEmployeeId.value = row.id
  }

  function resetEmployeeForm() {
    employeeForm.id = ''
    employeeForm.department_id = departmentResult.items[0]?.id ?? ''
    employeeForm.first_name = 'Иван'
    employeeForm.last_name = 'Петров'
    employeeForm.email = 'ivan.petrov@clicksafe.test'
    employeeForm.phone = '+79991234567'
    employeeForm.attributes = '{\n  "Role": "Security Analyst",\n  "Seniority": "Middle"\n}'
  }

  async function saveEmployee() {
    return withLoading('employees', async () => {
      const payload = {
        first_name: employeeForm.first_name,
        last_name: employeeForm.last_name,
        email: employeeForm.email,
        phone: employeeForm.phone,
        attributes: parseAttributes(employeeForm.attributes),
      }

      if (employeeForm.department_id) {
        payload.department_id = employeeForm.department_id
      }

      let saved
      if (employeeForm.id) {
        saved = await api.put(`/employee/${employeeForm.id}`, payload)
      } else {
        saved = await api.post('/employee', payload)
      }

      selectedEmployeeId.value = saved.id
      await loadEmployees()
    }, employeeForm.id ? 'Сотрудник обновлен' : 'Сотрудник создан')
  }

  async function deleteEmployee(id) {
    if (!(await confirmAction('Удалить сотрудника? Target-ы кампаний могут заблокировать удаление.'))) {
      return null
    }

    return withLoading('employees', async () => {
      await api.delete(`/employee/${id}`)
      if (selectedEmployeeId.value === id) {
        selectedEmployeeId.value = ''
      }
      if (employeeForm.id === id) {
        resetEmployeeForm()
      }
      await loadEmployees()
      await loadTargets()
    }, 'Сотрудник удален')
  }

  async function loadMessages() {
    return withLoading('messages', async () => {
      const query = buildQuery({
        page: messageQuery.page,
        rows: messageQuery.rows,
        label: messageQuery.label,
        subject: messageQuery.subject,
      })
      setQueryResult(messageResult, await api.get(`/message${query}`))
      keepSelectedId(messageResult.items, selectedMessageId)

      if (!campaignForm.message_id) {
        campaignForm.message_id = selectedMessageId.value
      }
    })
  }

  function editMessage(row) {
    messageForm.id = row.id
    messageForm.label = row.label
    messageForm.from_email = row.from_email
    messageForm.from_name = row.from_name
    messageForm.subject = row.subject
    selectedMessageId.value = row.id
  }

  function resetMessageForm() {
    messageForm.id = ''
    messageForm.label = 'Security Training Invite'
    messageForm.from_email = 'training@clicksafe.test'
    messageForm.from_name = 'ClickSafe Training'
    messageForm.subject = 'Security awareness training'
  }

  async function saveMessage() {
    return withLoading('messages', async () => {
      const payload = {
        label: messageForm.label,
        from_email: messageForm.from_email,
        from_name: messageForm.from_name,
        subject: messageForm.subject,
      }

      const saved = messageForm.id
        ? await api.put(`/message/${messageForm.id}`, payload)
        : await api.post('/message', payload)

      selectedMessageId.value = saved.id
      await loadMessages()
    }, messageForm.id ? 'Письмо обновлено' : 'Письмо создано')
  }

  async function deleteMessage(id) {
    if (!(await confirmAction('Удалить письмо? Кампании с привязкой могут заблокировать удаление.'))) {
      return null
    }

    return withLoading('messages', async () => {
      await api.delete(`/message/${id}`)
      if (selectedMessageId.value === id) {
        selectedMessageId.value = ''
        rawTemplateContent.value = ''
        renderedContent.value = ''
      }
      if (messageForm.id === id) {
        resetMessageForm()
      }
      await loadMessages()
      await loadCampaigns()
    }, 'Письмо удалено')
  }

  function onTemplateFileChange(uploadFile) {
    selectedTemplateFile.value = uploadFile.raw
  }

  function onTemplateFileRemove() {
    selectedTemplateFile.value = null
  }

  async function uploadTemplateFile() {
    if (!selectedMessageId.value) {
      ElMessage.warning('Выбери письмо')
      return null
    }

    if (!selectedTemplateFile.value) {
      ElMessage.warning('Выбери HTML-файл')
      return null
    }

    const source = await selectedTemplateFile.value.text().catch(() => '')
    return uploadMessageHtml(selectedMessageId.value, selectedTemplateFile.value, source)
  }

  async function uploadTemplateText() {
    if (!selectedMessageId.value) {
      ElMessage.warning('Выбери письмо')
      return null
    }

    const blob = new Blob([templateText.value], { type: 'text/html;charset=utf-8' })
    return uploadMessageHtml(selectedMessageId.value, blob, templateText.value, 'message-template.html')
  }

  async function uploadMessageHtml(messageID, file, source = '', filename = file.name) {
    return withLoading('content', async () => {
      const body = new FormData()
      body.append('file', file, filename)
      const updated = await api.put(`/message/${messageID}/content`, body)
      const index = messageResult.items.findIndex((message) => message.id === updated.id)

      if (index >= 0) {
        messageResult.items[index] = updated
      }

      if (source) {
        rawTemplateContent.value = source
      }
      await loadMessages()
    }, 'HTML письма загружен')
  }

  async function readTemplateContent() {
    if (!selectedMessageId.value) {
      ElMessage.warning('Выбери письмо')
      return null
    }

    return withLoading('content', async () => {
      rawTemplateContent.value = await api.get(`/message/${selectedMessageId.value}/content`)
      templateText.value = rawTemplateContent.value
    }, 'Исходный HTML письма получен')
  }

  async function loadLandings() {
    return withLoading('landings', async () => {
      const query = buildQuery({
        page: landingQuery.page,
        rows: landingQuery.rows,
        label: landingQuery.label,
      })
      setQueryResult(landingResult, await api.get(`/landing${query}`))
      keepSelectedId(landingResult.items, selectedLandingId)

      if (!campaignForm.landing_id) {
        campaignForm.landing_id = selectedLandingId.value
      }
    })
  }

  function editLanding(row) {
    landingForm.id = row.id
    landingForm.label = row.label
    selectedLandingId.value = row.id
  }

  function resetLandingForm() {
    landingForm.id = ''
    landingForm.label = 'Training Portal Landing'
  }

  async function saveLanding() {
    return withLoading('landings', async () => {
      const payload = { label: landingForm.label }
      const saved = landingForm.id
        ? await api.put(`/landing/${landingForm.id}`, payload)
        : await api.post('/landing', payload)

      selectedLandingId.value = saved.id
      await loadLandings()
    }, landingForm.id ? 'Лендинг обновлен' : 'Лендинг создан')
  }

  async function deleteLanding(id) {
    if (!(await confirmAction('Удалить лендинг? Кампании с привязкой могут заблокировать удаление.'))) {
      return null
    }

    return withLoading('landings', async () => {
      await api.delete(`/landing/${id}`)
      if (selectedLandingId.value === id) {
        selectedLandingId.value = ''
        rawLandingContent.value = ''
        renderedLandingContent.value = ''
      }
      if (landingForm.id === id) {
        resetLandingForm()
      }
      await loadLandings()
      await loadCampaigns()
    }, 'Лендинг удален')
  }

  function onLandingFileChange(uploadFile) {
    selectedLandingFile.value = uploadFile.raw
  }

  function onLandingFileRemove() {
    selectedLandingFile.value = null
  }

  async function uploadLandingFile() {
    if (!selectedLandingId.value) {
      ElMessage.warning('Выбери лендинг')
      return null
    }

    if (!selectedLandingFile.value) {
      ElMessage.warning('Выбери HTML-файл')
      return null
    }

    const source = await selectedLandingFile.value.text().catch(() => '')
    return uploadLandingHtml(selectedLandingId.value, selectedLandingFile.value, source)
  }

  async function uploadLandingText() {
    if (!selectedLandingId.value) {
      ElMessage.warning('Выбери лендинг')
      return null
    }

    const blob = new Blob([landingTemplateText.value], { type: 'text/html;charset=utf-8' })
    return uploadLandingHtml(selectedLandingId.value, blob, landingTemplateText.value, 'landing-template.html')
  }

  async function uploadLandingHtml(landingID, file, source = '', filename = file.name) {
    return withLoading('landingContent', async () => {
      const body = new FormData()
      body.append('file', file, filename)
      const updated = await api.put(`/landing/${landingID}/content`, body)
      const index = landingResult.items.findIndex((landing) => landing.id === updated.id)

      if (index >= 0) {
        landingResult.items[index] = updated
      }

      if (source) {
        rawLandingContent.value = source
      }
      await loadLandings()
    }, 'HTML лендинга загружен')
  }

  async function readLandingContent() {
    if (!selectedLandingId.value) {
      ElMessage.warning('Выбери лендинг')
      return null
    }

    return withLoading('landingContent', async () => {
      rawLandingContent.value = await api.get(`/landing/${selectedLandingId.value}/content`)
      landingTemplateText.value = rawLandingContent.value
    }, 'Исходный HTML лендинга получен')
  }

  async function renderMessage(messageID) {
    const id = typeof messageID === 'string' ? messageID : selectedMessageId.value

    if (!id) {
      ElMessage.warning('Выбери письмо')
      return null
    }

    if (!selectedTargetId.value) {
      ElMessage.warning('Выбери target в кампании')
      return null
    }

    return withLoading('render', async () => {
      const rendered = await api.post(`/message/${id}/render`, {
        target_id: selectedTargetId.value,
      })
      renderedContent.value = rendered.content
    }, 'Письмо отрендерено')
  }

  async function renderLanding(landingID) {
    const id = typeof landingID === 'string' ? landingID : selectedLandingId.value

    if (!id) {
      ElMessage.warning('Выбери лендинг')
      return null
    }

    if (!selectedTargetId.value) {
      ElMessage.warning('Выбери target в кампании')
      return null
    }

    return withLoading('render', async () => {
      const rendered = await api.post(`/landing/${id}/render`, {
        target_id: selectedTargetId.value,
      })
      renderedLandingContent.value = rendered.content
    }, 'Лендинг отрендерен')
  }

  async function loadCampaigns() {
    return withLoading('campaigns', async () => {
      const query = buildQuery({
        page: campaignQuery.page,
        rows: campaignQuery.rows,
        label: campaignQuery.label,
        status: campaignQuery.status,
      })
      setQueryResult(campaignResult, await api.get(`/campaign${query}`))
      keepSelectedId(campaignResult.items, selectedCampaignId)
    })
  }

  function editCampaign(row) {
    campaignForm.id = row.id
    campaignForm.message_id = row.message_id ?? ''
    campaignForm.landing_id = row.landing_id ?? ''
    campaignForm.label = row.label
    campaignForm.domain = row.domain
    campaignForm.date_from = row.date_from ?? addHours(-1)
    campaignForm.date_to = row.date_to ?? addHours(24)
    campaignForm.attributes = pretty(row.attributes ?? {})
    selectedCampaignId.value = row.id

    if (row.message_id) {
      selectedMessageId.value = row.message_id
    }
    if (row.landing_id) {
      selectedLandingId.value = row.landing_id
    }
  }

  function resetCampaignForm() {
    campaignForm.id = ''
    campaignForm.message_id = selectedMessageId.value
    campaignForm.landing_id = selectedLandingId.value
    campaignForm.label = 'ClickSafe Demo Campaign'
    campaignForm.domain = defaultDomain()
    campaignForm.date_from = addHours(-1)
    campaignForm.date_to = addHours(24)
    campaignForm.attributes = '{\n  "Scenario": "Training",\n  "Risk": "Medium"\n}'
  }

  function campaignPayload() {
    const payload = {
      label: campaignForm.label,
      domain: campaignForm.domain,
      date_from: toApiDate(campaignForm.date_from, 'date_from'),
      date_to: toApiDate(campaignForm.date_to, 'date_to'),
      attributes: parseAttributes(campaignForm.attributes),
    }

    if (campaignForm.message_id) {
      payload.message_id = campaignForm.message_id
    }
    if (campaignForm.landing_id) {
      payload.landing_id = campaignForm.landing_id
    }

    return payload
  }

  async function saveCampaign() {
    return withLoading('campaigns', async () => {
      const payload = campaignPayload()
      const saved = campaignForm.id
        ? await api.put(`/campaign/${campaignForm.id}`, payload)
        : await api.post('/campaign', payload)

      selectedCampaignId.value = saved.id
      await loadCampaigns()
      await loadTargets()
    }, campaignForm.id ? 'Кампания обновлена' : 'Кампания создана')
  }

  async function deleteCampaign(id) {
    if (!(await confirmAction('Удалить кампанию вместе с target-ами?'))) {
      return null
    }

    return withLoading('campaigns', async () => {
      await api.delete(`/campaign/${id}`)
      if (selectedCampaignId.value === id) {
        selectedCampaignId.value = ''
        selectedTargetId.value = ''
      }
      if (campaignForm.id === id) {
        resetCampaignForm()
      }
      await loadCampaigns()
      await loadTargets()
    }, 'Кампания удалена')
  }

  async function changeCampaignStatus(id, action) {
    const messages = {
      start: 'Кампания запущена',
      pause: 'Кампания поставлена на паузу',
      cancel: 'Кампания отменена',
    }

    if (action === 'cancel' && !(await confirmAction('Отменить кампанию?', 'Отменить'))) {
      return null
    }

    return withLoading('campaigns', async () => {
      const updated = await api.put(`/campaign/${id}/${action}`, {})
      const index = campaignResult.items.findIndex((campaign) => campaign.id === updated.id)

      if (index >= 0) {
        campaignResult.items[index] = updated
      }

      editCampaign(updated)
    }, messages[action])
  }

  async function loadTargets() {
    return withLoading('targets', async () => {
      const query = buildQuery({
        page: vtargetQuery.page,
        rows: vtargetQuery.rows,
        campaign_id: vtargetQuery.campaign_id,
        full_name: vtargetQuery.full_name,
        status: vtargetQuery.status,
      })
      setQueryResult(vtargetResult, await api.get(`/vtarget${query}`))
      keepSelectedId(vtargetResult.items, selectedTargetId)

      if (selectedTarget.value?.scheduled_at) {
        targetScheduleForm.scheduled_at = selectedTarget.value.scheduled_at
      }
    })
  }

  async function createTarget() {
    return withLoading('targets', async () => {
      const campaignID = targetForm.campaign_id || selectedCampaignId.value
      if (!campaignID) {
        throw new Error('Выбери кампанию')
      }
      if (!targetForm.employee_id) {
        throw new Error('Выбери сотрудника')
      }

      const created = await api.post('/target', {
        campaign_id: campaignID,
        employee_id: targetForm.employee_id,
      })

      selectedTargetId.value = created.id
      await loadTargets()
    }, 'Target добавлен')
  }

  async function deleteTarget(id) {
    if (!(await confirmAction('Удалить target из кампании?'))) {
      return null
    }

    return withLoading('targets', async () => {
      await api.delete(`/target/${id}`)
      if (selectedTargetId.value === id) {
        selectedTargetId.value = ''
      }
      await loadTargets()
    }, 'Target удален')
  }

  async function deleteCampaignTargets() {
    const campaignID = selectedCampaignId.value
    if (!campaignID) {
      ElMessage.warning('Выбери кампанию')
      return null
    }

    if (!(await confirmAction('Удалить все target-ы выбранной кампании?'))) {
      return null
    }

    return withLoading('targets', async () => {
      await api.delete(`/target/campaign/${campaignID}`)
      selectedTargetId.value = ''
      await loadTargets()
    }, 'Target-ы кампании удалены')
  }

  async function updateTargetSchedule() {
    if (!selectedTargetId.value) {
      ElMessage.warning('Выбери target')
      return null
    }

    return withLoading('targets', async () => {
      await api.put(`/target/${selectedTargetId.value}/schedule`, {
        scheduled_at: toApiDate(targetScheduleForm.scheduled_at, 'scheduled_at'),
      })
      await loadTargets()
    }, 'Расписание target обновлено')
  }

  async function autoDistributeTargets() {
    const campaignID = selectedCampaignId.value
    if (!campaignID) {
      ElMessage.warning('Выбери кампанию')
      return null
    }

    return withLoading('targets', async () => {
      await api.put(`/target/campaign/${campaignID}/distribute`, {})
      await loadTargets()
    }, 'Target-ы распределены')
  }

  function campaignByID(id) {
    return campaignResult.items.find((campaign) => campaign.id === id)
  }

  function messageByID(id) {
    return messageResult.items.find((message) => message.id === id)
  }

  function landingByID(id) {
    return landingResult.items.find((landing) => landing.id === id)
  }

  function employeeByID(id) {
    return employeeResult.items.find((employee) => employee.id === id)
  }

  function departmentByID(id) {
    return departmentResult.items.find((department) => department.id === id)
  }

  function employeeName(employee) {
    return employee ? `${employee.first_name} ${employee.last_name}` : 'не выбран'
  }

  function targetName(target) {
    return target ? `${target.first_name} ${target.last_name}` : 'не выбран'
  }

  function visitLink(target = selectedTarget.value) {
    if (!target?.token) {
      return ''
    }

    const campaign = campaignByID(target.campaign_id) ?? selectedCampaign.value
    if (!campaign?.domain) {
      return ''
    }

    return `${campaign.domain.replace(/\/$/, '')}/${target.token}`
  }

  async function uploadHtml(path, html, filename) {
    const body = new FormData()
    body.append('file', new Blob([html], { type: 'text/html;charset=utf-8' }), filename)
    return api.put(path, body)
  }

  async function queryFirst(path, params, predicate) {
    const result = await api.get(`${path}${buildQuery({ ...params, rows: 100 })}`)
    return result.items.find(predicate)
  }

  async function ensureOrganization() {
    const payload = {
      label: organizationForm.label,
      attributes: parseAttributes(organizationForm.attributes),
    }

    try {
      const existing = await api.get('/organization')
      await api.put('/organization', payload)
      return { ...existing, ...payload }
    } catch (error) {
      if (error.status !== 404) {
        throw error
      }
      return api.post('/organization', payload)
    }
  }

  async function ensureDepartment() {
    const label = departmentForm.label
    const existing = await queryFirst('/department', { label }, (department) => department.label === label)

    if (existing) {
      await api.put(`/department/${existing.id}`, {
        label,
        attributes: parseAttributes(departmentForm.attributes),
      })
      return { ...existing, label, attributes: parseAttributes(departmentForm.attributes) }
    }

    return api.post('/department', {
      label,
      attributes: parseAttributes(departmentForm.attributes),
    })
  }

  async function ensureEmployee(record) {
    const existing = await queryFirst('/employee', { email: record.email }, (employee) => employee.email === record.email)
    const payload = {
      department_id: record.department_id,
      first_name: record.first_name,
      last_name: record.last_name,
      email: record.email,
      phone: record.phone,
      attributes: record.attributes,
    }

    if (existing) {
      return api.put(`/employee/${existing.id}`, payload)
    }

    return api.post('/employee', payload)
  }

  async function ensureMessage() {
    const label = messageForm.label
    const existing = await queryFirst('/message', { label }, (message) => message.label === label)
    const payload = {
      label,
      from_email: messageForm.from_email,
      from_name: messageForm.from_name,
      subject: messageForm.subject,
    }
    const message = existing ? await api.put(`/message/${existing.id}`, payload) : await api.post('/message', payload)
    return uploadHtml(`/message/${message.id}/content`, defaultMessageTemplate, 'message-template.html')
  }

  async function ensureLanding() {
    const label = landingForm.label
    const existing = await queryFirst('/landing', { label }, (landing) => landing.label === label)
    const landing = existing ? await api.put(`/landing/${existing.id}`, { label }) : await api.post('/landing', { label })
    return uploadHtml(`/landing/${landing.id}/content`, defaultLandingTemplate, 'landing-template.html')
  }

  async function ensureCampaign(messageID, landingID) {
    const baseLabel = campaignForm.label || 'ClickSafe Demo Campaign'
    const result = await api.get(`/campaign${buildQuery({ label: baseLabel, rows: 100 })}`)
    const exact = result.items.find((campaign) => campaign.label === baseLabel)
    const existing = result.items.find((campaign) => (
      campaign.label === baseLabel && !['COMPLETED', 'CANCELED'].includes(campaign.status)
    ))
    const suffix = new Date().toISOString().slice(11, 19).replaceAll(':', '')
    const label = exact && !existing ? `${baseLabel} ${suffix}` : baseLabel
    const payload = {
      message_id: messageID,
      landing_id: landingID,
      label: existing ? existing.label : label,
      domain: campaignForm.domain || defaultDomain(),
      date_from: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
      date_to: addHours(24),
      attributes: parseAttributes(campaignForm.attributes),
    }

    if (existing) {
      if (['DRAFT', 'PAUSED'].includes(existing.status)) {
        return api.put(`/campaign/${existing.id}`, payload)
      }
      return existing
    }

    return api.post('/campaign', payload)
  }

  async function ensureTargets(campaignID, employees) {
    const targets = []

    for (const employee of employees) {
      const existing = await queryFirst('/vtarget', {
        campaign_id: campaignID,
        employee_id: employee.id,
      }, (target) => target.employee_id === employee.id)

      if (existing) {
        targets.push(existing)
        continue
      }

      targets.push(await api.post('/target', {
        campaign_id: campaignID,
        employee_id: employee.id,
      }))
    }

    return targets
  }

  async function bootstrapDemo() {
    loading.bootstrap = true
    lastError.value = null

    try {
      await ensureOrganization()
      const department = await ensureDepartment()
      employeeForm.department_id = department.id

      const employees = await Promise.all([
        ensureEmployee({
          department_id: department.id,
          first_name: 'Иван',
          last_name: 'Петров',
          email: 'ivan.petrov@clicksafe.test',
          phone: '+79991234567',
          attributes: { Role: 'Security Analyst', Seniority: 'Middle' },
        }),
        ensureEmployee({
          department_id: department.id,
          first_name: 'Мария',
          last_name: 'Смирнова',
          email: 'maria.smirnova@clicksafe.test',
          phone: '+79997654321',
          attributes: { Role: 'Account Manager', Seniority: 'Senior' },
        }),
        ensureEmployee({
          department_id: department.id,
          first_name: 'Алексей',
          last_name: 'Кузнецов',
          email: 'alexey.kuznetsov@clicksafe.test',
          phone: '+79995550101',
          attributes: { Role: 'Developer', Seniority: 'Junior' },
        }),
      ])

      const message = await ensureMessage()
      const landing = await ensureLanding()
      const campaign = await ensureCampaign(message.id, landing.id)
      await ensureTargets(campaign.id, employees)

      if (['DRAFT', 'PAUSED'].includes(campaign.status)) {
        await api.put(`/target/campaign/${campaign.id}/distribute`, {})
        await api.put(`/campaign/${campaign.id}/start`, {})
      }

      selectedCampaignId.value = campaign.id
      selectedMessageId.value = message.id
      selectedLandingId.value = landing.id
      selectedEmployeeId.value = employees[0].id
      activeSection.value = 'campaigns'

      await refreshAll()
      await loadTargets()
      ElMessage.success('Демо-данные созданы: организация, люди, письмо, лендинг, кампания и targets')
    } catch (error) {
      showError(error, 'Не удалось собрать демо')
    } finally {
      loading.bootstrap = false
    }
  }

  async function refreshAll() {
    await Promise.all([
      loadOrganization(),
      loadDepartments(),
      loadEmployees(),
      loadMessages(),
      loadLandings(),
      loadCampaigns(),
    ])
  }

  function clearLastError() {
    lastError.value = null
  }

  function clearTraces() {
    traces.value = []
  }

  onMounted(refreshAll)

  return {
    activeSection,
    apiBaseLabel,
    apiBaseUrl,
    autoDistributeTargets,
    bootstrapDemo,
    campaignByID,
    campaignForm,
    campaignQuery,
    campaignResult,
    campaignStatuses,
    changeCampaignStatus,
    clearLastError,
    clearTraces,
    createTarget,
    deleteCampaign,
    deleteCampaignTargets,
    deleteDepartment,
    deleteEmployee,
    deleteLanding,
    deleteMessage,
    deleteOrganizationLogo,
    deleteTarget,
    departmentByID,
    departmentForm,
    departmentQuery,
    departmentResult,
    editCampaign,
    editDepartment,
    editEmployee,
    editLanding,
    editMessage,
    employeeByID,
    employeeForm,
    employeeName,
    employeeQuery,
    employeeResult,
    formatDate,
    landingByID,
    landingForm,
    landingQuery,
    landingResult,
    landingTemplateText,
    lastError,
    loadCampaigns,
    loadDepartments,
    loadEmployees,
    loadLandings,
    loadMessages,
    loadOrganization,
    loadTargets,
    loading,
    messageByID,
    messageForm,
    messageQuery,
    messageResult,
    onLandingFileChange,
    onLandingFileRemove,
    onLogoFileChange,
    onLogoFileRemove,
    onTemplateFileChange,
    onTemplateFileRemove,
    organization,
    organizationForm,
    pretty,
    rawLandingContent,
    rawTemplateContent,
    readLandingContent,
    readTemplateContent,
    refreshAll,
    renderLanding,
    renderMessage,
    renderedContent,
    renderedLandingContent,
    resetCampaignForm,
    resetDepartmentForm,
    resetEmployeeForm,
    resetLandingForm,
    resetMessageForm,
    saveCampaign,
    saveDepartment,
    saveEmployee,
    saveLanding,
    saveMessage,
    saveOrganization,
    selectedCampaign,
    selectedCampaignId,
    selectedEmployee,
    selectedEmployeeId,
    selectedLanding,
    selectedLandingId,
    selectedLogoFile,
    selectedMessage,
    selectedMessageId,
    selectedTarget,
    selectedTargetId,
    shortId,
    targetForm,
    targetName,
    targetScheduleForm,
    targetStatuses,
    templateText,
    traces,
    updateTargetSchedule,
    uploadLandingFile,
    uploadLandingText,
    uploadOrganizationLogo,
    uploadTemplateFile,
    uploadTemplateText,
    visitLink,
    vtargetQuery,
    vtargetResult,
  }
}

import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import { createApiClient, getStoredApiBase, storeApiBase } from '../api'
import { defaultTemplate } from '../data/defaultTemplate'

export function useApiConsole() {
  const apiBaseUrl = ref(getStoredApiBase())
  const activeSection = ref('messages')
  const traces = ref([])
  const lastError = ref(null)

  watch(apiBaseUrl, (value) => storeApiBase(value))

  const api = createApiClient(() => apiBaseUrl.value, addTrace)

  const loading = reactive({
    bootstrap: false,
    organization: false,
    departments: false,
    employees: false,
    messages: false,
    content: false,
    render: false,
    event: false,
  })

  const organization = ref(null)
  const organizationForm = reactive({
    label: 'ClickSafe Demo',
    attributes: '{\n  "Domain": "clicksafe.test",\n  "Industry": "Education"\n}',
  })

  const departmentQuery = reactive({ page: 1, rows: 25, label: '' })
  const departmentResult = reactive({ items: [], total: 0, page: 1, rowsPerPage: 25 })
  const departmentForm = reactive({
    label: 'Security Operations',
    attributes: '{\n  "City": "Moscow",\n  "Level": "Blue Team"\n}',
  })

  const employeeQuery = reactive({ page: 1, rows: 25, full_name: '', email: '' })
  const employeeResult = reactive({ items: [], total: 0, page: 1, rowsPerPage: 25 })
  const employeeForm = reactive({
    department_id: '',
    first_name: 'Иван',
    last_name: 'Петров',
    email: 'ivan.petrov@clicksafe.test',
    phone: '+79991234567',
    attributes: '{\n  "Role": "Security Analyst",\n  "Seniority": "Middle"\n}',
  })

  const messageQuery = reactive({ page: 1, rows: 25, label: '', subject: '' })
  const messageResult = reactive({ items: [], total: 0, page: 1, rowsPerPage: 25 })
  const messageForm = reactive({
    label: 'Security Training Invite',
    from_email: 'training@clicksafe.test',
    from_name: 'ClickSafe Training',
    subject: 'Security awareness training',
  })
  const selectedMessageId = ref('')
  const selectedEmployeeId = ref('')
  const selectedTemplateFile = ref(null)
  const templateText = ref(defaultTemplate)
  const rawTemplateContent = ref('')
  const renderedContent = ref('')

  const eventForm = reactive({
    campaign_id: '',
    employee_id: '',
    type: 'EMAIL_OPENED',
  })

  const selectedMessage = computed(() =>
    messageResult.items.find((message) => message.id === selectedMessageId.value),
  )

  const selectedEmployee = computed(() =>
    employeeResult.items.find((employee) => employee.id === selectedEmployeeId.value),
  )

  const apiBaseLabel = computed(() => apiBaseUrl.value || 'same origin / Vite proxy')

  function addTrace(trace) {
    const id = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`
    traces.value = [
      {
        ...trace,
        id,
        at: new Date().toLocaleTimeString('ru-RU'),
      },
      ...traces.value,
    ].slice(0, 8)
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

  async function createDepartment() {
    return withLoading('departments', async () => {
      await api.post('/department', {
        label: departmentForm.label,
        attributes: parseAttributes(departmentForm.attributes),
      })
      await loadDepartments()
    }, 'Отдел создан')
  }

  async function deleteDepartment(id) {
    return withLoading('departments', async () => {
      await api.delete(`/department/${id}`)
      if (employeeForm.department_id === id) {
        employeeForm.department_id = ''
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
      })
      setQueryResult(employeeResult, await api.get(`/employee${query}`))
      keepSelectedId(employeeResult.items, selectedEmployeeId)
    })
  }

  async function createEmployee() {
    return withLoading('employees', async () => {
      const created = await api.post('/employee', {
        department_id: employeeForm.department_id,
        first_name: employeeForm.first_name,
        last_name: employeeForm.last_name,
        email: employeeForm.email,
        phone: employeeForm.phone,
        attributes: parseAttributes(employeeForm.attributes),
      })

      selectedEmployeeId.value = created.id
      await loadEmployees()
    }, 'Сотрудник создан')
  }

  async function deleteEmployee(id) {
    return withLoading('employees', async () => {
      await api.delete(`/employee/${id}`)
      if (selectedEmployeeId.value === id) {
        selectedEmployeeId.value = ''
      }
      await loadEmployees()
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
    })
  }

  async function createMessage() {
    return withLoading('messages', async () => {
      const created = await api.post('/message', {
        label: messageForm.label,
        from_email: messageForm.from_email,
        from_name: messageForm.from_name,
        subject: messageForm.subject,
      })

      selectedMessageId.value = created.id
      await loadMessages()
    }, 'Шаблон создан')
  }

  async function deleteMessage(id) {
    return withLoading('messages', async () => {
      await api.delete(`/message/${id}`)
      if (selectedMessageId.value === id) {
        selectedMessageId.value = ''
        rawTemplateContent.value = ''
        renderedContent.value = ''
      }
      await loadMessages()
    }, 'Шаблон удален')
  }

  function onTemplateFileChange(uploadFile) {
    selectedTemplateFile.value = uploadFile.raw
  }

  function onTemplateFileRemove() {
    selectedTemplateFile.value = null
  }

  async function uploadTemplateFile() {
    if (!selectedMessageId.value) {
      ElMessage.warning('Выбери message')
      return
    }

    if (!selectedTemplateFile.value) {
      ElMessage.warning('Выбери HTML-файл')
      return
    }

    const body = new FormData()
    body.append('file', selectedTemplateFile.value)
    const source = await selectedTemplateFile.value.text().catch(() => '')

    return uploadTemplateBody(body, source)
  }

  async function uploadTemplateText() {
    if (!selectedMessageId.value) {
      ElMessage.warning('Выбери message')
      return
    }

    const body = new FormData()
    const blob = new Blob([templateText.value], { type: 'text/html;charset=utf-8' })
    body.append('file', blob, 'message-template.html')

    return uploadTemplateBody(body, templateText.value)
  }

  async function uploadTemplateBody(body, source = '') {
    return withLoading('content', async () => {
      const updated = await api.put(`/message/${selectedMessageId.value}/content`, body)
      const index = messageResult.items.findIndex((message) => message.id === updated.id)

      if (index >= 0) {
        messageResult.items[index] = updated
      }

      if (source) {
        rawTemplateContent.value = source
      }
      await loadMessages()
    }, 'HTML-шаблон загружен')
  }

  async function readTemplateContent() {
    if (!selectedMessageId.value) {
      ElMessage.warning('Выбери message')
      return
    }

    return withLoading('content', async () => {
      rawTemplateContent.value = await api.get(`/message/${selectedMessageId.value}/content`)
      templateText.value = rawTemplateContent.value
    }, 'Исходный HTML получен')
  }

  async function renderMessage() {
    if (!selectedMessageId.value) {
      ElMessage.warning('Выбери message')
      return
    }

    if (!selectedEmployeeId.value) {
      ElMessage.warning('Выбери сотрудника')
      return
    }

    return withLoading('render', async () => {
      const rendered = await api.post(`/message/${selectedMessageId.value}/render`, {
        employee_id: selectedEmployeeId.value,
      })
      renderedContent.value = rendered.content
    }, 'Шаблон отрендерен')
  }

  async function publishEvent() {
    return withLoading('event', async () => {
      await api.post('/events', {
        campaign_id: eventForm.campaign_id,
        employee_id: eventForm.employee_id || selectedEmployeeId.value,
        type: eventForm.type,
      })
    }, 'Событие опубликовано')
  }

  async function ensureOrganization() {
    const payload = {
      label: organizationForm.label,
      attributes: parseAttributes(organizationForm.attributes),
    }

    try {
      const existing = await api.get('/organization')
      await api.put('/organization', payload)
      organization.value = { ...existing, ...payload }
    } catch (error) {
      if (error.status !== 404) {
        throw error
      }
      await api.post('/organization', payload)
      organization.value = await api.get('/organization')
    }
  }

  async function ensureDepartment() {
    const label = departmentForm.label
    const query = buildQuery({ label, rows: 100 })
    const result = await api.get(`/department${query}`)
    const existing = result.items.find((department) => department.label === label)

    if (existing) {
      return existing
    }

    await api.post('/department', {
      label,
      attributes: parseAttributes(departmentForm.attributes),
    })

    const created = await api.get(`/department${query}`)
    return created.items.find((department) => department.label === label)
  }

  async function ensureEmployee(departmentId) {
    const email = employeeForm.email
    const result = await api.get(`/employee${buildQuery({ email, rows: 100 })}`)
    const existing = result.items.find((employee) => employee.email === email)

    if (existing) {
      return existing
    }

    return api.post('/employee', {
      department_id: departmentId,
      first_name: employeeForm.first_name,
      last_name: employeeForm.last_name,
      email,
      phone: employeeForm.phone,
      attributes: parseAttributes(employeeForm.attributes),
    })
  }

  async function ensureMessage() {
    const label = messageForm.label
    const result = await api.get(`/message${buildQuery({ label, rows: 100 })}`)
    const existing = result.items.find((message) => message.label === label)

    if (existing) {
      return existing
    }

    return api.post('/message', {
      label,
      from_email: messageForm.from_email,
      from_name: messageForm.from_name,
      subject: messageForm.subject,
    })
  }

  async function bootstrapDemo() {
    loading.bootstrap = true
    lastError.value = null

    try {
      await ensureOrganization()
      const department = await ensureDepartment()
      employeeForm.department_id = department.id
      const employee = await ensureEmployee(department.id)
      const message = await ensureMessage()

      selectedEmployeeId.value = employee.id
      selectedMessageId.value = message.id
      await uploadTemplateText()
      await renderMessage()
      await Promise.all([loadOrganization(), loadDepartments(), loadEmployees(), loadMessages()])
      ElMessage.success('Демо-цепочка готова')
    } catch (error) {
      showError(error, 'Не удалось собрать демо')
    } finally {
      loading.bootstrap = false
    }
  }

  async function refreshAll() {
    await Promise.all([loadOrganization(), loadDepartments(), loadEmployees(), loadMessages()])
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
    clearLastError,
    clearTraces,
    createDepartment,
    createEmployee,
    createMessage,
    deleteDepartment,
    deleteEmployee,
    deleteMessage,
    departmentForm,
    departmentQuery,
    departmentResult,
    employeeForm,
    employeeQuery,
    employeeResult,
    eventForm,
    lastError,
    loadDepartments,
    loadEmployees,
    loadMessages,
    loadOrganization,
    loading,
    messageForm,
    messageQuery,
    messageResult,
    onTemplateFileChange,
    onTemplateFileRemove,
    organization,
    organizationForm,
    pretty,
    publishEvent,
    rawTemplateContent,
    readTemplateContent,
    refreshAll,
    renderMessage,
    renderedContent,
    saveOrganization,
    selectedEmployee,
    selectedEmployeeId,
    selectedMessage,
    selectedMessageId,
    shortId,
    templateText,
    traces,
    uploadTemplateFile,
    uploadTemplateText,
    bootstrapDemo,
  }
}

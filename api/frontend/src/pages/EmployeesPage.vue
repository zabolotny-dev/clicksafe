<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Pencil, Plus, Trash2, Upload } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import AttributesEditor from '../components/ui/AttributesEditor.vue'
import ConfirmDialog from '../components/ui/ConfirmDialog.vue'
import IconButton from '../components/ui/IconButton.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import ResourcePagination from '../components/ui/ResourcePagination.vue'
import ResourceDrawer from '../components/ui/ResourceDrawer.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiInput from '../components/ui/UiInput.vue'
import { useNotifications } from '../composables/useNotifications'
import { useResourceActions } from '../composables/useResourceActions'
import { listDepartments } from '../resources/departments'
import {
  createEmployee,
  deleteEmployee,
  importEmployees,
  listEmployees,
  updateEmployee,
} from '../resources/employees'
import {
  employeeName,
  errorMessage,
  formatAttributes,
  importErrorInfo,
  nullableId,
} from '../utils/resourceFormat'
import { loadAllPages } from '../utils/resourcePagination'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const steps = computed(() => [
  t('pages.employees.steps.basicInfo'),
  t('common.labels.department'),
  t('pages.employees.steps.review'),
])

const employees = ref([])
const departments = ref([])
const page = ref(1)
const rows = ref(10)
const total = ref(0)
const selectedId = ref('')
const selectedEmployeeRecord = ref(null)
const isLoading = ref(false)
const isImporting = ref(false)
const isSaving = ref(false)
const isDeleting = ref(false)
const loadError = ref('')
const importErrorMessage = ref('')
const importErrorItems = ref([])
const editFormError = ref('')
const createFormError = ref('')
const editDrawerOpen = ref(false)
const deleteDialogOpen = ref(false)
const discardDialogOpen = ref(false)
const activeStep = ref(0)
const departmentSearch = ref('')
const originalEditSnapshot = ref('')
const importFileInput = ref(null)
const filters = reactive({
  fullName: '',
  email: '',
  phone: '',
  departmentId: '',
})
const editForm = reactive(makeEmployeeForm())
const createForm = reactive(makeEmployeeForm())

const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()

const isCreateMode = computed(() => route.name === 'employees-new')
const selectedEmployee = computed(() => selectedEmployeeRecord.value || employees.value.find((employee) => employee.id === selectedId.value) || null)
const selectedCreateDepartment = computed(() => departments.value.find((department) => department.id === createForm.department_id) || null)
const departmentById = computed(() => departments.value.reduce((acc, department) => {
  acc[department.id] = department
  return acc
}, {}))
const filteredDepartments = computed(() => {
  const query = departmentSearch.value.trim().toLowerCase()

  if (!query) {
    return departments.value
  }

  return departments.value.filter((department) => {
    const attributes = formatAttributes(department.attributes).toLowerCase()
    return department.label?.toLowerCase().includes(query) || attributes.includes(query)
  })
})
const editFormDirty = computed(() => {
  if (!editDrawerOpen.value || !originalEditSnapshot.value) {
    return false
  }

  return originalEditSnapshot.value !== stableStringify(employeeFormSnapshot(editForm))
})
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / rows.value)))

function makeEmployeeForm() {
  return {
    department_id: '',
    first_name: '',
    last_name: '',
    email: '',
    phone: '',
    attributes: {},
  }
}

function stableStringify(value) {
  if (Array.isArray(value)) {
    return JSON.stringify(value.map((item) => JSON.parse(stableStringify(item))))
  }

  if (value && typeof value === 'object') {
    return JSON.stringify(Object.keys(value).sort().reduce((acc, key) => {
      acc[key] = value[key] && typeof value[key] === 'object'
        ? JSON.parse(stableStringify(value[key]))
        : value[key]
      return acc
    }, {}))
  }

  return JSON.stringify(value)
}

function employeeFormSnapshot(form) {
  return {
    department_id: form.department_id || '',
    first_name: form.first_name || '',
    last_name: form.last_name || '',
    email: form.email || '',
    phone: form.phone || '',
    attributes: Object.keys(form.attributes || {}).sort().reduce((acc, key) => {
      acc[key] = form.attributes[key]
      return acc
    }, {}),
  }
}

function assignEmployeeForm(target, source) {
  target.department_id = source.department_id || ''
  target.first_name = source.first_name || ''
  target.last_name = source.last_name || ''
  target.email = source.email || ''
  target.phone = source.phone || ''
  target.attributes = { ...(source.attributes || {}) }
}

function resetCreateFlow() {
  activeStep.value = 0
  departmentSearch.value = ''
  createFormError.value = ''
  assignEmployeeForm(createForm, makeEmployeeForm())
}

function departmentLabel(id) {
  return departmentById.value[nullableId(id) || id]?.label || t('common.selection.noDepartment')
}

function clearFilters() {
  filters.fullName = ''
  filters.email = ''
  filters.phone = ''
  filters.departmentId = ''
  page.value = 1
  loadData()
}

function clearImportError() {
  importErrorMessage.value = ''
  importErrorItems.value = []
}

function resetImportInput() {
  if (importFileInput.value) {
    importFileInput.value.value = ''
  }
}

function openImportPicker() {
  if (!isImporting.value) {
    importFileInput.value?.click()
  }
}

function formatImportRowError(rowError) {
  return t('pages.employees.import.rowError', {
    row: rowError.row || t('common.placeholder'),
    field: rowError.field ? `, ${rowError.field}` : '',
    error: rowError.error || t('pages.employees.import.failedMessage'),
  })
}

async function handleImportFileChange(event) {
  const file = event.target.files?.[0]

  if (!file) {
    return
  }

  clearImportError()

  if (!(await ensureCsrfToken())) {
    resetImportInput()
    return
  }

  isImporting.value = true

  try {
    await importEmployees(file, mutationOptions())
    notifySuccess(t('pages.employees.import.success'))
    await loadData()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    const importError = importErrorInfo(error, formatImportRowError)
    importErrorMessage.value = importError.message || t('pages.employees.import.failedMessage')
    importErrorItems.value = importError.items
  } finally {
    isImporting.value = false
    resetImportInput()
  }
}

function openEditDrawer(employee) {
  selectedId.value = employee.id
  selectedEmployeeRecord.value = employee
  assignEmployeeForm(editForm, {
    department_id: nullableId(employee.department_id),
    first_name: employee.first_name,
    last_name: employee.last_name,
    email: employee.email,
    phone: employee.phone,
    attributes: employee.attributes,
  })
  editFormError.value = ''
  originalEditSnapshot.value = stableStringify(employeeFormSnapshot(editForm))
  editDrawerOpen.value = true
}

function closeEditDrawer() {
  editDrawerOpen.value = false
  discardDialogOpen.value = false
  selectedId.value = ''
  selectedEmployeeRecord.value = null
  editFormError.value = ''
  originalEditSnapshot.value = ''
  assignEmployeeForm(editForm, makeEmployeeForm())
}

function requestCloseDrawer() {
  if (editFormDirty.value) {
    discardDialogOpen.value = true
    return
  }

  closeEditDrawer()
}

function validateForm(form) {
  if (!form.first_name.trim()) {
    return t('pages.employees.validation.firstNameRequired')
  }

  if (!form.last_name.trim()) {
    return t('pages.employees.validation.lastNameRequired')
  }

  if (!form.email.trim()) {
    return t('pages.employees.validation.emailRequired')
  }

  return ''
}

function buildCreatePayload() {
  return {
    department_id: createForm.department_id,
    first_name: createForm.first_name.trim(),
    last_name: createForm.last_name.trim(),
    email: createForm.email.trim(),
    phone: createForm.phone.trim(),
    attributes: createForm.attributes,
  }
}

function buildEditPayload() {
  const payload = {
    first_name: editForm.first_name.trim(),
    last_name: editForm.last_name.trim(),
    email: editForm.email.trim(),
    phone: editForm.phone.trim(),
    attributes: editForm.attributes,
  }

  if (editForm.department_id) {
    payload.department_id = editForm.department_id
  }

  return payload
}

async function loadData(options = {}) {
  isLoading.value = true
  loadError.value = ''

  try {
    const [employeeResponse, departmentResponse] = await Promise.all([
      listEmployees({
        ...filters,
        page: page.value,
        rows: rows.value,
      }),
      loadAllPages(listDepartments),
    ])

    employees.value = Array.isArray(employeeResponse?.items) ? employeeResponse.items : []
    total.value = Number.isFinite(employeeResponse?.total) ? employeeResponse.total : employees.value.length
    page.value = Number.isFinite(employeeResponse?.page) ? employeeResponse.page : page.value
    rows.value = Number.isFinite(employeeResponse?.rowsPerPage) ? employeeResponse.rowsPerPage : rows.value
    departments.value = Array.isArray(departmentResponse) ? departmentResponse : []

    if (options.backfillEmptyPage && total.value > 0 && !employees.value.length && page.value > 1) {
      page.value -= 1
      await loadData()
    }
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    loadError.value = errorMessage(error, 'errors.employeesLoad')
  } finally {
    isLoading.value = false
  }
}

function applyFilters() {
  page.value = 1
  loadData()
}

function setPage(nextPage) {
  page.value = Math.min(Math.max(1, nextPage), totalPages.value)
  loadData()
}

function setRows(nextRows) {
  rows.value = nextRows
  page.value = 1
  loadData()
}

async function submitCreate() {
  const validation = validateForm(createForm)
  if (validation) {
    createFormError.value = validation
    return
  }

  if (!(await ensureCsrfToken())) {
    return
  }

  isSaving.value = true
  createFormError.value = ''

  try {
    await createEmployee(buildCreatePayload(), mutationOptions())
    notifySuccess(t('pages.employees.notifications.created'))
    resetCreateFlow()
    await router.push({ name: 'employees' })
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    createFormError.value = errorMessage(error, 'errors.employeesLoad')
    notifyError(t('pages.employees.notifications.createFailedTitle'), createFormError.value)
  } finally {
    isSaving.value = false
  }
}

async function submitEdit() {
  const validation = validateForm(editForm)
  if (validation) {
    editFormError.value = validation
    return
  }

  if (!selectedEmployee.value || !(await ensureCsrfToken())) {
    return
  }

  isSaving.value = true
  editFormError.value = ''

  try {
    const saved = await updateEmployee(selectedEmployee.value.id, buildEditPayload(), mutationOptions())
    notifySuccess(t('pages.employees.notifications.updated'))
    await loadData()
    openEditDrawer(saved)
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    editFormError.value = errorMessage(error, 'errors.employeesLoad')
    notifyError(t('pages.employees.notifications.saveFailedTitle'), editFormError.value)
  } finally {
    isSaving.value = false
  }
}

function openDeleteDialog(employee = selectedEmployee.value) {
  if (!employee) {
    return
  }

  selectedId.value = employee.id
  selectedEmployeeRecord.value = employee
  deleteDialogOpen.value = true
}

async function confirmDelete() {
  if (!selectedEmployee.value || !(await ensureCsrfToken())) {
    return
  }

  isDeleting.value = true

  try {
    await deleteEmployee(selectedEmployee.value.id, mutationOptions())
    notifySuccess(t('pages.employees.notifications.deleted'))
    deleteDialogOpen.value = false
    closeEditDrawer()
    await loadData({ backfillEmptyPage: true })
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    notifyError(t('pages.employees.notifications.deleteFailedTitle'), errorMessage(error, 'errors.employeesLoad'))
  } finally {
    isDeleting.value = false
  }
}

watch(() => route.name, (name) => {
  if (name === 'employees-new') {
    resetCreateFlow()
  } else {
    closeEditDrawer()
    loadData()
  }
})

onMounted(() => {
  loadData()
})
</script>

<template>
  <section class="resource-page">
    <template v-if="isCreateMode">
      <PageHeader
        :eyebrow="t('sections.people')"
        :title="t('pages.employees.createTitle')"
        :description="t('pages.employees.createDescription')"
      >
        <template #actions>
          <RouterLink class="ui-button ui-button--secondary" :to="{ name: 'employees' }">
            {{ t('pages.employees.backToList') }}
          </RouterLink>
        </template>
      </PageHeader>

      <UiAlert
        v-if="createFormError"
        variant="error"
        :title="t('pages.employees.formAttention')"
        :message="createFormError"
        dismissible
        @dismiss="createFormError = ''"
      />

      <section class="resource-editor-shell">
        <nav class="resource-stepper" :aria-label="t('pages.employees.stepsAria')">
          <button
            v-for="(step, index) in steps"
            :key="step"
            type="button"
            :class="{ 'is-active': activeStep === index }"
            @click="activeStep = index"
          >
            <span>{{ index + 1 }}</span>
            {{ step }}
          </button>
        </nav>

        <section v-if="activeStep === 0" class="resource-editor-panel">
          <header class="resource-panel-header resource-panel-header--plain">
            <div>
              <h2>{{ t('pages.employees.steps.basicInfo') }}</h2>
              <p>{{ t('pages.employees.basicInfoDescription') }}</p>
            </div>
          </header>
          <div class="resource-form-grid">
            <UiInput v-model="createForm.first_name" :label="t('pages.employees.labels.firstName')" />
            <UiInput v-model="createForm.last_name" :label="t('pages.employees.labels.lastName')" />
            <UiInput v-model="createForm.email" :label="t('common.labels.email')" />
            <UiInput v-model="createForm.phone" :label="t('common.labels.phone')" />
          </div>
          <section class="resource-subpanel">
              <h3>{{ t('common.labels.attributes') }}</h3>
            <AttributesEditor v-model="createForm.attributes" />
          </section>
        </section>

        <section v-else-if="activeStep === 1" class="resource-editor-panel">
          <header class="resource-panel-header resource-panel-header--plain">
            <div>
              <h2>{{ t('common.labels.department') }}</h2>
              <p>{{ t('pages.employees.departmentDescription') }}</p>
            </div>
          </header>
          <UiInput v-model="departmentSearch" :label="t('pages.employees.departmentSearch')" :placeholder="t('pages.employees.departmentSearchPlaceholder')" />
          <div class="resource-selection-list">
            <label class="resource-selection-row">
              <input v-model="createForm.department_id" type="radio" value="" />
              <span class="resource-selection-icon" aria-hidden="true">-</span>
              <span class="resource-selection-copy">
                <strong>{{ t('common.selection.noDepartment') }}</strong>
                <small>{{ t('pages.employees.noDepartmentHint') }}</small>
              </span>
            </label>
            <label
              v-for="department in filteredDepartments"
              :key="department.id"
              class="resource-selection-row"
            >
              <input v-model="createForm.department_id" type="radio" :value="department.id" />
              <span class="resource-selection-icon" aria-hidden="true">{{ department.label.slice(0, 1).toUpperCase() }}</span>
              <span class="resource-selection-copy">
                <strong>{{ department.label }}</strong>
                <small>{{ formatAttributes(department.attributes) }}</small>
              </span>
            </label>
          </div>
        </section>

        <section v-else class="resource-editor-panel">
          <header class="resource-panel-header resource-panel-header--plain">
            <div>
              <h2>{{ t('pages.employees.steps.review') }}</h2>
              <p>{{ t('pages.employees.reviewDescription') }}</p>
            </div>
          </header>
          <dl class="resource-review-list">
            <div><dt>{{ t('common.labels.fullName') }}</dt><dd>{{ [createForm.first_name, createForm.last_name].filter(Boolean).join(' ') || t('common.placeholder') }}</dd></div>
            <div><dt>{{ t('common.labels.email') }}</dt><dd>{{ createForm.email || t('common.placeholder') }}</dd></div>
            <div><dt>{{ t('common.labels.phone') }}</dt><dd>{{ createForm.phone || t('common.placeholder') }}</dd></div>
            <div><dt>{{ t('common.labels.department') }}</dt><dd>{{ selectedCreateDepartment?.label || t('common.selection.noDepartment') }}</dd></div>
            <div><dt>{{ t('common.labels.attributes') }}</dt><dd>{{ formatAttributes(createForm.attributes) }}</dd></div>
          </dl>
        </section>

        <footer class="resource-editor-actions">
          <RouterLink class="ui-button ui-button--ghost" :to="{ name: 'employees' }">{{ t('common.actions.cancel') }}</RouterLink>
          <UiButton variant="secondary" :disabled="activeStep === 0 || isSaving" @click="activeStep -= 1">
            {{ t('common.actions.back') }}
          </UiButton>
          <UiButton
            v-if="activeStep < steps.length - 1"
            variant="secondary"
            :disabled="isSaving"
            @click="activeStep += 1"
          >
            {{ t('common.actions.continue') }}
          </UiButton>
          <UiButton v-else :loading="isSaving" @click="submitCreate">
            {{ t('common.actions.createEmployee') }}
          </UiButton>
        </footer>
      </section>
    </template>

    <template v-else>
      <PageHeader
        :eyebrow="t('sections.people')"
        :title="t('pages.employees.title')"
        :description="t('pages.employees.description')"
      >
        <template #actions>
          <input
            ref="importFileInput"
            class="resource-file-input"
            type="file"
            accept=".csv,.txt,text/csv,text/plain"
            @change="handleImportFileChange"
          />
          <UiButton variant="secondary" :loading="isImporting" :disabled="isLoading" @click="openImportPicker">
            <Upload :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ isImporting ? t('common.states.uploading') : t('pages.employees.import.action') }}
          </UiButton>
          <RouterLink class="ui-button ui-button--primary" :to="{ name: 'employees-new' }">
            <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('common.actions.createEmployee') }}
          </RouterLink>
        </template>
      </PageHeader>

      <UiAlert
        v-if="importErrorMessage || importErrorItems.length"
        variant="error"
        :title="t('pages.employees.import.failedTitle')"
        :message="importErrorMessage || t('pages.employees.import.failedMessage')"
        :items="importErrorItems"
        dismissible
        @dismiss="clearImportError"
      />

      <section class="resource-filter-panel">
        <UiInput v-model="filters.fullName" :label="t('common.labels.fullName')" :placeholder="t('pages.employees.searchPlaceholder')" />
        <UiInput v-model="filters.email" :label="t('common.labels.email')" placeholder="employee@example.com" />
        <UiInput v-model="filters.phone" :label="t('common.labels.phone')" :placeholder="t('common.labels.phone')" />
        <label class="ui-field">
          <span class="ui-field-label">{{ t('common.labels.department') }}</span>
          <select v-model="filters.departmentId" class="ui-select">
            <option value="">{{ t('common.filters.allDepartments') }}</option>
            <option v-for="department in departments" :key="department.id" :value="department.id">
              {{ department.label }}
            </option>
          </select>
        </label>
        <div class="resource-filter-actions">
          <UiButton variant="secondary" :loading="isLoading" @click="applyFilters">
            {{ t('common.filters.apply') }}
          </UiButton>
          <UiButton variant="ghost" :disabled="isLoading" @click="clearFilters">
            {{ t('common.actions.clear') }}
          </UiButton>
        </div>
      </section>

      <UiAlert
        v-if="loadError"
        variant="error"
        :title="t('pages.employees.unavailable')"
        :message="loadError"
      >
        <UiButton variant="secondary" @click="loadData">{{ t('common.actions.retry') }}</UiButton>
      </UiAlert>

      <section class="resource-panel resource-panel--table">
        <header class="resource-panel-header">
          <div>
            <h2>{{ t('pages.employees.tableTitle') }}</h2>
            <p>{{ t('common.loadedCount', { count: employees.length }) }}</p>
          </div>
        </header>

        <div class="resource-table resource-table--compact resource-employee-table">
          <div class="resource-table-row resource-table-head">
            <span>{{ t('common.labels.fullName') }}</span>
            <span>{{ t('common.labels.email') }}</span>
            <span>{{ t('common.labels.phone') }}</span>
            <span>{{ t('common.labels.department') }}</span>
            <span>{{ t('common.labels.attributes') }}</span>
            <span>{{ t('pages.campaigns.columns.actions') }}</span>
          </div>
          <template v-if="isLoading">
            <div v-for="index in 4" :key="index" class="resource-table-row">
              <span v-for="cell in 6" :key="cell"><SkeletonBlock :rows="1" /></span>
            </div>
          </template>
          <template v-else-if="employees.length">
            <div
              v-for="employee in employees"
              :key="employee.id"
              class="resource-table-row"
            >
              <span><strong>{{ employeeName(employee) }}</strong></span>
              <span>{{ employee.email || t('common.placeholder') }}</span>
              <span>{{ employee.phone || t('common.placeholder') }}</span>
              <span>{{ departmentLabel(employee.department_id) }}</span>
              <span :title="formatAttributes(employee.attributes)">{{ formatAttributes(employee.attributes) }}</span>
              <span class="resource-row-actions">
                <IconButton :label="t('common.actions.edit')" variant="secondary" @click="openEditDrawer(employee)">
                  <Pencil :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
                <IconButton :label="t('common.actions.delete')" variant="danger" @click="openDeleteDialog(employee)">
                  <Trash2 :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
              </span>
            </div>
          </template>
          <div v-else class="resource-empty-row">
            {{ t('common.empty.noMatchingRecords') }}
          </div>
        </div>
        <ResourcePagination
          :page="page"
          :rows="rows"
          :total="total"
          :loaded-count="employees.length"
          :loading="isLoading"
          @update:page="setPage"
          @update:rows="setRows"
        />
      </section>

      <ResourceDrawer
        :open="editDrawerOpen"
        :title="t('pages.employees.editTitle')"
        :subtitle="t('pages.employees.basicInfoDescription')"
        @close="requestCloseDrawer"
      >
        <UiAlert
          v-if="editFormError"
          variant="error"
          :title="t('pages.employees.formAttention')"
          :message="editFormError"
          dismissible
          @dismiss="editFormError = ''"
        />

        <div class="resource-form-grid">
          <UiInput v-model="editForm.first_name" :label="t('pages.employees.labels.firstName')" />
          <UiInput v-model="editForm.last_name" :label="t('pages.employees.labels.lastName')" />
          <UiInput v-model="editForm.email" :label="t('common.labels.email')" />
          <UiInput v-model="editForm.phone" :label="t('common.labels.phone')" />
          <label class="ui-field resource-field-wide">
            <span class="ui-field-label">{{ t('common.labels.department') }}</span>
            <select v-model="editForm.department_id" class="ui-select">
              <option value="">{{ t('common.selection.noDepartment') }}</option>
              <option v-for="department in departments" :key="department.id" :value="department.id">
                {{ department.label }}
              </option>
            </select>
              <span v-if="!editForm.department_id" class="ui-field-helper">
                {{ t('pages.employees.emptyDepartmentUpdateHint') }}
              </span>
            </label>
          </div>

          <section class="resource-subpanel">
            <h3>{{ t('common.labels.attributes') }}</h3>
            <AttributesEditor v-model="editForm.attributes" />
          </section>

          <template #footer>
            <UiButton variant="danger" :disabled="isSaving" @click="openDeleteDialog()">
              {{ t('common.actions.delete') }}
            </UiButton>
            <UiButton variant="secondary" :disabled="isSaving" @click="requestCloseDrawer">
              {{ t('common.actions.cancel') }}
            </UiButton>
            <UiButton :loading="isSaving" @click="submitEdit">
              {{ t('common.actions.saveChanges') }}
            </UiButton>
          </template>
        </ResourceDrawer>
    </template>

    <ConfirmDialog
      :open="deleteDialogOpen"
      :title="t('pages.employees.delete.title')"
      :message="t('pages.employees.delete.message')"
      :confirm-label="t('common.actions.delete')"
      :loading-label="t('common.states.deleting')"
      :cancel-label="t('common.actions.cancel')"
      variant="danger"
      :loading="isDeleting"
      @cancel="deleteDialogOpen = false"
      @confirm="confirmDelete"
    />

    <ConfirmDialog
      :open="discardDialogOpen"
      :title="t('pages.employees.discard.title')"
      :message="t('common.resource.discardMessage')"
      :confirm-label="t('common.actions.discardChanges')"
      :cancel-label="t('common.actions.continueEditing')"
      variant="danger"
      @cancel="discardDialogOpen = false"
      @confirm="closeEditDrawer"
    />
  </section>
</template>

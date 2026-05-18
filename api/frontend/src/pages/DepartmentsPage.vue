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
import {
  createDepartment,
  deleteDepartment,
  importDepartments,
  listDepartments,
  updateDepartment,
} from '../resources/departments'
import { listEmployees } from '../resources/employees'
import { errorMessage, formatAttributes, importErrorInfo, nullableId } from '../utils/resourceFormat'
import { loadAllPages } from '../utils/resourcePagination'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const steps = computed(() => [
  t('pages.departments.steps.basicInfo'),
  t('pages.departments.steps.review'),
])

const departments = ref([])
const employees = ref([])
const page = ref(1)
const rows = ref(10)
const total = ref(0)
const selectedId = ref('')
const selectedDepartmentRecord = ref(null)
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
const originalEditSnapshot = ref('')
const importFileInput = ref(null)
const filters = reactive({
  label: '',
})
const editForm = reactive(makeDepartmentForm())
const createForm = reactive(makeDepartmentForm())

const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()

const isCreateMode = computed(() => route.name === 'departments-new')
const selectedDepartment = computed(() => selectedDepartmentRecord.value || departments.value.find((department) => department.id === selectedId.value) || null)
const employeeCounts = computed(() => employees.value.reduce((acc, employee) => {
  const departmentId = nullableId(employee.department_id)
  if (departmentId) {
    acc[departmentId] = (acc[departmentId] || 0) + 1
  }
  return acc
}, {}))
const editFormDirty = computed(() => {
  if (!editDrawerOpen.value || !originalEditSnapshot.value) {
    return false
  }

  return originalEditSnapshot.value !== stableStringify(departmentFormSnapshot(editForm))
})
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / rows.value)))

function makeDepartmentForm() {
  return {
    label: '',
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

function departmentFormSnapshot(form) {
  return {
    label: form.label || '',
    attributes: Object.keys(form.attributes || {}).sort().reduce((acc, key) => {
      acc[key] = form.attributes[key]
      return acc
    }, {}),
  }
}

function assignDepartmentForm(target, source) {
  target.label = source.label || ''
  target.attributes = { ...(source.attributes || {}) }
}

function resetCreateFlow() {
  activeStep.value = 0
  createFormError.value = ''
  assignDepartmentForm(createForm, makeDepartmentForm())
}

function clearFilters() {
  filters.label = ''
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
  return t('pages.departments.import.rowError', {
    row: rowError.row || t('common.placeholder'),
    field: rowError.field ? `, ${rowError.field}` : '',
    error: rowError.error || t('pages.departments.import.failedMessage'),
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
    await importDepartments(file, mutationOptions())
    notifySuccess(t('pages.departments.import.success'))
    await loadData()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    const importError = importErrorInfo(error, formatImportRowError)
    importErrorMessage.value = importError.message || t('pages.departments.import.failedMessage')
    importErrorItems.value = importError.items
  } finally {
    isImporting.value = false
    resetImportInput()
  }
}

function openEditDrawer(department) {
  selectedId.value = department.id
  selectedDepartmentRecord.value = department
  assignDepartmentForm(editForm, department)
  editFormError.value = ''
  originalEditSnapshot.value = stableStringify(departmentFormSnapshot(editForm))
  editDrawerOpen.value = true
}

function closeEditDrawer() {
  editDrawerOpen.value = false
  discardDialogOpen.value = false
  selectedId.value = ''
  selectedDepartmentRecord.value = null
  editFormError.value = ''
  originalEditSnapshot.value = ''
  assignDepartmentForm(editForm, makeDepartmentForm())
}

function requestCloseDrawer() {
  if (editFormDirty.value) {
    discardDialogOpen.value = true
    return
  }

  closeEditDrawer()
}

function validateForm(form) {
  if (!form.label.trim()) {
    return t('pages.departments.validation.labelRequired')
  }

  return ''
}

function buildPayload(form) {
  return {
    label: form.label.trim(),
    attributes: form.attributes,
  }
}

async function loadData(options = {}) {
  isLoading.value = true
  loadError.value = ''

  try {
    const [departmentResponse, employeeResponse] = await Promise.all([
      listDepartments({
        ...filters,
        page: page.value,
        rows: rows.value,
      }),
      loadAllPages(listEmployees),
    ])

    departments.value = Array.isArray(departmentResponse?.items) ? departmentResponse.items : []
    total.value = Number.isFinite(departmentResponse?.total) ? departmentResponse.total : departments.value.length
    page.value = Number.isFinite(departmentResponse?.page) ? departmentResponse.page : page.value
    rows.value = Number.isFinite(departmentResponse?.rowsPerPage) ? departmentResponse.rowsPerPage : rows.value
    employees.value = Array.isArray(employeeResponse) ? employeeResponse : []

    if (options.backfillEmptyPage && total.value > 0 && !departments.value.length && page.value > 1) {
      page.value -= 1
      await loadData()
    }
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    loadError.value = errorMessage(error, 'errors.departmentsLoad')
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
    await createDepartment(buildPayload(createForm), mutationOptions())
    notifySuccess(t('pages.departments.notifications.created'))
    resetCreateFlow()
    await router.push({ name: 'departments' })
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    createFormError.value = errorMessage(error, 'errors.departmentsLoad')
    notifyError(t('pages.departments.notifications.createFailedTitle'), createFormError.value)
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

  if (!selectedDepartment.value || !(await ensureCsrfToken())) {
    return
  }

  isSaving.value = true
  editFormError.value = ''

  try {
    const saved = await updateDepartment(selectedDepartment.value.id, buildPayload(editForm), mutationOptions())
    notifySuccess(t('pages.departments.notifications.updated'))
    await loadData()
    openEditDrawer(saved)
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    editFormError.value = errorMessage(error, 'errors.departmentsLoad')
    notifyError(t('pages.departments.notifications.saveFailedTitle'), editFormError.value)
  } finally {
    isSaving.value = false
  }
}

function openDeleteDialog(department = selectedDepartment.value) {
  if (!department) {
    return
  }

  selectedId.value = department.id
  selectedDepartmentRecord.value = department
  deleteDialogOpen.value = true
}

async function confirmDelete() {
  if (!selectedDepartment.value || !(await ensureCsrfToken())) {
    return
  }

  isDeleting.value = true

  try {
    await deleteDepartment(selectedDepartment.value.id, mutationOptions())
    notifySuccess(t('pages.departments.notifications.deleted'))
    deleteDialogOpen.value = false
    closeEditDrawer()
    await loadData({ backfillEmptyPage: true })
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    notifyError(t('pages.departments.notifications.deleteFailedTitle'), errorMessage(error, 'errors.departmentsLoad'))
  } finally {
    isDeleting.value = false
  }
}

watch(() => route.name, (name) => {
  if (name === 'departments-new') {
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
        :title="t('pages.departments.createTitle')"
        :description="t('pages.departments.createDescription')"
      >
        <template #actions>
          <RouterLink class="ui-button ui-button--secondary" :to="{ name: 'departments' }">
            {{ t('pages.departments.backToList') }}
          </RouterLink>
        </template>
      </PageHeader>

      <UiAlert
        v-if="createFormError"
        variant="error"
        :title="t('pages.departments.formAttention')"
        :message="createFormError"
        dismissible
        @dismiss="createFormError = ''"
      />

      <section class="resource-editor-shell">
        <nav class="resource-stepper" :aria-label="t('pages.departments.stepsAria')">
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
              <h2>{{ t('pages.departments.steps.basicInfo') }}</h2>
              <p>{{ t('pages.departments.basicInfoDescription') }}</p>
            </div>
          </header>
          <div class="resource-form-grid resource-form-grid--single">
            <UiInput v-model="createForm.label" :label="t('common.labels.name')" :placeholder="t('pages.departments.labelPlaceholder')" />
          </div>
          <section class="resource-subpanel">
            <h3>{{ t('common.labels.attributes') }}</h3>
            <AttributesEditor v-model="createForm.attributes" />
          </section>
        </section>

        <section v-else class="resource-editor-panel">
          <header class="resource-panel-header resource-panel-header--plain">
            <div>
              <h2>{{ t('pages.departments.steps.review') }}</h2>
              <p>{{ t('pages.departments.reviewDescription') }}</p>
            </div>
          </header>
          <dl class="resource-review-list">
            <div><dt>{{ t('common.labels.name') }}</dt><dd>{{ createForm.label || t('common.placeholder') }}</dd></div>
            <div><dt>{{ t('common.labels.attributes') }}</dt><dd>{{ formatAttributes(createForm.attributes) }}</dd></div>
          </dl>
        </section>

        <footer class="resource-editor-actions">
          <RouterLink class="ui-button ui-button--ghost" :to="{ name: 'departments' }">{{ t('common.actions.cancel') }}</RouterLink>
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
            {{ t('common.actions.createDepartment') }}
          </UiButton>
        </footer>
      </section>
    </template>

    <template v-else>
      <PageHeader
        :eyebrow="t('sections.people')"
        :title="t('pages.departments.title')"
        :description="t('pages.departments.description')"
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
            {{ isImporting ? t('common.states.uploading') : t('pages.departments.import.action') }}
          </UiButton>
          <RouterLink class="ui-button ui-button--primary" :to="{ name: 'departments-new' }">
            <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('common.actions.createDepartment') }}
          </RouterLink>
        </template>
      </PageHeader>

      <UiAlert
        v-if="importErrorMessage || importErrorItems.length"
        variant="error"
        :title="t('pages.departments.import.failedTitle')"
        :message="importErrorMessage || t('pages.departments.import.failedMessage')"
        :items="importErrorItems"
        dismissible
        @dismiss="clearImportError"
      />

      <section class="resource-filter-panel resource-filter-panel--compact">
        <UiInput v-model="filters.label" :label="t('common.labels.name')" :placeholder="t('pages.departments.searchPlaceholder')" />
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
        :title="t('pages.departments.unavailable')"
        :message="loadError"
      >
        <UiButton variant="secondary" @click="loadData">{{ t('common.actions.retry') }}</UiButton>
      </UiAlert>

      <section class="resource-panel resource-panel--table">
        <header class="resource-panel-header">
          <div>
            <h2>{{ t('pages.departments.tableTitle') }}</h2>
            <p>{{ t('common.loadedCount', { count: departments.length }) }}</p>
          </div>
        </header>

        <div class="resource-table resource-table--compact resource-department-table">
          <div class="resource-table-row resource-table-head">
            <span>{{ t('common.labels.name') }}</span>
            <span>{{ t('pages.departments.employeeCount') }}</span>
            <span>{{ t('common.labels.attributes') }}</span>
            <span>{{ t('pages.campaigns.columns.actions') }}</span>
          </div>
          <template v-if="isLoading">
            <div v-for="index in 4" :key="index" class="resource-table-row">
              <span v-for="cell in 4" :key="cell"><SkeletonBlock :rows="1" /></span>
            </div>
          </template>
          <template v-else-if="departments.length">
            <div
              v-for="department in departments"
              :key="department.id"
              class="resource-table-row"
            >
              <span><strong>{{ department.label }}</strong></span>
              <span>{{ employeeCounts[department.id] || 0 }}</span>
              <span :title="formatAttributes(department.attributes)">{{ formatAttributes(department.attributes) }}</span>
              <span class="resource-row-actions">
                <IconButton :label="t('common.actions.edit')" variant="secondary" @click="openEditDrawer(department)">
                  <Pencil :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
                <IconButton :label="t('common.actions.delete')" variant="danger" @click="openDeleteDialog(department)">
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
          :loaded-count="departments.length"
          :loading="isLoading"
          @update:page="setPage"
          @update:rows="setRows"
        />
      </section>

      <ResourceDrawer
        :open="editDrawerOpen"
        :title="t('pages.departments.editTitle')"
        :subtitle="t('pages.departments.basicInfoDescription')"
        @close="requestCloseDrawer"
      >
        <UiAlert
          v-if="editFormError"
          variant="error"
          :title="t('pages.departments.formAttention')"
          :message="editFormError"
          dismissible
          @dismiss="editFormError = ''"
        />

        <div class="resource-form-grid resource-form-grid--single">
          <UiInput v-model="editForm.label" :label="t('common.labels.name')" :placeholder="t('pages.departments.labelPlaceholder')" />
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
      :title="t('pages.departments.delete.title')"
      :message="t('pages.departments.delete.message')"
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
      :title="t('pages.departments.discard.title')"
      :message="t('common.resource.discardMessage')"
      :confirm-label="t('common.actions.discardChanges')"
      :cancel-label="t('common.actions.continueEditing')"
      variant="danger"
      @cancel="discardDialogOpen = false"
      @confirm="closeEditDrawer"
    />
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AttachmentPickerPanel from '../components/ui/AttachmentPickerPanel.vue'
import AttributesEditor from '../components/ui/AttributesEditor.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiInput from '../components/ui/UiInput.vue'
import { useNotifications } from '../composables/useNotifications'
import { useResourceActions } from '../composables/useResourceActions'
import { getAttachmentUrl } from '../resources/attachments'
import {
  createOrganization,
  getOrganization,
  isOrganizationNotFoundError,
  updateOrganization,
} from '../resources/organization'
import {
  errorMessage,
  formatTechnicalId,
  nullableId,
} from '../utils/resourceFormat'

const BRANDING_ATTACHMENT_TYPES = ['.png', '.jpg', '.jpeg', '.gif', '.webp']

const organization = ref(null)
const selectedBrandingAttachment = ref(null)
const isLoading = ref(false)
const isSaving = ref(false)
const loadError = ref('')
const formError = ref('')
const createMode = ref(false)
const form = reactive({
  label: '',
  attachment_id: '',
  attributes: {},
})

const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()
const { t } = useI18n()

const selectedAttachment = computed(() => (
  nullableId(selectedBrandingAttachment.value?.id) === form.attachment_id
    ? selectedBrandingAttachment.value
    : null
))
const selectedAttachmentUrl = computed(() => selectedAttachment.value ? getAttachmentUrl(selectedAttachment.value.id) : '')
const brandingTypesLabel = computed(() => BRANDING_ATTACHMENT_TYPES.join(', '))

function resetFormState() {
  organization.value = null
  createMode.value = true
  selectedBrandingAttachment.value = null
  form.label = ''
  form.attachment_id = ''
  form.attributes = {}
}

function fillForm(org) {
  organization.value = org
  createMode.value = false
  form.label = org.label || ''
  form.attachment_id = nullableId(org.attachment_id)
  form.attributes = { ...(org.attributes || {}) }

  if (nullableId(selectedBrandingAttachment.value?.id) !== form.attachment_id) {
    selectedBrandingAttachment.value = null
  }
}

async function loadData() {
  isLoading.value = true
  loadError.value = ''
  formError.value = ''

  try {
    const org = await getOrganization()
    fillForm(org)
  } catch (error) {
    if (isOrganizationNotFoundError(error)) {
      resetFormState()
    } else if (await handleAuthError(error)) {
      return
    } else {
      loadError.value = errorMessage(error, 'errors.organizationLoad')
    }
  } finally {
    isLoading.value = false
  }
}

function buildPayload() {
  return {
    label: form.label.trim(),
    attachment_id: form.attachment_id,
    attributes: form.attributes,
  }
}

function trackBrandingSelection(attachment) {
  selectedBrandingAttachment.value = attachment || null
  form.attachment_id = nullableId(attachment?.id) || ''
}

function handleBrandingUploaded(attachment) {
  trackBrandingSelection(attachment)
}

async function saveOrganization() {
  if (!form.label.trim()) {
    formError.value = t('pages.organization.validation.labelRequired')
    return
  }

  if (!(await ensureCsrfToken())) {
    return
  }

  isSaving.value = true
  formError.value = ''
  const wasCreate = createMode.value

  try {
    const saved = wasCreate
      ? await createOrganization(buildPayload(), mutationOptions())
      : await updateOrganization(buildPayload(), mutationOptions())

    fillForm(saved)
    notifySuccess(wasCreate ? t('pages.organization.notifications.created') : t('pages.organization.notifications.updated'))
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.organizationSave')
    notifyError(t('pages.organization.notifications.saveFailedTitle'), formError.value)
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <section class="resource-page">
    <PageHeader
      :eyebrow="t('sections.settings')"
      :title="t('pages.organization.title')"
      :description="t('pages.organization.description')"
    />

    <UiAlert
      v-if="createMode && !isLoading"
      variant="info"
      :title="t('pages.organization.notCreatedTitle')"
      :message="t('pages.organization.notCreatedMessage')"
    />

    <UiAlert
      v-if="loadError"
      variant="error"
      :title="t('pages.organization.unavailable')"
      :message="loadError"
    >
      <UiButton variant="secondary" @click="loadData">{{ t('common.actions.retry') }}</UiButton>
    </UiAlert>

    <section class="resource-panel organization-settings-panel">
      <template v-if="isLoading">
        <SkeletonBlock :rows="7" />
      </template>

      <template v-else>
        <UiAlert
          v-if="formError"
          variant="error"
          :title="t('pages.organization.formAttention')"
          :message="formError"
          dismissible
          @dismiss="formError = ''"
        />

        <section class="resource-subpanel organization-settings-section">
          <header class="organization-section-header">
            <div>
              <h3>{{ t('pages.organization.profileTitle') }}</h3>
              <p>{{ t('pages.organization.profileDescription') }}</p>
            </div>
          </header>

          <div class="resource-form-grid">
            <UiInput v-model="form.label" :label="t('common.labels.name')" :placeholder="t('pages.organization.labelPlaceholder')" />
            <UiInput
              :model-value="organization?.id ? formatTechnicalId(organization.id) : t('common.placeholder')"
              :label="t('common.labels.organizationId')"
              disabled
            />
          </div>
        </section>

        <section class="resource-subpanel organization-settings-section">
          <header class="organization-section-header">
            <div>
              <h3>{{ t('pages.organization.branding.title') }}</h3>
              <p>{{ t('pages.organization.branding.description') }}</p>
            </div>
          </header>

          <div class="organization-branding-image-card">
            <template v-if="selectedAttachment">
              <img
                :src="selectedAttachmentUrl"
                :alt="t('pages.organization.brandingPreviewAlt', { label: selectedAttachment.label })"
              />
            </template>
            <p v-else>{{ t('pages.organization.branding.empty') }}</p>
          </div>

          <AttachmentPickerPanel
            v-model="form.attachment_id"
            selection-mode="single"
            :title="t('pages.organization.branding.select')"
            :empty-text="t('pages.organization.branding.empty')"
            :allowed-types="BRANDING_ATTACHMENT_TYPES"
            :unsupported-type-message="t('pages.organization.branding.unsupportedType', { types: brandingTypesLabel })"
            :upload-success-message="t('pages.organization.branding.uploadSuccess')"
            show-header-clear-action
            :show-selected-summary="false"
            :show-required-vars-column="false"
            restrict-to-allowed-types
            show-download-action
            show-upload-action
            @selection-change="trackBrandingSelection"
            @uploaded="handleBrandingUploaded"
          />

          <p class="resource-muted">{{ t('pages.organization.branding.saveHint') }}</p>
        </section>

        <section class="resource-subpanel organization-settings-section">
          <header class="organization-section-header">
            <div>
              <h3>{{ t('common.labels.attributes') }}</h3>
              <p>{{ t('pages.organization.attributesDescription') }}</p>
            </div>
          </header>
          <AttributesEditor v-model="form.attributes" />
        </section>

        <footer class="resource-form-actions organization-settings-actions">
          <span />
          <UiButton variant="secondary" :disabled="isSaving" @click="loadData">
            {{ t('common.actions.reset') }}
          </UiButton>
          <UiButton :loading="isSaving" @click="saveOrganization">
            {{ isSaving ? t('common.states.saving') : createMode ? t('pages.organization.createAction') : t('common.actions.saveChanges') }}
          </UiButton>
        </footer>
      </template>
    </section>

  </section>
</template>

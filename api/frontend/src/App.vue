<script setup>
import { computed } from 'vue'
import AppSidebar from './components/AppSidebar.vue'
import DirectorySection from './components/DirectorySection.vue'
import EventsSection from './components/EventsSection.vue'
import MessagesSection from './components/MessagesSection.vue'
import OrganizationSection from './components/OrganizationSection.vue'
import RoutesSection from './components/RoutesSection.vue'
import TopBar from './components/TopBar.vue'
import TracePanel from './components/TracePanel.vue'
import { useApiConsole } from './composables/useApiConsole'
import { endpointGroups, navItems } from './data/navigation'

const consoleState = useApiConsole()

const {
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
} = consoleState

const activeTitle = computed(() => {
  const item = navItems.find((nav) => nav.key === activeSection.value)
  return item?.label ?? 'ClickSafe API'
})
</script>

<template>
  <div class="app-shell">
    <AppSidebar
      v-model:active-section="activeSection"
      v-model:api-base-url="apiBaseUrl"
      :api-base-label="apiBaseLabel"
      :nav-items="navItems"
    />

    <main class="workspace">
      <TopBar
        :active-title="activeTitle"
        :loading="loading"
        @bootstrap="bootstrapDemo"
        @refresh="refreshAll"
      />

      <el-alert
        v-if="lastError"
        class="error-alert"
        :title="lastError.title"
        :description="lastError.message"
        type="error"
        show-icon
        :closable="true"
        @close="clearLastError"
      />

      <MessagesSection
        v-if="activeSection === 'messages'"
        v-model:selected-message-id="selectedMessageId"
        v-model:selected-employee-id="selectedEmployeeId"
        v-model:template-text="templateText"
        :create-message="createMessage"
        :delete-message="deleteMessage"
        :employee-result="employeeResult"
        :loading="loading"
        :load-messages="loadMessages"
        :message-form="messageForm"
        :message-query="messageQuery"
        :message-result="messageResult"
        :on-template-file-change="onTemplateFileChange"
        :on-template-file-remove="onTemplateFileRemove"
        :raw-template-content="rawTemplateContent"
        :read-template-content="readTemplateContent"
        :render-message="renderMessage"
        :rendered-content="renderedContent"
        :selected-employee="selectedEmployee"
        :selected-message="selectedMessage"
        :short-id="shortId"
        :upload-template-file="uploadTemplateFile"
        :upload-template-text="uploadTemplateText"
      />

      <DirectorySection
        v-else-if="activeSection === 'directory'"
        v-model:selected-employee-id="selectedEmployeeId"
        :create-department="createDepartment"
        :create-employee="createEmployee"
        :delete-department="deleteDepartment"
        :delete-employee="deleteEmployee"
        :department-form="departmentForm"
        :department-query="departmentQuery"
        :department-result="departmentResult"
        :employee-form="employeeForm"
        :employee-query="employeeQuery"
        :employee-result="employeeResult"
        :loading="loading"
        :load-departments="loadDepartments"
        :load-employees="loadEmployees"
        :short-id="shortId"
      />

      <OrganizationSection
        v-else-if="activeSection === 'organization'"
        :load-organization="loadOrganization"
        :loading="loading"
        :organization="organization"
        :organization-form="organizationForm"
        :pretty="pretty"
        :save-organization="saveOrganization"
      />

      <EventsSection
        v-else-if="activeSection === 'events'"
        :event-form="eventForm"
        :loading="loading"
        :publish-event="publishEvent"
        :selected-employee-id="selectedEmployeeId"
      />

      <RoutesSection
        v-else
        :endpoint-groups="endpointGroups"
      />

      <TracePanel
        :traces="traces"
        @clear="clearTraces"
      />
    </main>
  </div>
</template>

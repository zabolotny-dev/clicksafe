<script setup>
import { computed } from 'vue'
import AppSidebar from './components/AppSidebar.vue'
import CampaignsSection from './components/CampaignsSection.vue'
import DirectorySection from './components/DirectorySection.vue'
import LandingsSection from './components/LandingsSection.vue'
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
} = consoleState

const activeTitle = computed(() => {
  const item = navItems.find((nav) => nav.key === activeSection.value)
  return item?.label ?? 'ClickSafe Admin'
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

      <CampaignsSection
        v-if="activeSection === 'campaigns'"
        v-model:selected-campaign-id="selectedCampaignId"
        v-model:selected-target-id="selectedTargetId"
        :auto-distribute-targets="autoDistributeTargets"
        :campaign-form="campaignForm"
        :campaign-query="campaignQuery"
        :campaign-result="campaignResult"
        :campaign-statuses="campaignStatuses"
        :change-campaign-status="changeCampaignStatus"
        :create-target="createTarget"
        :delete-campaign="deleteCampaign"
        :delete-campaign-targets="deleteCampaignTargets"
        :delete-target="deleteTarget"
        :edit-campaign="editCampaign"
        :employee-name="employeeName"
        :employee-result="employeeResult"
        :format-date="formatDate"
        :landing-by-id="landingByID"
        :landing-result="landingResult"
        :load-campaigns="loadCampaigns"
        :load-targets="loadTargets"
        :loading="loading"
        :message-by-id="messageByID"
        :message-result="messageResult"
        :render-landing="renderLanding"
        :render-message="renderMessage"
        :rendered-content="renderedContent"
        :rendered-landing-content="renderedLandingContent"
        :reset-campaign-form="resetCampaignForm"
        :save-campaign="saveCampaign"
        :selected-campaign="selectedCampaign"
        :selected-target="selectedTarget"
        :short-id="shortId"
        :target-form="targetForm"
        :target-name="targetName"
        :target-schedule-form="targetScheduleForm"
        :target-statuses="targetStatuses"
        :update-target-schedule="updateTargetSchedule"
        :visit-link="visitLink"
        :vtarget-query="vtargetQuery"
        :vtarget-result="vtargetResult"
      />

      <MessagesSection
        v-else-if="activeSection === 'messages'"
        v-model:selected-message-id="selectedMessageId"
        v-model:selected-target-id="selectedTargetId"
        v-model:template-text="templateText"
        :delete-message="deleteMessage"
        :edit-message="editMessage"
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
        :reset-message-form="resetMessageForm"
        :save-message="saveMessage"
        :selected-message="selectedMessage"
        :selected-target="selectedTarget"
        :short-id="shortId"
        :target-name="targetName"
        :upload-template-file="uploadTemplateFile"
        :upload-template-text="uploadTemplateText"
        :vtarget-result="vtargetResult"
      />

      <LandingsSection
        v-else-if="activeSection === 'landings'"
        v-model:selected-landing-id="selectedLandingId"
        v-model:selected-target-id="selectedTargetId"
        v-model:landing-template-text="landingTemplateText"
        :delete-landing="deleteLanding"
        :edit-landing="editLanding"
        :landing-form="landingForm"
        :landing-query="landingQuery"
        :landing-result="landingResult"
        :loading="loading"
        :load-landings="loadLandings"
        :on-landing-file-change="onLandingFileChange"
        :on-landing-file-remove="onLandingFileRemove"
        :raw-landing-content="rawLandingContent"
        :read-landing-content="readLandingContent"
        :render-landing="renderLanding"
        :rendered-landing-content="renderedLandingContent"
        :reset-landing-form="resetLandingForm"
        :save-landing="saveLanding"
        :selected-landing="selectedLanding"
        :selected-target="selectedTarget"
        :short-id="shortId"
        :target-name="targetName"
        :upload-landing-file="uploadLandingFile"
        :upload-landing-text="uploadLandingText"
        :vtarget-result="vtargetResult"
      />

      <DirectorySection
        v-else-if="activeSection === 'directory'"
        v-model:selected-employee-id="selectedEmployeeId"
        :delete-department="deleteDepartment"
        :delete-employee="deleteEmployee"
        :department-by-id="departmentByID"
        :department-form="departmentForm"
        :department-query="departmentQuery"
        :department-result="departmentResult"
        :edit-department="editDepartment"
        :edit-employee="editEmployee"
        :employee-form="employeeForm"
        :employee-query="employeeQuery"
        :employee-result="employeeResult"
        :loading="loading"
        :load-departments="loadDepartments"
        :load-employees="loadEmployees"
        :reset-department-form="resetDepartmentForm"
        :reset-employee-form="resetEmployeeForm"
        :save-department="saveDepartment"
        :save-employee="saveEmployee"
        :short-id="shortId"
      />

      <OrganizationSection
        v-else-if="activeSection === 'organization'"
        :delete-organization-logo="deleteOrganizationLogo"
        :load-organization="loadOrganization"
        :loading="loading"
        :on-logo-file-change="onLogoFileChange"
        :on-logo-file-remove="onLogoFileRemove"
        :organization="organization"
        :organization-form="organizationForm"
        :pretty="pretty"
        :save-organization="saveOrganization"
        :upload-organization-logo="uploadOrganizationLogo"
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

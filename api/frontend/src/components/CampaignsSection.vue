<script setup>
import {
  Delete,
  Edit,
  Link,
  MagicStick,
  Plus,
  Refresh,
  Select,
  Close,
  VideoPause,
  VideoPlay,
  Warning,
} from '@element-plus/icons-vue'

const selectedCampaignId = defineModel('selectedCampaignId', { type: String, required: true })
const selectedTargetId = defineModel('selectedTargetId', { type: String, required: true })

defineProps({
  autoDistributeTargets: {
    type: Function,
    required: true,
  },
  campaignForm: {
    type: Object,
    required: true,
  },
  campaignQuery: {
    type: Object,
    required: true,
  },
  campaignResult: {
    type: Object,
    required: true,
  },
  campaignStatuses: {
    type: Array,
    required: true,
  },
  changeCampaignStatus: {
    type: Function,
    required: true,
  },
  createTarget: {
    type: Function,
    required: true,
  },
  deleteCampaign: {
    type: Function,
    required: true,
  },
  deleteCampaignTargets: {
    type: Function,
    required: true,
  },
  deleteTarget: {
    type: Function,
    required: true,
  },
  editCampaign: {
    type: Function,
    required: true,
  },
  employeeName: {
    type: Function,
    required: true,
  },
  employeeResult: {
    type: Object,
    required: true,
  },
  formatDate: {
    type: Function,
    required: true,
  },
  landingById: {
    type: Function,
    required: true,
  },
  landingResult: {
    type: Object,
    required: true,
  },
  loadCampaigns: {
    type: Function,
    required: true,
  },
  loadTargets: {
    type: Function,
    required: true,
  },
  loading: {
    type: Object,
    required: true,
  },
  messageById: {
    type: Function,
    required: true,
  },
  messageResult: {
    type: Object,
    required: true,
  },
  renderLanding: {
    type: Function,
    required: true,
  },
  renderMessage: {
    type: Function,
    required: true,
  },
  renderedContent: {
    type: String,
    default: '',
  },
  renderedLandingContent: {
    type: String,
    default: '',
  },
  resetCampaignForm: {
    type: Function,
    required: true,
  },
  saveCampaign: {
    type: Function,
    required: true,
  },
  selectedCampaign: {
    type: Object,
    default: null,
  },
  selectedTarget: {
    type: Object,
    default: null,
  },
  shortId: {
    type: Function,
    required: true,
  },
  targetForm: {
    type: Object,
    required: true,
  },
  targetName: {
    type: Function,
    required: true,
  },
  targetScheduleForm: {
    type: Object,
    required: true,
  },
  targetStatuses: {
    type: Array,
    required: true,
  },
  updateTargetSchedule: {
    type: Function,
    required: true,
  },
  visitLink: {
    type: Function,
    required: true,
  },
  vtargetQuery: {
    type: Object,
    required: true,
  },
  vtargetResult: {
    type: Object,
    required: true,
  },
})

function campaignTag(status) {
  return {
    ACTIVE: 'success',
    PAUSED: 'warning',
    CANCELED: 'danger',
    COMPLETED: 'info',
  }[status] ?? 'primary'
}

function targetTag(status) {
  return {
    SENT: 'success',
    OPENED: 'success',
    CLICKED: 'warning',
    SUBMITTED: 'danger',
    FAILED: 'danger',
  }[status] ?? 'info'
}
</script>

<template>
  <section class="section-stack">
    <div class="grid two-columns wide-left">
      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">GET /campaign</div>
            <h2>Кампании</h2>
          </div>
          <el-button :icon="Refresh" :loading="loading.campaigns" @click="loadCampaigns" />
        </div>

        <div class="toolbar">
          <el-input v-model="campaignQuery.label" placeholder="label" clearable />
          <el-select v-model="campaignQuery.status" placeholder="status" clearable>
            <el-option
              v-for="status in campaignStatuses"
              :key="status"
              :label="status"
              :value="status"
            />
          </el-select>
          <el-button :icon="Select" type="primary" @click="loadCampaigns">Фильтр</el-button>
        </div>

        <el-table
          class="data-table"
          :data="campaignResult.items"
          highlight-current-row
          :row-class-name="({ row }) => row.id === selectedCampaignId ? 'selected-row' : ''"
          @row-click="editCampaign"
        >
          <el-table-column prop="label" label="Label" min-width="180" />
          <el-table-column label="Status" width="118">
            <template #default="{ row }">
              <el-tag :type="campaignTag(row.status)" effect="plain">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="Message" min-width="150">
            <template #default="{ row }">{{ messageById(row.message_id)?.label || 'не выбрано' }}</template>
          </el-table-column>
          <el-table-column label="Landing" min-width="150">
            <template #default="{ row }">{{ landingById(row.landing_id)?.label || 'не выбрано' }}</template>
          </el-table-column>
          <el-table-column label="Education" min-width="150">
            <template #default="{ row }">{{ landingById(row.education_id)?.label || 'не выбрано' }}</template>
          </el-table-column>
          <el-table-column prop="domain" label="Domain" min-width="170" />
          <el-table-column label="Actions" width="238" fixed="right">
            <template #default="{ row }">
              <el-button
                v-if="row.status === 'DRAFT' || row.status === 'PAUSED'"
                :icon="VideoPlay"
                size="small"
                type="success"
                @click.stop="changeCampaignStatus(row.id, 'start')"
              />
              <el-button
                v-if="row.status === 'ACTIVE'"
                :icon="VideoPause"
                size="small"
                @click.stop="changeCampaignStatus(row.id, 'pause')"
              />
              <el-button
                v-if="['DRAFT', 'ACTIVE', 'PAUSED'].includes(row.status)"
                :icon="Warning"
                size="small"
                type="warning"
                @click.stop="changeCampaignStatus(row.id, 'cancel')"
              />
              <el-button :icon="Edit" size="small" @click.stop="editCampaign(row)" />
              <el-button :icon="Delete" size="small" @click.stop="deleteCampaign(row.id)" />
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">{{ campaignForm.id ? 'PUT /campaign/:id' : 'POST /campaign' }}</div>
            <h2>{{ campaignForm.id ? 'Редактировать кампанию' : 'Новая кампания' }}</h2>
          </div>
          <el-button v-if="campaignForm.id" :icon="Close" @click="resetCampaignForm" />
        </div>

        <el-form label-position="top">
          <el-form-item label="Label">
            <el-input v-model="campaignForm.label" />
          </el-form-item>
          <div class="form-grid">
            <el-form-item label="Письмо">
              <el-select v-model="campaignForm.message_id" filterable clearable class="full-width">
                <el-option
                  v-for="message in messageResult.items"
                  :key="message.id"
                  :label="message.label"
                  :value="message.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="Лендинг">
              <el-select v-model="campaignForm.landing_id" filterable clearable class="full-width">
                <el-option
                  v-for="landing in landingResult.items"
                  :key="landing.id"
                  :label="landing.label"
                  :value="landing.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="Обучение">
              <el-select v-model="campaignForm.education_id" filterable clearable class="full-width">
                <el-option
                  v-for="landing in landingResult.items"
                  :key="landing.id"
                  :label="landing.label"
                  :value="landing.id"
                />
              </el-select>
            </el-form-item>
          </div>
          <el-form-item label="Domain">
            <el-input v-model="campaignForm.domain" placeholder="http://localhost" />
          </el-form-item>
          <div class="form-grid">
            <el-form-item label="Date from">
              <el-date-picker
                v-model="campaignForm.date_from"
                type="datetime"
                value-format="YYYY-MM-DDTHH:mm:ssZ"
                class="full-width"
              />
            </el-form-item>
            <el-form-item label="Date to">
              <el-date-picker
                v-model="campaignForm.date_to"
                type="datetime"
                value-format="YYYY-MM-DDTHH:mm:ssZ"
                class="full-width"
              />
            </el-form-item>
          </div>
          <el-form-item label="Attributes JSON">
            <el-input v-model="campaignForm.attributes" type="textarea" :rows="5" />
          </el-form-item>
          <el-button
            type="primary"
            :icon="campaignForm.id ? Edit : Plus"
            :loading="loading.campaigns"
            @click="saveCampaign"
          >
            {{ campaignForm.id ? 'Сохранить кампанию' : 'Создать кампанию' }}
          </el-button>
        </el-form>
      </section>
    </div>

    <section class="panel">
      <div class="panel-heading">
        <div>
          <div class="mini-label">GET /vtarget | POST /target</div>
          <h2>Targets кампании</h2>
        </div>
        <div class="panel-actions">
          <el-button :icon="Refresh" :loading="loading.targets" @click="loadTargets" />
          <el-button :icon="MagicStick" :loading="loading.targets" @click="autoDistributeTargets">
            Распределить
          </el-button>
          <el-button :icon="Delete" :loading="loading.targets" @click="deleteCampaignTargets">
            Очистить
          </el-button>
        </div>
      </div>

      <div class="toolbar">
        <el-input v-model="vtargetQuery.full_name" placeholder="full_name" clearable />
        <el-select v-model="vtargetQuery.status" placeholder="target status" clearable>
          <el-option
            v-for="status in targetStatuses"
            :key="status"
            :label="status"
            :value="status"
          />
        </el-select>
        <el-button :icon="Select" type="primary" @click="loadTargets">Фильтр</el-button>
      </div>

      <div class="grid target-tools">
        <el-form label-position="top" class="inline-form">
          <el-form-item label="Сотрудник">
            <el-select v-model="targetForm.employee_id" filterable class="full-width">
              <el-option
                v-for="employee in employeeResult.items"
                :key="employee.id"
                :label="`${employee.first_name} ${employee.last_name} - ${employee.email}`"
                :value="employee.id"
              />
            </el-select>
          </el-form-item>
          <el-button type="primary" :icon="Plus" :loading="loading.targets" @click="createTarget">
            Добавить target
          </el-button>
        </el-form>

        <el-form label-position="top" class="inline-form">
          <el-form-item label="Scheduled at">
            <el-date-picker
              v-model="targetScheduleForm.scheduled_at"
              type="datetime"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
              class="full-width"
            />
          </el-form-item>
          <el-button :icon="Edit" :loading="loading.targets" @click="updateTargetSchedule">
            Сохранить schedule
          </el-button>
        </el-form>
      </div>

      <el-table
        class="data-table"
        :data="vtargetResult.items"
        highlight-current-row
        :row-class-name="({ row }) => row.id === selectedTargetId ? 'selected-row' : ''"
        @row-click="selectedTargetId = $event.id"
      >
        <el-table-column label="Target" min-width="190">
          <template #default="{ row }">
            <strong>{{ row.first_name }} {{ row.last_name }}</strong>
            <div class="muted">{{ shortId(row.id) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="Status" width="116">
          <template #default="{ row }">
            <el-tag :type="targetTag(row.status)" effect="plain">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Schedule" min-width="170">
          <template #default="{ row }">{{ formatDate(row.scheduled_at) }}</template>
        </el-table-column>
        <el-table-column label="Events" min-width="220">
          <template #default="{ row }">
            <div class="tag-line">
              <el-tag
                v-for="event in row.events.slice(0, 3)"
                :key="`${event.type}-${event.occurred_at}`"
                size="small"
                effect="plain"
              >
                {{ event.type }}
              </el-tag>
              <span v-if="!row.events.length" class="muted">нет событий</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Visit" min-width="160">
          <template #default="{ row }">
            <el-link v-if="visitLink(row)" :icon="Link" :href="visitLink(row)" target="_blank">
              открыть
            </el-link>
            <span v-else class="muted">нет ссылки</span>
          </template>
        </el-table-column>
        <el-table-column width="86" fixed="right">
          <template #default="{ row }">
            <el-button :icon="Delete" size="small" @click.stop="deleteTarget(row.id)" />
          </template>
        </el-table-column>
      </el-table>
    </section>

    <section class="panel">
      <div class="panel-heading">
        <div>
          <div class="mini-label">render as target</div>
          <h2>Демонстрация от лица сотрудника</h2>
        </div>
      </div>

      <div class="selection-summary three-up">
        <div>
          <span>Campaign</span>
          <strong>{{ selectedCampaign?.label || 'не выбрана' }}</strong>
        </div>
        <div>
          <span>Target</span>
          <strong>{{ targetName(selectedTarget) }}</strong>
        </div>
        <div>
          <span>Visit</span>
          <strong>
            <el-link v-if="visitLink(selectedTarget)" :href="visitLink(selectedTarget)" target="_blank">
              {{ visitLink(selectedTarget) }}
            </el-link>
            <span v-else>нет ссылки</span>
          </strong>
        </div>
      </div>

      <div class="template-actions">
        <el-button
          type="primary"
          :icon="MagicStick"
          :loading="loading.render"
          :disabled="!selectedCampaign?.message_id"
          @click="renderMessage(selectedCampaign?.message_id)"
        >
          Показать письмо
        </el-button>
        <el-button
          :icon="MagicStick"
          :loading="loading.render"
          :disabled="!selectedCampaign?.landing_id"
          @click="renderLanding(selectedCampaign?.landing_id)"
        >
          Показать сайт
        </el-button>
        <el-button
          :icon="MagicStick"
          :loading="loading.render"
          :disabled="!selectedCampaign?.education_id"
          @click="renderLanding(selectedCampaign?.education_id)"
        >
          Показать обучение
        </el-button>
      </div>

      <div class="grid two-columns">
        <div class="preview-shell">
          <iframe
            v-if="renderedContent"
            title="Campaign message preview"
            sandbox=""
            :srcdoc="renderedContent"
          />
          <el-empty v-else description="Письмо появится здесь" />
        </div>
        <div class="preview-shell">
          <iframe
            v-if="renderedLandingContent"
            title="Campaign landing preview"
            sandbox=""
            :srcdoc="renderedLandingContent"
          />
          <el-empty v-else description="Сайт появится здесь" />
        </div>
      </div>
    </section>
  </section>
</template>

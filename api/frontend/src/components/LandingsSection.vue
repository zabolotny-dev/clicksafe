<script setup>
import {
  Delete,
  DocumentChecked,
  Edit,
  MagicStick,
  Plus,
  Refresh,
  Select,
  Close,
  UploadFilled,
  View,
} from '@element-plus/icons-vue'

const selectedLandingId = defineModel('selectedLandingId', { type: String, required: true })
const selectedTargetId = defineModel('selectedTargetId', { type: String, required: true })
const landingTemplateText = defineModel('landingTemplateText', { type: String, required: true })

defineProps({
  deleteLanding: {
    type: Function,
    required: true,
  },
  editLanding: {
    type: Function,
    required: true,
  },
  landingForm: {
    type: Object,
    required: true,
  },
  landingQuery: {
    type: Object,
    required: true,
  },
  landingResult: {
    type: Object,
    required: true,
  },
  loading: {
    type: Object,
    required: true,
  },
  loadLandings: {
    type: Function,
    required: true,
  },
  onLandingFileChange: {
    type: Function,
    required: true,
  },
  onLandingFileRemove: {
    type: Function,
    required: true,
  },
  rawLandingContent: {
    type: String,
    default: '',
  },
  readLandingContent: {
    type: Function,
    required: true,
  },
  renderLanding: {
    type: Function,
    required: true,
  },
  renderedLandingContent: {
    type: String,
    default: '',
  },
  resetLandingForm: {
    type: Function,
    required: true,
  },
  saveLanding: {
    type: Function,
    required: true,
  },
  selectedLanding: {
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
  targetName: {
    type: Function,
    required: true,
  },
  uploadLandingFile: {
    type: Function,
    required: true,
  },
  uploadLandingText: {
    type: Function,
    required: true,
  },
  vtargetResult: {
    type: Object,
    required: true,
  },
})
</script>

<template>
  <section class="section-stack">
    <div class="grid two-columns">
      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">GET /landing</div>
            <h2>Лендинги</h2>
          </div>
          <el-button :icon="Refresh" :loading="loading.landings" @click="loadLandings" />
        </div>

        <div class="toolbar">
          <el-input v-model="landingQuery.label" placeholder="label" clearable />
          <el-button :icon="Select" type="primary" @click="loadLandings">Фильтр</el-button>
        </div>

        <el-table
          class="data-table"
          :data="landingResult.items"
          highlight-current-row
          :row-class-name="({ row }) => row.id === selectedLandingId ? 'selected-row' : ''"
          @row-click="editLanding"
        >
          <el-table-column prop="label" label="Label" min-width="220" />
          <el-table-column label="Content" width="110">
            <template #default="{ row }">
              <el-tag :type="row.has_content ? 'success' : 'info'" effect="plain">
                {{ row.has_content ? 'есть' : 'нет' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="Vars" min-width="220">
            <template #default="{ row }">
              <div class="tag-line">
                <el-tag
                  v-for="variable in row.required_vars"
                  :key="variable"
                  size="small"
                  effect="plain"
                >
                  {{ variable }}
                </el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column width="116" fixed="right">
            <template #default="{ row }">
              <el-button :icon="Edit" size="small" @click.stop="editLanding(row)" />
              <el-button :icon="Delete" size="small" @click.stop="deleteLanding(row.id)" />
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">{{ landingForm.id ? 'PUT /landing/:id' : 'POST /landing' }}</div>
            <h2>{{ landingForm.id ? 'Редактировать лендинг' : 'Новый лендинг' }}</h2>
          </div>
          <el-button v-if="landingForm.id" :icon="Close" @click="resetLandingForm" />
        </div>

        <el-form label-position="top">
          <el-form-item label="Label">
            <el-input v-model="landingForm.label" />
          </el-form-item>
          <el-button
            type="primary"
            :icon="landingForm.id ? Edit : Plus"
            :loading="loading.landings"
            @click="saveLanding"
          >
            {{ landingForm.id ? 'Сохранить лендинг' : 'Создать лендинг' }}
          </el-button>
        </el-form>
      </section>
    </div>

    <div class="grid two-columns wide-left">
      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">PUT /landing/:id/content</div>
            <h2>HTML лендинга</h2>
          </div>
          <el-tag effect="plain">{{ shortId(selectedLandingId) }}</el-tag>
        </div>

        <el-input
          v-model="landingTemplateText"
          type="textarea"
          :rows="15"
          resize="vertical"
          class="template-editor"
        />

        <div class="template-actions">
          <el-upload
            :auto-upload="false"
            :limit="1"
            accept=".html,text/html"
            :on-change="onLandingFileChange"
            :on-remove="onLandingFileRemove"
          >
            <el-button :icon="UploadFilled">HTML-файл</el-button>
          </el-upload>
          <el-button
            type="primary"
            :icon="UploadFilled"
            :loading="loading.landingContent"
            @click="uploadLandingText"
          >
            Загрузить текст
          </el-button>
          <el-button
            :icon="DocumentChecked"
            :loading="loading.landingContent"
            @click="uploadLandingFile"
          >
            Загрузить файл
          </el-button>
          <el-button
            :icon="View"
            :loading="loading.landingContent"
            @click="readLandingContent"
          >
            Прочитать
          </el-button>
        </div>

        <div v-if="selectedLanding?.required_vars?.length" class="vars-strip">
          <span>Required vars</span>
          <el-tag
            v-for="variable in selectedLanding.required_vars"
            :key="variable"
            effect="plain"
          >
            {{ variable }}
          </el-tag>
        </div>
      </section>

      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">POST /landing/:id/render</div>
            <h2>Предпросмотр</h2>
          </div>
        </div>

        <el-form label-position="top">
          <el-form-item label="Target">
            <el-select
              v-model="selectedTargetId"
              filterable
              placeholder="Выбери target из кампании"
              class="full-width"
            >
              <el-option
                v-for="target in vtargetResult.items"
                :key="target.id"
                :label="`${target.first_name} ${target.last_name} - ${target.status}`"
                :value="target.id"
              />
            </el-select>
          </el-form-item>
          <el-button
            type="primary"
            :icon="MagicStick"
            :loading="loading.render"
            @click="renderLanding()"
          >
            Показать лендинг
          </el-button>
        </el-form>

        <div class="selection-summary">
          <div>
            <span>Landing</span>
            <strong>{{ selectedLanding?.label || 'не выбран' }}</strong>
          </div>
          <div>
            <span>Target</span>
            <strong>{{ targetName(selectedTarget) }}</strong>
          </div>
        </div>

        <div class="preview-shell">
          <iframe
            v-if="renderedLandingContent"
            title="Rendered landing"
            sandbox=""
            :srcdoc="renderedLandingContent"
          />
          <el-empty v-else description="Результат появится здесь" />
        </div>
      </section>
    </div>

    <section class="panel">
      <el-tabs>
        <el-tab-pane label="Rendered HTML">
          <pre class="code-block">{{ renderedLandingContent || 'Пока пусто' }}</pre>
        </el-tab-pane>
        <el-tab-pane label="Source HTML">
          <pre class="code-block">{{ rawLandingContent || landingTemplateText }}</pre>
        </el-tab-pane>
      </el-tabs>
    </section>
  </section>
</template>

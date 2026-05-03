<script setup>
import {
  Delete,
  DocumentChecked,
  MagicStick,
  Plus,
  Refresh,
  Select,
  UploadFilled,
  View,
} from '@element-plus/icons-vue'

const selectedMessageId = defineModel('selectedMessageId', { type: String, required: true })
const selectedEmployeeId = defineModel('selectedEmployeeId', { type: String, required: true })
const templateText = defineModel('templateText', { type: String, required: true })

defineProps({
  createMessage: {
    type: Function,
    required: true,
  },
  deleteMessage: {
    type: Function,
    required: true,
  },
  employeeResult: {
    type: Object,
    required: true,
  },
  loading: {
    type: Object,
    required: true,
  },
  loadMessages: {
    type: Function,
    required: true,
  },
  messageForm: {
    type: Object,
    required: true,
  },
  messageQuery: {
    type: Object,
    required: true,
  },
  messageResult: {
    type: Object,
    required: true,
  },
  onTemplateFileChange: {
    type: Function,
    required: true,
  },
  onTemplateFileRemove: {
    type: Function,
    required: true,
  },
  rawTemplateContent: {
    type: String,
    default: '',
  },
  readTemplateContent: {
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
  selectedEmployee: {
    type: Object,
    default: null,
  },
  selectedMessage: {
    type: Object,
    default: null,
  },
  shortId: {
    type: Function,
    required: true,
  },
  uploadTemplateFile: {
    type: Function,
    required: true,
  },
  uploadTemplateText: {
    type: Function,
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
            <div class="mini-label">GET /message</div>
            <h2>Шаблоны</h2>
          </div>
          <el-button :icon="Refresh" :loading="loading.messages" @click="loadMessages" />
        </div>

        <div class="toolbar">
          <el-input v-model="messageQuery.label" placeholder="label" clearable />
          <el-input v-model="messageQuery.subject" placeholder="subject" clearable />
          <el-button :icon="Select" type="primary" @click="loadMessages">Фильтр</el-button>
        </div>

        <el-table
          class="data-table"
          :data="messageResult.items"
          highlight-current-row
          :row-class-name="({ row }) => row.id === selectedMessageId ? 'selected-row' : ''"
          @row-click="selectedMessageId = $event.id"
        >
          <el-table-column prop="label" label="Label" min-width="180" />
          <el-table-column prop="subject" label="Subject" min-width="180" />
          <el-table-column label="Content" width="110">
            <template #default="{ row }">
              <el-tag :type="row.has_content ? 'success' : 'info'" effect="plain">
                {{ row.has_content ? 'есть' : 'нет' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="Vars" min-width="180">
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
          <el-table-column width="148" fixed="right">
            <template #default="{ row }">
              <el-button
                size="small"
                :type="row.id === selectedMessageId ? 'primary' : 'default'"
                @click.stop="selectedMessageId = row.id"
              >
                Выбрать
              </el-button>
              <el-button
                size="small"
                :icon="Delete"
                @click.stop="deleteMessage(row.id)"
              />
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">POST /message</div>
            <h2>Новый message</h2>
          </div>
        </div>

        <el-form label-position="top">
          <el-form-item label="Label">
            <el-input v-model="messageForm.label" />
          </el-form-item>
          <div class="form-grid">
            <el-form-item label="From email">
              <el-input v-model="messageForm.from_email" />
            </el-form-item>
            <el-form-item label="From name">
              <el-input v-model="messageForm.from_name" />
            </el-form-item>
          </div>
          <el-form-item label="Subject">
            <el-input v-model="messageForm.subject" />
          </el-form-item>
          <el-button
            type="primary"
            :icon="Plus"
            :loading="loading.messages"
            @click="createMessage"
          >
            Создать
          </el-button>
        </el-form>
      </section>
    </div>

    <div class="grid two-columns wide-left">
      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">PUT /message/:id/content</div>
            <h2>HTML-шаблон</h2>
          </div>
          <el-tag effect="plain">{{ shortId(selectedMessageId) }}</el-tag>
        </div>

        <el-input
          v-model="templateText"
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
            :on-change="onTemplateFileChange"
            :on-remove="onTemplateFileRemove"
          >
            <el-button :icon="UploadFilled">HTML-файл</el-button>
          </el-upload>
          <el-button
            type="primary"
            :icon="UploadFilled"
            :loading="loading.content"
            @click="uploadTemplateText"
          >
            Загрузить текст
          </el-button>
          <el-button
            :icon="DocumentChecked"
            :loading="loading.content"
            @click="uploadTemplateFile"
          >
            Загрузить файл
          </el-button>
          <el-button
            :icon="View"
            :loading="loading.content"
            @click="readTemplateContent"
          >
            Прочитать
          </el-button>
        </div>

        <div v-if="selectedMessage?.required_vars?.length" class="vars-strip">
          <span>Required vars</span>
          <el-tag
            v-for="variable in selectedMessage.required_vars"
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
            <div class="mini-label">POST /message/:id/render</div>
            <h2>Рендер</h2>
          </div>
        </div>

        <el-form label-position="top">
          <el-form-item label="Employee">
            <el-select
              v-model="selectedEmployeeId"
              filterable
              placeholder="Выбери сотрудника"
              class="full-width"
            >
              <el-option
                v-for="employee in employeeResult.items"
                :key="employee.id"
                :label="`${employee.first_name} ${employee.last_name} - ${employee.email}`"
                :value="employee.id"
              />
            </el-select>
          </el-form-item>
          <el-button
            type="primary"
            :icon="MagicStick"
            :loading="loading.render"
            @click="renderMessage"
          >
            Получить результат
          </el-button>
        </el-form>

        <div class="selection-summary">
          <div>
            <span>Message</span>
            <strong>{{ selectedMessage?.label || 'не выбран' }}</strong>
          </div>
          <div>
            <span>Employee</span>
            <strong>
              {{ selectedEmployee ? `${selectedEmployee.first_name} ${selectedEmployee.last_name}` : 'не выбран' }}
            </strong>
          </div>
        </div>

        <div class="preview-shell">
          <iframe
            v-if="renderedContent"
            title="Rendered message"
            sandbox=""
            :srcdoc="renderedContent"
          />
          <el-empty v-else description="Результат появится здесь" />
        </div>
      </section>
    </div>

    <section class="panel">
      <el-tabs>
        <el-tab-pane label="Rendered HTML">
          <pre class="code-block">{{ renderedContent || 'Пока пусто' }}</pre>
        </el-tab-pane>
        <el-tab-pane label="Source HTML">
          <pre class="code-block">{{ rawTemplateContent || templateText }}</pre>
        </el-tab-pane>
      </el-tabs>
    </section>
  </section>
</template>

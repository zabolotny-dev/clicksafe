<script setup>
import { Promotion } from '@element-plus/icons-vue'

defineProps({
  eventForm: {
    type: Object,
    required: true,
  },
  loading: {
    type: Object,
    required: true,
  },
  publishEvent: {
    type: Function,
    required: true,
  },
  selectedEmployeeId: {
    type: String,
    default: '',
  },
})
</script>

<template>
  <section class="section-stack">
    <section class="panel compact-panel">
      <div class="panel-heading">
        <div>
          <div class="mini-label">POST /events</div>
          <h2>Событие</h2>
        </div>
      </div>
      <el-form label-position="top">
        <el-form-item label="Campaign ID">
          <el-input v-model="eventForm.campaign_id" placeholder="uuid" />
        </el-form-item>
        <el-form-item label="Employee ID">
          <el-input v-model="eventForm.employee_id" :placeholder="selectedEmployeeId || 'uuid'" />
        </el-form-item>
        <el-form-item label="Type">
          <el-select v-model="eventForm.type" class="full-width">
            <el-option label="MESSAGE_SENT" value="MESSAGE_SENT" />
            <el-option label="DELIVERY_FAILED" value="DELIVERY_FAILED" />
            <el-option label="EMAIL_OPENED" value="EMAIL_OPENED" />
            <el-option label="LINK_OPENED" value="LINK_OPENED" />
            <el-option label="DATA_SENT" value="DATA_SENT" />
          </el-select>
        </el-form-item>
        <el-button
          type="primary"
          :icon="Promotion"
          :loading="loading.event"
          @click="publishEvent"
        >
          Опубликовать
        </el-button>
      </el-form>
    </section>
  </section>
</template>

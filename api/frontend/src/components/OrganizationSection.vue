<script setup>
import { Check, Refresh } from '@element-plus/icons-vue'

defineProps({
  loadOrganization: {
    type: Function,
    required: true,
  },
  loading: {
    type: Object,
    required: true,
  },
  organization: {
    type: Object,
    default: null,
  },
  organizationForm: {
    type: Object,
    required: true,
  },
  pretty: {
    type: Function,
    required: true,
  },
  saveOrganization: {
    type: Function,
    required: true,
  },
})
</script>

<template>
  <section class="section-stack">
    <section class="panel">
      <div class="panel-heading">
        <div>
          <div class="mini-label">GET /organization | POST /organization | PUT /organization</div>
          <h2>Организация</h2>
        </div>
        <el-button :icon="Refresh" :loading="loading.organization" @click="loadOrganization" />
      </div>

      <div class="grid two-columns">
        <el-form label-position="top">
          <el-form-item label="Label">
            <el-input v-model="organizationForm.label" />
          </el-form-item>
          <el-form-item label="Attributes JSON">
            <el-input v-model="organizationForm.attributes" type="textarea" :rows="7" />
          </el-form-item>
          <el-button
            type="primary"
            :icon="Check"
            :loading="loading.organization"
            @click="saveOrganization"
          >
            Сохранить
          </el-button>
        </el-form>

        <div class="object-view">
          <div class="mini-label">current</div>
          <pre class="code-block">{{ organization ? pretty(organization) : 'Организация еще не создана' }}</pre>
        </div>
      </div>
    </section>
  </section>
</template>

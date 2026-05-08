<script setup>
import { Check, Delete, Refresh, UploadFilled } from '@element-plus/icons-vue'

defineProps({
  deleteOrganizationLogo: {
    type: Function,
    required: true,
  },
  loadOrganization: {
    type: Function,
    required: true,
  },
  loading: {
    type: Object,
    required: true,
  },
  onLogoFileChange: {
    type: Function,
    required: true,
  },
  onLogoFileRemove: {
    type: Function,
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
  uploadOrganizationLogo: {
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

    <section class="panel compact-panel">
      <div class="panel-heading">
        <div>
          <div class="mini-label">PUT /organization/logo | DELETE /organization/logo</div>
          <h2>Логотип</h2>
        </div>
      </div>

      <div class="template-actions">
        <el-upload
          :auto-upload="false"
          :limit="1"
          accept="image/*"
          :on-change="onLogoFileChange"
          :on-remove="onLogoFileRemove"
        >
          <el-button :icon="UploadFilled">Изображение</el-button>
        </el-upload>
        <el-button
          type="primary"
          :icon="UploadFilled"
          :loading="loading.logo"
          @click="uploadOrganizationLogo"
        >
          Загрузить
        </el-button>
        <el-button
          :icon="Delete"
          :loading="loading.logo"
          @click="deleteOrganizationLogo"
        >
          Удалить
        </el-button>
      </div>

      <el-link v-if="organization?.logo_path" :href="organization.logo_path" target="_blank">
        {{ organization.logo_path }}
      </el-link>
      <el-empty v-else description="Логотип не загружен" />
    </section>
  </section>
</template>

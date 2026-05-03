<script setup>
defineProps({
  endpointGroups: {
    type: Array,
    required: true,
  },
})
</script>

<template>
  <section class="section-stack">
    <section class="panel">
      <div class="panel-heading">
        <div>
          <div class="mini-label">routes</div>
          <h2>Маршруты app слоя</h2>
        </div>
      </div>

      <div class="route-groups">
        <div
          v-for="group in endpointGroups"
          :key="group.name"
          class="route-group"
        >
          <div class="route-title">
            <el-icon><component :is="group.icon" /></el-icon>
            <strong>{{ group.name }}</strong>
          </div>
          <el-table :data="group.routes" class="route-table">
            <el-table-column width="92">
              <template #default="{ row }">
                <el-tag effect="plain" :type="row[0] === 'GET' ? 'success' : row[0] === 'DELETE' ? 'danger' : 'warning'">
                  {{ row[0] }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Path" width="260">
              <template #default="{ row }">
                <code>{{ row[1] }}</code>
              </template>
            </el-table-column>
            <el-table-column label="Назначение">
              <template #default="{ row }">{{ row[2] }}</template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </section>
  </section>
</template>

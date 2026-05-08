<script setup>
import { Close, Delete, Edit, Plus, Refresh, Select } from '@element-plus/icons-vue'

const selectedEmployeeId = defineModel('selectedEmployeeId', { type: String, required: true })

defineProps({
  deleteDepartment: {
    type: Function,
    required: true,
  },
  deleteEmployee: {
    type: Function,
    required: true,
  },
  departmentById: {
    type: Function,
    required: true,
  },
  departmentForm: {
    type: Object,
    required: true,
  },
  departmentQuery: {
    type: Object,
    required: true,
  },
  departmentResult: {
    type: Object,
    required: true,
  },
  editDepartment: {
    type: Function,
    required: true,
  },
  editEmployee: {
    type: Function,
    required: true,
  },
  employeeForm: {
    type: Object,
    required: true,
  },
  employeeQuery: {
    type: Object,
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
  loadDepartments: {
    type: Function,
    required: true,
  },
  loadEmployees: {
    type: Function,
    required: true,
  },
  resetDepartmentForm: {
    type: Function,
    required: true,
  },
  resetEmployeeForm: {
    type: Function,
    required: true,
  },
  saveDepartment: {
    type: Function,
    required: true,
  },
  saveEmployee: {
    type: Function,
    required: true,
  },
  shortId: {
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
            <div class="mini-label">{{ departmentForm.id ? 'PUT /department/:id' : 'POST /department' }}</div>
            <h2>{{ departmentForm.id ? 'Редактировать отдел' : 'Новый отдел' }}</h2>
          </div>
          <el-button v-if="departmentForm.id" :icon="Close" @click="resetDepartmentForm" />
        </div>
        <el-form label-position="top">
          <el-form-item label="Label">
            <el-input v-model="departmentForm.label" />
          </el-form-item>
          <el-form-item label="Attributes JSON">
            <el-input v-model="departmentForm.attributes" type="textarea" :rows="5" />
          </el-form-item>
          <el-button
            type="primary"
            :icon="departmentForm.id ? Edit : Plus"
            :loading="loading.departments"
            @click="saveDepartment"
          >
            {{ departmentForm.id ? 'Сохранить отдел' : 'Создать отдел' }}
          </el-button>
        </el-form>
      </section>

      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">{{ employeeForm.id ? 'PUT /employee/:id' : 'POST /employee' }}</div>
            <h2>{{ employeeForm.id ? 'Редактировать сотрудника' : 'Новый сотрудник' }}</h2>
          </div>
          <el-button v-if="employeeForm.id" :icon="Close" @click="resetEmployeeForm" />
        </div>
        <el-form label-position="top">
          <el-form-item label="Department">
            <el-select v-model="employeeForm.department_id" clearable filterable class="full-width">
              <el-option
                v-for="department in departmentResult.items"
                :key="department.id"
                :label="department.label"
                :value="department.id"
              />
            </el-select>
          </el-form-item>
          <div class="form-grid">
            <el-form-item label="First name">
              <el-input v-model="employeeForm.first_name" />
            </el-form-item>
            <el-form-item label="Last name">
              <el-input v-model="employeeForm.last_name" />
            </el-form-item>
          </div>
          <div class="form-grid">
            <el-form-item label="Email">
              <el-input v-model="employeeForm.email" />
            </el-form-item>
            <el-form-item label="Phone">
              <el-input v-model="employeeForm.phone" />
            </el-form-item>
          </div>
          <el-form-item label="Attributes JSON">
            <el-input v-model="employeeForm.attributes" type="textarea" :rows="5" />
          </el-form-item>
          <el-button
            type="primary"
            :icon="employeeForm.id ? Edit : Plus"
            :loading="loading.employees"
            @click="saveEmployee"
          >
            {{ employeeForm.id ? 'Сохранить сотрудника' : 'Создать сотрудника' }}
          </el-button>
        </el-form>
      </section>
    </div>

    <div class="grid two-columns">
      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">GET /department</div>
            <h2>Отделы</h2>
          </div>
          <el-button :icon="Refresh" :loading="loading.departments" @click="loadDepartments" />
        </div>
        <div class="toolbar">
          <el-input v-model="departmentQuery.label" placeholder="label" clearable />
          <el-button :icon="Select" type="primary" @click="loadDepartments">Фильтр</el-button>
        </div>
        <el-table :data="departmentResult.items" class="data-table" @row-click="editDepartment">
          <el-table-column prop="label" label="Label" min-width="170" />
          <el-table-column label="ID" width="150">
            <template #default="{ row }">{{ shortId(row.id) }}</template>
          </el-table-column>
          <el-table-column width="116" fixed="right">
            <template #default="{ row }">
              <el-button :icon="Edit" size="small" @click.stop="editDepartment(row)" />
              <el-button :icon="Delete" size="small" @click.stop="deleteDepartment(row.id)" />
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="panel">
        <div class="panel-heading">
          <div>
            <div class="mini-label">GET /employee</div>
            <h2>Сотрудники</h2>
          </div>
          <el-button :icon="Refresh" :loading="loading.employees" @click="loadEmployees" />
        </div>
        <div class="toolbar">
          <el-input v-model="employeeQuery.full_name" placeholder="full_name" clearable />
          <el-input v-model="employeeQuery.email" placeholder="email" clearable />
          <el-select v-model="employeeQuery.department_id" placeholder="department" clearable filterable>
            <el-option
              v-for="department in departmentResult.items"
              :key="department.id"
              :label="department.label"
              :value="department.id"
            />
          </el-select>
          <el-button :icon="Select" type="primary" @click="loadEmployees">Фильтр</el-button>
        </div>
        <el-table
          :data="employeeResult.items"
          class="data-table"
          highlight-current-row
          :row-class-name="({ row }) => row.id === selectedEmployeeId ? 'selected-row' : ''"
          @row-click="editEmployee"
        >
          <el-table-column label="Name" min-width="180">
            <template #default="{ row }">{{ row.first_name }} {{ row.last_name }}</template>
          </el-table-column>
          <el-table-column prop="email" label="Email" min-width="210" />
          <el-table-column label="Department" min-width="150">
            <template #default="{ row }">{{ departmentById(row.department_id)?.label || 'без отдела' }}</template>
          </el-table-column>
          <el-table-column label="Selected" width="96">
            <template #default="{ row }">
              <el-tag v-if="row.id === selectedEmployeeId" type="success" effect="plain">да</el-tag>
            </template>
          </el-table-column>
          <el-table-column width="116" fixed="right">
            <template #default="{ row }">
              <el-button :icon="Edit" size="small" @click.stop="editEmployee(row)" />
              <el-button :icon="Delete" size="small" @click.stop="deleteEmployee(row.id)" />
            </template>
          </el-table-column>
        </el-table>
      </section>
    </div>
  </section>
</template>

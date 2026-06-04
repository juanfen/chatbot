<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormRules } from 'naive-ui';
import { NButton, NPopconfirm } from 'naive-ui';
import { useBoolean } from '@sa/hooks';
import {
  fetchSysRoleCreate,
  fetchSysRoleDelete,
  fetchSysRoleList,
  fetchSysRoleSetRoutes,
  fetchSysRoleUpdate,
  fetchSysRouteList
} from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'SysRole' });

const loading = ref(false);
const list = ref<Api.Sys.Role[]>([]);

const routeRows = ref<Api.Sys.RouteRow[]>([]);
const routeCheckboxOptions = computed(() =>
  routeRows.value.map(r => ({
    label: `${r.title || r.name} (${r.path})`,
    value: r.id
  }))
);

const { bool: showModal, setBool: setShowModal } = useBoolean();
const { bool: showRoutesModal, setBool: setShowRoutesModal } = useBoolean();
const isEdit = ref(false);
const routeRole = ref<Api.Sys.Role | null>(null);
const routeIdsDraft = ref<string[]>([]);

const formRef = ref();
const formModel = reactive({
  id: '',
  code: '',
  name: '',
  status: 1
});

const rules: FormRules = {
  code: { required: true, message: () => $t('page.sysRole.rules.code'), trigger: 'blur' },
  name: { required: true, message: () => $t('page.sysRole.rules.name'), trigger: 'blur' }
};

async function loadRoutes() {
  const data = await fetchSysRouteList();
  routeRows.value = data.list ?? [];
}

async function loadList() {
  loading.value = true;
  try {
    const data = await fetchSysRoleList();
    list.value = data.list ?? [];
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  isEdit.value = false;
  Object.assign(formModel, { id: '', code: '', name: '', status: 1 });
  setShowModal(true);
}

function openEdit(row: Api.Sys.Role) {
  isEdit.value = true;
  Object.assign(formModel, { id: row.id, code: row.code, name: row.name, status: row.status });
  setShowModal(true);
}

function openRoutes(row: Api.Sys.Role) {
  routeRole.value = row;
  routeIdsDraft.value = [...(row.routeIds || [])];
  setShowRoutesModal(true);
}

async function saveRoutes() {
  if (!routeRole.value) return;
  await fetchSysRoleSetRoutes({ roleId: routeRole.value.id, routeIds: routeIdsDraft.value });
  window.$message?.success($t('common.updateSuccess'));
  setShowRoutesModal(false);
  await loadList();
}

async function submitForm() {
  await formRef.value?.validate();
  if (isEdit.value) {
    await fetchSysRoleUpdate({ id: formModel.id, name: formModel.name, status: formModel.status });
    window.$message?.success($t('common.updateSuccess'));
  } else {
    await fetchSysRoleCreate({ code: formModel.code, name: formModel.name });
    window.$message?.success($t('common.addSuccess'));
  }
  setShowModal(false);
  await loadList();
}

async function handleDelete(row: Api.Sys.Role) {
  await fetchSysRoleDelete(row.id);
  window.$message?.success($t('common.deleteSuccess'));
  await loadList();
}

const columns = computed<DataTableColumns<Api.Sys.Role>>(() => [
  { title: () => $t('page.sysRole.code'), key: 'code', width: 140 },
  { title: () => $t('page.sysRole.name'), key: 'name', width: 160 },
  {
    title: () => $t('page.sysRole.status'),
    key: 'status',
    width: 90,
    render: row => (row.status === 1 ? $t('common.yesOrNo.yes') : $t('common.yesOrNo.no'))
  },
  {
    title: () => $t('page.sysRole.routeCount'),
    key: 'routeIds',
    width: 100,
    render: row => String(row.routeIds?.length ?? 0)
  },
  {
    title: () => $t('common.action'),
    key: 'actions',
    width: 280,
    fixed: 'right',
    render: row =>
      h('div', { class: 'flex flex-wrap gap-8px' }, [
        h(
          NButton,
          { size: 'small', type: 'primary', ghost: true, onClick: () => openEdit(row) },
          { default: () => $t('common.edit') }
        ),
        h(
          NButton,
          { size: 'small', type: 'info', ghost: true, onClick: () => openRoutes(row) },
          { default: () => $t('page.sysRole.routes') }
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDelete(row) },
          {
            default: () => $t('common.confirmDelete'),
            trigger: () =>
              h(
                NButton,
                { size: 'small', type: 'error', ghost: true, disabled: row.code === 'admin' },
                { default: () => $t('common.delete') }
              )
          }
        )
      ])
  }
]);

onMounted(async () => {
  await loadRoutes();
  await loadList();
});

const cardContentStyle = {
  display: 'flex',
  flexDirection: 'column' as const,
  height: '100%',
  minHeight: 0,
  padding: 0
};
</script>

<template>
  <div class="sys-mgmt-page box-border min-h-0 flex flex-1 flex-col">
    <NCard :bordered="false" class="card-wrapper min-h-0 flex flex-1 flex-col overflow-hidden" :content-style="cardContentStyle">
      <div class="shrink-0 px-16px pb-12px pt-16px">
        <NSpace justify="end">
          <NButton type="primary" @click="openCreate">{{ $t('common.add') }}</NButton>
        </NSpace>
      </div>
      <div class="table-fill min-h-0 flex flex-1 flex-col px-16px pb-16px">
        <NDataTable :columns="columns" :data="list" :loading="loading" :scroll-x="900" flex-height class="h-full" />
      </div>
    </NCard>

    <NModal v-model:show="showModal" preset="card" :title="isEdit ? $t('common.edit') : $t('common.add')" class="w-480px">
      <NForm ref="formRef" :model="formModel" :rules="rules" label-placement="left" label-width="88">
        <NFormItem path="code" :label="$t('page.sysRole.code')">
          <NInput v-model:value="formModel.code" :disabled="isEdit" />
        </NFormItem>
        <NFormItem path="name" :label="$t('page.sysRole.name')">
          <NInput v-model:value="formModel.name" />
        </NFormItem>
        <NFormItem v-if="isEdit" path="status" :label="$t('page.sysRole.status')">
          <NRadioGroup v-model:value="formModel.status">
            <NSpace>
              <NRadio :value="1">{{ $t('common.yesOrNo.yes') }}</NRadio>
              <NRadio :value="0">{{ $t('common.yesOrNo.no') }}</NRadio>
            </NSpace>
          </NRadioGroup>
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="setShowModal(false)">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="submitForm">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>

    <NModal
      v-model:show="showRoutesModal"
      preset="card"
      :title="$t('page.sysRole.routesTitle')"
      class="w-640px"
      @after-leave="routeRole = null"
    >
      <NScrollbar class="max-h-420px">
        <NCheckboxGroup v-model:value="routeIdsDraft">
          <NSpace vertical :size="8">
            <NCheckbox v-for="opt in routeCheckboxOptions" :key="opt.value" :value="opt.value" :label="opt.label" />
          </NSpace>
        </NCheckboxGroup>
      </NScrollbar>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="setShowRoutesModal(false)">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="saveRoutes">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.sys-mgmt-page {
  min-height: 0;
}
.table-fill :deep(.n-data-table) {
  flex: 1;
  min-height: 0;
}
</style>

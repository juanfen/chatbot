<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormRules } from 'naive-ui';
import { NButton, NPopconfirm } from 'naive-ui';
import { useBoolean } from '@sa/hooks';
import {
  fetchSysPackageCreate,
  fetchSysPackageDelete,
  fetchSysPackageList,
  fetchSysPackageUpdate
} from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'SysPackage' });

const loading = ref(false);
const list = ref<Api.Biz.SysPackage[]>([]);
const { bool: showModal, setBool: setShowModal } = useBoolean();
const isEdit = ref(false);
const formRef = ref();
const formModel = reactive({
  id: '',
  code: '',
  name: '',
  description: '',
  quotaCalls: 0,
  sortNo: 0,
  status: 1
});

const rules: FormRules = {
  code: { required: true, message: () => $t('page.sysPackage.rules.code'), trigger: 'blur' },
  name: { required: true, message: () => $t('page.sysPackage.rules.name'), trigger: 'blur' }
};

async function loadList() {
  loading.value = true;
  try {
    list.value = await fetchSysPackageList();
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  isEdit.value = false;
  Object.assign(formModel, { id: '', code: '', name: '', description: '', quotaCalls: 100, sortNo: 50, status: 1 });
  setShowModal(true);
}

function openEdit(row: Api.Biz.SysPackage) {
  isEdit.value = true;
  Object.assign(formModel, {
    id: row.id,
    code: row.code,
    name: row.name,
    description: row.description,
    quotaCalls: row.quotaCalls,
    sortNo: row.sortNo,
    status: row.status
  });
  setShowModal(true);
}

async function submitForm() {
  await formRef.value?.validate();
  if (isEdit.value) {
    await fetchSysPackageUpdate({
      id: formModel.id,
      code: formModel.code,
      name: formModel.name,
      description: formModel.description,
      quotaCalls: formModel.quotaCalls,
      sortNo: formModel.sortNo,
      status: formModel.status
    });
    window.$message?.success($t('common.updateSuccess'));
  } else {
    await fetchSysPackageCreate({
      code: formModel.code,
      name: formModel.name,
      description: formModel.description,
      quotaCalls: formModel.quotaCalls,
      sortNo: formModel.sortNo,
      status: formModel.status
    });
    window.$message?.success($t('common.addSuccess'));
  }
  setShowModal(false);
  await loadList();
}

async function handleDelete(row: Api.Biz.SysPackage) {
  await fetchSysPackageDelete(row.id);
  window.$message?.success($t('common.deleteSuccess'));
  await loadList();
}

const columns = computed<DataTableColumns<Api.Biz.SysPackage>>(() => [
  { title: () => $t('page.sysPackage.code'), key: 'code', width: 120 },
  { title: () => $t('page.sysPackage.name'), key: 'name', width: 140 },
  {
    title: () => $t('page.sysPackage.description'),
    key: 'description',
    minWidth: 180,
    ellipsis: { tooltip: true }
  },
  { title: () => $t('page.sysPackage.quotaCalls'), key: 'quotaCalls', width: 100 },
  { title: () => $t('page.sysPackage.sortNo'), key: 'sortNo', width: 72 },
  {
    title: () => $t('page.sysPackage.status'),
    key: 'status',
    width: 80,
    render: row => (row.status === 1 ? $t('common.yesOrNo.yes') : $t('common.yesOrNo.no'))
  },
  {
    title: () => $t('common.action'),
    key: 'actions',
    width: 160,
    fixed: 'right',
    render: row =>
      h('div', { class: 'flex gap-8px' }, [
        h(
          NButton,
          { size: 'small', type: 'primary', ghost: true, onClick: () => openEdit(row) },
          { default: () => $t('common.edit') }
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDelete(row) },
          {
            default: () => $t('common.confirmDelete'),
            trigger: () =>
              h(NButton, { size: 'small', type: 'error', ghost: true }, { default: () => $t('common.delete') })
          }
        )
      ])
  }
]);

const cardContentStyle = {
  display: 'flex',
  flexDirection: 'column' as const,
  height: '100%',
  minHeight: 0,
  padding: 0
};

onMounted(loadList);
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
        <NDataTable :columns="columns" :data="list" :loading="loading" :scroll-x="960" flex-height class="h-full" />
      </div>
    </NCard>

    <NModal v-model:show="showModal" preset="card" :title="isEdit ? $t('common.edit') : $t('common.add')" class="w-520px">
      <NForm ref="formRef" :model="formModel" :rules="rules" label-placement="left" label-width="96">
        <NFormItem path="code" :label="$t('page.sysPackage.code')">
          <NInput v-model:value="formModel.code" :disabled="isEdit" />
        </NFormItem>
        <NFormItem path="name" :label="$t('page.sysPackage.name')">
          <NInput v-model:value="formModel.name" />
        </NFormItem>
        <NFormItem path="description" :label="$t('page.sysPackage.description')">
          <NInput v-model:value="formModel.description" type="textarea" :rows="3" />
        </NFormItem>
        <NFormItem path="quotaCalls" :label="$t('page.sysPackage.quotaCalls')">
          <NInputNumber v-model:value="formModel.quotaCalls" class="w-full" :min="0" />
        </NFormItem>
        <NFormItem path="sortNo" :label="$t('page.sysPackage.sortNo')">
          <NInputNumber v-model:value="formModel.sortNo" class="w-full" />
        </NFormItem>
        <NFormItem v-if="isEdit" path="status" :label="$t('page.sysPackage.status')">
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

<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormRules } from 'naive-ui';
import { NButton, NPopconfirm } from 'naive-ui';
import { useBoolean } from '@sa/hooks';
import { fetchSysRouteCreate, fetchSysRouteDelete, fetchSysRouteList, fetchSysRouteUpdate } from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'SysRoute' });

const loading = ref(false);
const list = ref<Api.Sys.RouteRow[]>([]);

const parentOptions = computed(() =>
  [{ label: $t('page.sysRoute.root'), value: '' }].concat(
    list.value.map(r => ({ label: `${r.title || r.name} (${r.name})`, value: r.id }))
  )
);

const { bool: showModal, setBool: setShowModal } = useBoolean();
const isEdit = ref(false);
const formRef = ref();
const formModel = reactive({
  id: '',
  parentId: '',
  name: '',
  path: '',
  component: '',
  title: '',
  i18nKey: '',
  icon: '',
  orderNo: 0,
  hideInMenu: 0,
  keepAlive: 0,
  status: 1
});

const rules: FormRules = {
  name: { required: true, message: () => $t('page.sysRoute.rules.name'), trigger: 'blur' },
  path: { required: true, message: () => $t('page.sysRoute.rules.path'), trigger: 'blur' },
  component: { required: true, message: () => $t('page.sysRoute.rules.component'), trigger: 'blur' }
};

async function loadList() {
  loading.value = true;
  try {
    const data = await fetchSysRouteList();
    list.value = data.list ?? [];
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  isEdit.value = false;
  Object.assign(formModel, {
    id: '',
    parentId: '',
    name: '',
    path: '',
    component: 'layout.base$view.',
    title: '',
    i18nKey: '',
    icon: '',
    orderNo: 50,
    hideInMenu: 0,
    keepAlive: 0,
    status: 1
  });
  setShowModal(true);
}

function openEdit(row: Api.Sys.RouteRow) {
  isEdit.value = true;
  Object.assign(formModel, {
    id: row.id,
    parentId: row.parentId,
    name: row.name,
    path: row.path,
    component: row.component,
    title: row.title,
    i18nKey: row.i18nKey,
    icon: row.icon,
    orderNo: row.orderNo,
    hideInMenu: row.hideInMenu,
    keepAlive: row.keepAlive,
    status: row.status
  });
  setShowModal(true);
}

async function submitForm() {
  await formRef.value?.validate();
  if (isEdit.value) {
    await fetchSysRouteUpdate({
      id: formModel.id,
      parentId: formModel.parentId,
      path: formModel.path,
      component: formModel.component,
      title: formModel.title,
      i18nKey: formModel.i18nKey,
      icon: formModel.icon,
      orderNo: formModel.orderNo,
      hideInMenu: formModel.hideInMenu,
      keepAlive: formModel.keepAlive,
      status: formModel.status
    });
    window.$message?.success($t('common.updateSuccess'));
  } else {
    await fetchSysRouteCreate({
      parentId: formModel.parentId || undefined,
      name: formModel.name,
      path: formModel.path,
      component: formModel.component,
      title: formModel.title,
      i18nKey: formModel.i18nKey,
      icon: formModel.icon,
      orderNo: formModel.orderNo,
      hideInMenu: formModel.hideInMenu,
      keepAlive: formModel.keepAlive,
      status: formModel.status
    });
    window.$message?.success($t('common.addSuccess'));
  }
  setShowModal(false);
  await loadList();
}

async function handleDelete(row: Api.Sys.RouteRow) {
  await fetchSysRouteDelete(row.id);
  window.$message?.success($t('common.deleteSuccess'));
  await loadList();
}

const columns = computed<DataTableColumns<Api.Sys.RouteRow>>(() => [
  { title: () => $t('page.sysRoute.name'), key: 'name', width: 120 },
  { title: () => $t('page.sysRoute.path'), key: 'path', width: 140, ellipsis: { tooltip: true } },
  {
    title: () => $t('page.sysRoute.component'),
    key: 'component',
    minWidth: 220,
    ellipsis: { tooltip: true }
  },
  { title: () => $t('page.sysRoute.title'), key: 'title', width: 120 },
  { title: () => $t('page.sysRoute.order'), key: 'orderNo', width: 72 },
  {
    title: () => $t('page.sysRoute.status'),
    key: 'status',
    width: 72,
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
              h(
                NButton,
                { size: 'small', type: 'error', ghost: true, disabled: row.name === 'home' },
                { default: () => $t('common.delete') }
              )
          }
        )
      ])
  }
]);

onMounted(loadList);

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
        <NSpace justify="space-between" align="center">
          <NAlert type="info" class="min-w-0 flex-1">{{ $t('page.sysRoute.hint') }}</NAlert>
          <NButton type="primary" @click="openCreate">{{ $t('common.add') }}</NButton>
        </NSpace>
      </div>
      <div class="table-fill min-h-0 flex flex-1 flex-col px-16px pb-16px">
        <NDataTable :columns="columns" :data="list" :loading="loading" :scroll-x="1100" flex-height class="h-full" />
      </div>
    </NCard>

    <NModal v-model:show="showModal" preset="card" :title="isEdit ? $t('common.edit') : $t('common.add')" class="w-640px">
      <NForm ref="formRef" :model="formModel" :rules="rules" label-placement="left" label-width="96">
        <NFormItem path="parentId" :label="$t('page.sysRoute.parent')">
          <NSelect v-model:value="formModel.parentId" filterable clearable :options="parentOptions" />
        </NFormItem>
        <NFormItem path="name" :label="$t('page.sysRoute.name')">
          <NInput v-model:value="formModel.name" :disabled="isEdit && formModel.name === 'home'" />
        </NFormItem>
        <NFormItem path="path" :label="$t('page.sysRoute.path')">
          <NInput v-model:value="formModel.path" />
        </NFormItem>
        <NFormItem path="component" :label="$t('page.sysRoute.component')">
          <NInput v-model:value="formModel.component" />
        </NFormItem>
        <NFormItem path="title" :label="$t('page.sysRoute.title')">
          <NInput v-model:value="formModel.title" />
        </NFormItem>
        <NFormItem path="i18nKey" :label="$t('page.sysRoute.i18nKey')">
          <NInput v-model:value="formModel.i18nKey" />
        </NFormItem>
        <NFormItem path="icon" :label="$t('page.sysRoute.icon')">
          <NInput v-model:value="formModel.icon" />
        </NFormItem>
        <NFormItem path="orderNo" :label="$t('page.sysRoute.order')">
          <NInputNumber v-model:value="formModel.orderNo" class="w-full" />
        </NFormItem>
        <NFormItem path="hideInMenu" :label="$t('page.sysRoute.hideInMenu')">
          <NRadioGroup v-model:value="formModel.hideInMenu">
            <NSpace>
              <NRadio :value="1">{{ $t('common.yesOrNo.yes') }}</NRadio>
              <NRadio :value="0">{{ $t('common.yesOrNo.no') }}</NRadio>
            </NSpace>
          </NRadioGroup>
        </NFormItem>
        <NFormItem path="keepAlive" :label="$t('page.sysRoute.keepAlive')">
          <NRadioGroup v-model:value="formModel.keepAlive">
            <NSpace>
              <NRadio :value="1">{{ $t('common.yesOrNo.yes') }}</NRadio>
              <NRadio :value="0">{{ $t('common.yesOrNo.no') }}</NRadio>
            </NSpace>
          </NRadioGroup>
        </NFormItem>
        <NFormItem path="status" :label="$t('page.sysRoute.status')">
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

<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormRules } from 'naive-ui';
import { NButton, NPopconfirm } from 'naive-ui';
import { useBoolean } from '@sa/hooks';
import {
  fetchSysAnnouncementCreate,
  fetchSysAnnouncementDelete,
  fetchSysAnnouncementList,
  fetchSysAnnouncementUpdate
} from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'SysAnnouncement' });

const loading = ref(false);
const list = ref<Api.Biz.SysAnnouncement[]>([]);
const { bool: showModal, setBool: setShowModal } = useBoolean();
const isEdit = ref(false);
const formRef = ref();
const formModel = reactive({
  id: '',
  title: '',
  content: '',
  sortNo: 0,
  status: 1
});

const rules: FormRules = {
  title: { required: true, message: () => $t('page.sysAnnouncement.rules.title'), trigger: 'blur' },
  content: { required: true, message: () => $t('page.sysAnnouncement.rules.content'), trigger: 'blur' }
};

async function loadList() {
  loading.value = true;
  try {
    list.value = await fetchSysAnnouncementList();
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  isEdit.value = false;
  Object.assign(formModel, { id: '', title: '', content: '', sortNo: 10, status: 1 });
  setShowModal(true);
}

function openEdit(row: Api.Biz.SysAnnouncement) {
  isEdit.value = true;
  Object.assign(formModel, {
    id: row.id,
    title: row.title,
    content: row.content,
    sortNo: row.sortNo,
    status: row.status
  });
  setShowModal(true);
}

async function submitForm() {
  await formRef.value?.validate();
  if (isEdit.value) {
    await fetchSysAnnouncementUpdate({
      id: formModel.id,
      title: formModel.title,
      content: formModel.content,
      sortNo: formModel.sortNo,
      status: formModel.status
    });
    window.$message?.success($t('common.updateSuccess'));
  } else {
    await fetchSysAnnouncementCreate({
      title: formModel.title,
      content: formModel.content,
      sortNo: formModel.sortNo,
      status: formModel.status
    });
    window.$message?.success($t('common.addSuccess'));
  }
  setShowModal(false);
  await loadList();
}

async function handleDelete(row: Api.Biz.SysAnnouncement) {
  await fetchSysAnnouncementDelete(row.id);
  window.$message?.success($t('common.deleteSuccess'));
  await loadList();
}

const columns = computed<DataTableColumns<Api.Biz.SysAnnouncement>>(() => [
  { title: () => $t('page.sysAnnouncement.title'), key: 'title', minWidth: 160, ellipsis: { tooltip: true } },
  {
    title: () => $t('page.sysAnnouncement.content'),
    key: 'content',
    minWidth: 220,
    ellipsis: { tooltip: true }
  },
  { title: () => $t('page.sysAnnouncement.sortNo'), key: 'sortNo', width: 72 },
  {
    title: () => $t('page.sysAnnouncement.status'),
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
        <NDataTable :columns="columns" :data="list" :loading="loading" :scroll-x="900" flex-height class="h-full" />
      </div>
    </NCard>

    <NModal v-model:show="showModal" preset="card" :title="isEdit ? $t('common.edit') : $t('common.add')" class="w-600px">
      <NForm ref="formRef" :model="formModel" :rules="rules" label-placement="left" label-width="88">
        <NFormItem path="title" :label="$t('page.sysAnnouncement.title')">
          <NInput v-model:value="formModel.title" />
        </NFormItem>
        <NFormItem path="content" :label="$t('page.sysAnnouncement.content')">
          <NInput v-model:value="formModel.content" type="textarea" :rows="6" />
        </NFormItem>
        <NFormItem path="sortNo" :label="$t('page.sysAnnouncement.sortNo')">
          <NInputNumber v-model:value="formModel.sortNo" class="w-full" />
        </NFormItem>
        <NFormItem v-if="isEdit" path="status" :label="$t('page.sysAnnouncement.status')">
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

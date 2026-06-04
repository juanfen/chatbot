<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormRules } from 'naive-ui';
import { NButton, NPopconfirm } from 'naive-ui';
import { useBoolean } from '@sa/hooks';
import {
  fetchSysPackageList,
  fetchSysRoleList,
  fetchSysUserCreate,
  fetchSysUserDelete,
  fetchSysUserList,
  fetchSysUserUpdate
} from '@/service/api';
import { useAuthStore } from '@/store/modules/auth';
import { $t } from '@/locales';

defineOptions({ name: 'SysUser' });

const authStore = useAuthStore();

const keyword = ref('');
const loading = ref(false);
const list = ref<Api.Sys.User[]>([]);
const pagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });

const roleOptions = ref<{ label: string; value: string }[]>([]);
const roleMap = computed(() => Object.fromEntries(roleOptions.value.map(r => [r.value, r.label])));

const packageOptions = ref<{ label: string; value: string }[]>([]);

const { bool: showModal, setBool: setShowModal } = useBoolean();
const isEdit = ref(false);
const formRef = ref();
const formModel = reactive({
  id: '',
  username: '',
  password: '',
  nickname: '',
  status: 1,
  isAdmin: 0,
  roleIds: [] as string[],
  trialQuota: null as number | null,
  defaultPackageId: '' as string
});

const rules: FormRules = {
  username: { required: true, message: () => $t('page.sysUser.rules.username'), trigger: 'blur' },
  password: {
    validator() {
      if (!isEdit.value && !formModel.password) {
        return new Error($t('page.sysUser.rules.password'));
      }
      return true;
    },
    trigger: 'blur'
  }
};

async function loadRoles() {
  const data = await fetchSysRoleList();
  const roles = data.list ?? [];
  roleOptions.value = roles.map((r: Api.Sys.Role) => ({ label: `${r.name} (${r.code})`, value: r.id }));
}

async function loadPackages() {
  const rows = await fetchSysPackageList();
  packageOptions.value = rows.map(p => ({
    label: `${p.name} (${p.quotaCalls})`,
    value: p.id
  }));
}

async function loadList() {
  loading.value = true;
  try {
    const data = await fetchSysUserList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: keyword.value.trim() || undefined
    });
    list.value = data.list ?? [];
    pagination.itemCount = Number(data.total ?? 0);
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  isEdit.value = false;
  Object.assign(formModel, {
    id: '',
    username: '',
    password: '',
    nickname: '',
    status: 1,
    isAdmin: 0,
    roleIds: [],
    trialQuota: null,
    defaultPackageId: ''
  });
  setShowModal(true);
}

function openEdit(row: Api.Sys.User) {
  isEdit.value = true;
  Object.assign(formModel, {
    id: row.id,
    username: row.username,
    password: '',
    nickname: row.nickname,
    status: row.status,
    isAdmin: row.isAdmin,
    roleIds: [...row.roleIds],
    trialQuota: row.totalQuota != null ? Number(row.totalQuota) : null,
    defaultPackageId: ''
  });
  setShowModal(true);
}

async function submitForm() {
  await formRef.value?.validate();
  if (isEdit.value) {
    const body: Parameters<typeof fetchSysUserUpdate>[0] = {
      id: formModel.id,
      nickname: formModel.nickname,
      password: formModel.password || undefined,
      status: formModel.status,
      isAdmin: formModel.isAdmin,
      roleIds: formModel.roleIds
    };
    if (formModel.trialQuota != null) body.trialQuota = formModel.trialQuota;
    if (formModel.defaultPackageId) body.defaultPackageId = formModel.defaultPackageId;
    await fetchSysUserUpdate(body);
    window.$message?.success($t('common.updateSuccess'));
  } else {
    await fetchSysUserCreate({
      username: formModel.username,
      password: formModel.password,
      nickname: formModel.nickname,
      isAdmin: formModel.isAdmin,
      roleIds: formModel.roleIds,
      trialQuota: formModel.trialQuota ?? undefined,
      defaultPackageId: formModel.defaultPackageId || undefined
    });
    window.$message?.success($t('common.addSuccess'));
  }
  setShowModal(false);
  await loadList();
}

async function handleDelete(row: Api.Sys.User) {
  if (row.id === authStore.userInfo.userId) {
    window.$message?.warning($t('page.sysUser.cannotDeleteSelf'));
    return;
  }
  await fetchSysUserDelete(row.id);
  window.$message?.success($t('common.deleteSuccess'));
  await loadList();
}

const columns = computed<DataTableColumns<Api.Sys.User>>(() => [
  { title: () => $t('page.sysUser.username'), key: 'username', width: 140 },
  { title: () => $t('page.sysUser.nickname'), key: 'nickname', width: 140 },
  {
    title: () => $t('page.sysUser.status'),
    key: 'status',
    width: 90,
    render: row => (row.status === 1 ? $t('common.yesOrNo.yes') : $t('common.yesOrNo.no'))
  },
  {
    title: () => $t('page.sysUser.isAdmin'),
    key: 'isAdmin',
    width: 90,
    render: row => (row.isAdmin === 1 ? $t('common.yesOrNo.yes') : $t('common.yesOrNo.no'))
  },
  {
    title: () => $t('page.sysUser.roles'),
    key: 'roleIds',
    minWidth: 200,
    ellipsis: { tooltip: true },
    render: row =>
      row.roleIds?.length ? row.roleIds.map(id => roleMap.value[id] || id).join(', ') : $t('common.noData')
  },
  {
    title: () => $t('page.sysUser.quota'),
    key: 'quota',
    width: 140,
    render: row =>
      row.totalQuota != null || row.usedQuota != null
        ? `${row.usedQuota ?? 0}/${row.totalQuota ?? 0}`
        : '—'
  },
  {
    title: () => $t('common.action'),
    key: 'actions',
    width: 200,
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
                {
                  size: 'small',
                  type: 'error',
                  ghost: true,
                  disabled: row.id === authStore.userInfo.userId
                },
                { default: () => $t('common.delete') }
              )
          }
        )
      ])
  }
]);

function onPageChange(page: number) {
  pagination.page = page;
  loadList();
}

function onPageSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.page = 1;
  loadList();
}

onMounted(async () => {
  await loadRoles();
  await loadPackages();
  await loadList();
});

/** NCard 内容区：纵向 flex，表格区域 flex-1 撑满剩余高度（配合 NDataTable flex-height） */
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
          <NSpace>
            <NInput
              v-model:value="keyword"
              clearable
              :placeholder="$t('common.keywordSearch')"
              class="w-240px"
              @keyup.enter="loadList"
            />
            <NButton type="primary" @click="loadList">{{ $t('common.search') }}</NButton>
          </NSpace>
          <NButton type="primary" @click="openCreate">{{ $t('common.add') }}</NButton>
        </NSpace>
      </div>
      <div class="table-fill min-h-0 flex flex-1 flex-col px-16px pb-16px">
        <NDataTable
          :columns="columns"
          :data="list"
          :loading="loading"
          :scroll-x="1120"
          :pagination="{
            page: pagination.page,
            pageSize: pagination.pageSize,
            itemCount: pagination.itemCount,
            showSizePicker: true,
            pageSizes: [10, 20, 50],
            onUpdatePage: onPageChange,
            onUpdatePageSize: onPageSizeChange
          }"
          flex-height
          class="h-full"
        />
      </div>
    </NCard>

    <NModal v-model:show="showModal" preset="card" :title="isEdit ? $t('common.edit') : $t('common.add')" class="w-520px">
      <NForm ref="formRef" :model="formModel" :rules="rules" label-placement="left" label-width="96">
        <NFormItem path="username" :label="$t('page.sysUser.username')">
          <NInput v-model:value="formModel.username" :disabled="isEdit" />
        </NFormItem>
        <NFormItem path="password" :label="$t('page.sysUser.password')">
          <NInput
            v-model:value="formModel.password"
            type="password"
            show-password-on="click"
            :placeholder="isEdit ? $t('page.sysUser.passwordOptional') : ''"
          />
        </NFormItem>
        <NFormItem path="nickname" :label="$t('page.sysUser.nickname')">
          <NInput v-model:value="formModel.nickname" />
        </NFormItem>
        <NFormItem path="status" :label="$t('page.sysUser.status')">
          <NRadioGroup v-model:value="formModel.status">
            <NSpace>
              <NRadio :value="1">{{ $t('common.yesOrNo.yes') }}</NRadio>
              <NRadio :value="0">{{ $t('common.yesOrNo.no') }}</NRadio>
            </NSpace>
          </NRadioGroup>
        </NFormItem>
        <NFormItem path="isAdmin" :label="$t('page.sysUser.isAdmin')">
          <NRadioGroup v-model:value="formModel.isAdmin">
            <NSpace>
              <NRadio :value="1">{{ $t('common.yesOrNo.yes') }}</NRadio>
              <NRadio :value="0">{{ $t('common.yesOrNo.no') }}</NRadio>
            </NSpace>
          </NRadioGroup>
        </NFormItem>
        <NFormItem path="roleIds" :label="$t('page.sysUser.roles')">
          <NSelect
            v-model:value="formModel.roleIds"
            multiple
            clearable
            filterable
            :options="roleOptions"
            :placeholder="$t('page.sysUser.rolesPlaceholder')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.sysUser.defaultPackage')">
          <NSelect
            v-model:value="formModel.defaultPackageId"
            clearable
            filterable
            :options="packageOptions"
            :placeholder="$t('page.sysUser.defaultPackagePlaceholder')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.sysUser.trialQuota')">
          <NInputNumber v-model:value="formModel.trialQuota" class="w-full" :min="0" clearable :precision="0" />
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
  /* 主内容区为 flex 子项时，允许高度传递给表格（否则 flex-height 无效） */
  min-height: 0;
}
.table-fill :deep(.n-data-table) {
  flex: 1;
  min-height: 0;
}
</style>

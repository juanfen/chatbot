const local: App.I18n.Schema = {
  system: {
    title: 'ChatBot 管理后台',
    updateTitle: '系统版本更新通知',
    updateContent: '检测到系统有新版本发布，是否立即刷新页面？',
    updateConfirm: '立即刷新',
    updateCancel: '稍后再说'
  },
  common: {
    action: '操作',
    add: '新增',
    addSuccess: '添加成功',
    backToHome: '返回首页',
    batchDelete: '批量删除',
    cancel: '取消',
    close: '关闭',
    check: '勾选',
    selectAll: '全选',
    expandColumn: '展开列',
    columnSetting: '列设置',
    config: '配置',
    confirm: '确认',
    delete: '删除',
    deleteSuccess: '删除成功',
    confirmDelete: '确认删除吗？',
    edit: '编辑',
    warning: '警告',
    error: '错误',
    index: '序号',
    keywordSearch: '请输入关键词搜索',
    logout: '退出登录',
    logoutConfirm: '确认退出登录吗？',
    lookForward: '敬请期待',
    modify: '修改',
    modifySuccess: '修改成功',
    noData: '无数据',
    operate: '操作',
    pleaseCheckValue: '请检查输入的值是否合法',
    refresh: '刷新',
    reset: '重置',
    search: '搜索',
    switch: '切换',
    tip: '提示',
    trigger: '触发',
    update: '更新',
    updateSuccess: '更新成功',
    userCenter: '个人中心',
    yesOrNo: {
      yes: '是',
      no: '否'
    }
  },
  request: {
    logout: '请求失败后登出用户',
    logoutMsg: '用户状态失效，请重新登录',
    logoutWithModal: '请求失败后弹出模态框再登出用户',
    logoutWithModalMsg: '用户状态失效，请重新登录',
    refreshToken: '请求的token已过期，刷新token',
    tokenExpired: 'token已过期'
  },
  theme: {
    themeDrawerTitle: '主题配置',
    tabs: {
      appearance: '外观',
      layout: '布局',
      general: '通用',
      preset: '预设'
    },
    appearance: {
      themeSchema: {
        title: '主题模式',
        light: '亮色模式',
        dark: '暗黑模式',
        auto: '跟随系统'
      },
      grayscale: '灰色模式',
      colourWeakness: '色弱模式',
      themeColor: {
        title: '主题颜色',
        primary: '主色',
        info: '信息色',
        success: '成功色',
        warning: '警告色',
        error: '错误色',
        followPrimary: '跟随主色'
      },
      themeRadius: {
        title: '主题圆角'
      },
      recommendColor: '应用推荐算法的颜色',
      recommendColorDesc: '推荐颜色的算法参照',
      preset: {
        title: '主题预设',
        apply: '应用',
        applySuccess: '预设应用成功',
        default: {
          name: '默认预设',
          desc: 'Soybean 默认主题预设'
        },
        dark: {
          name: '暗色预设',
          desc: '适用于夜间使用的暗色主题预设'
        },
        compact: {
          name: '紧凑型',
          desc: '适用于小屏幕的紧凑布局预设'
        },
        azir: {
          name: 'Azir的预设',
          desc: '是 Azir 比较喜欢的莫兰迪色系冷淡风'
        }
      }
    },
    layout: {
      layoutMode: {
        title: '布局模式',
        vertical: '左侧菜单模式',
        'vertical-mix': '左侧菜单混合模式',
        'vertical-hybrid-header-first': '左侧混合-顶部优先',
        horizontal: '顶部菜单模式',
        'top-hybrid-sidebar-first': '顶部混合-侧边优先',
        'top-hybrid-header-first': '顶部混合-顶部优先',
        vertical_detail: '左侧菜单布局，菜单在左，内容在右。',
        'vertical-mix_detail': '左侧双菜单布局，一级菜单在左侧深色区域，二级菜单在左侧浅色区域。',
        'vertical-hybrid-header-first_detail':
          '左侧混合布局，一级菜单在顶部，二级菜单在左侧深色区域，三级菜单在左侧浅色区域。',
        horizontal_detail: '顶部菜单布局，菜单在顶部，内容在下方。',
        'top-hybrid-sidebar-first_detail': '顶部混合布局，一级菜单在左侧，二级菜单在顶部。',
        'top-hybrid-header-first_detail': '顶部混合布局，一级菜单在顶部，二级菜单在左侧。'
      },
      tab: {
        title: '标签栏设置',
        visible: '显示标签栏',
        cache: '标签栏信息缓存',
        cacheTip: '离开页面后仍然保留标签栏信息',
        height: '标签栏高度',
        mode: {
          title: '标签栏风格',
          slider: '滑块风格',
          chrome: '谷歌风格',
          button: '按钮风格'
        },
        closeByMiddleClick: '鼠标中键关闭标签页',
        closeByMiddleClickTip: '启用后可以使用鼠标中键点击标签页进行关闭'
      },
      header: {
        title: '头部设置',
        height: '头部高度',
        breadcrumb: {
          visible: '显示面包屑',
          showIcon: '显示面包屑图标'
        }
      },
      sider: {
        title: '侧边栏设置',
        inverted: '深色侧边栏',
        width: '侧边栏宽度',
        collapsedWidth: '侧边栏折叠宽度',
        mixWidth: '混合布局侧边栏宽度',
        mixCollapsedWidth: '混合布局侧边栏折叠宽度',
        mixChildMenuWidth: '混合布局子菜单宽度',
        autoSelectFirstMenu: '自动选择第一个子菜单',
        autoSelectFirstMenuTip: '点击一级菜单时，自动选择并导航到第一个子菜单的最深层级'
      },
      footer: {
        title: '底部设置',
        visible: '显示底部',
        fixed: '固定底部',
        height: '底部高度',
        right: '底部居右'
      },
      content: {
        title: '内容区域设置',
        scrollMode: {
          title: '滚动模式',
          tip: '主题滚动仅 main 部分滚动，外层滚动可携带头部底部一起滚动',
          wrapper: '外层滚动',
          content: '主体滚动'
        },
        page: {
          animate: '页面切换动画',
          mode: {
            title: '页面切换动画类型',
            'fade-slide': '滑动',
            fade: '淡入淡出',
            'fade-bottom': '底部消退',
            'fade-scale': '缩放消退',
            'zoom-fade': '渐变',
            'zoom-out': '闪现',
            none: '无'
          }
        },
        fixedHeaderAndTab: '固定头部和标签栏'
      }
    },
    general: {
      title: '通用设置',
      watermark: {
        title: '水印设置',
        visible: '显示全屏水印',
        text: '自定义水印文本',
        enableUserName: '启用用户名水印',
        enableTime: '显示当前时间',
        timeFormat: '时间格式'
      },
      multilingual: {
        title: '多语言设置',
        visible: '显示多语言按钮'
      },
      globalSearch: {
        title: '全局搜索设置',
        visible: '显示全局搜索按钮'
      }
    },
    configOperation: {
      copyConfig: '复制配置',
      copySuccessMsg: '复制成功，请替换 src/theme/settings.ts 中的变量 themeSettings',
      resetConfig: '重置配置',
      resetSuccessMsg: '重置成功'
    }
  },
  route: {
    login: '登录',
    403: '无权限',
    404: '页面不存在',
    500: '服务器错误',
    'iframe-page': '外链页面',
    home: '首页',
    'package-mall': '套餐购买',
    'sys-user': '用户管理',
    'sys-role': '角色管理',
    'sys-route': '菜单路由',
    'sys-package': '套餐配置',
    'sys-announcement': '公告管理',
    'sys-runtime-config': '算法运行时配置',
    'voice-recognition': '音色识别'
  },
  page: {
    login: {
      common: {
        loginOrRegister: '登录 / 注册',
        userNamePlaceholder: '请输入用户名',
        phonePlaceholder: '请输入手机号',
        codePlaceholder: '请输入验证码',
        passwordPlaceholder: '请输入密码',
        confirmPasswordPlaceholder: '请再次输入密码',
        codeLogin: '验证码登录',
        confirm: '确定',
        back: '返回',
        validateSuccess: '验证成功',
        loginSuccess: '登录成功',
        welcomeBack: '欢迎回来，{userName} ！'
      },
      pwdLogin: {
        title: '密码登录',
        rememberMe: '记住我',
        forgetPassword: '忘记密码？',
        register: '注册账号',
        otherAccountLogin: '其他账号登录',
        otherLoginMode: '其他登录方式',
        superAdmin: '超级管理员',
        admin: '管理员',
        user: '普通用户'
      },
      codeLogin: {
        title: '验证码登录',
        getCode: '获取验证码',
        reGetCode: '{time}秒后重新获取',
        sendCodeSuccess: '验证码发送成功',
        imageCodePlaceholder: '请输入图片验证码'
      },
      register: {
        title: '注册账号',
        agreement: '我已经仔细阅读并接受',
        protocol: '《用户协议》',
        policy: '《隐私权政策》'
      },
      resetPwd: {
        title: '重置密码'
      },
      bindWeChat: {
        title: '绑定微信'
      }
    },
    home: {
      branchDesc:
        '为了方便大家开发和更新合并，我们对main分支的代码进行了精简，只保留了首页菜单，其余内容已移至example分支进行维护。预览地址显示的内容即为example分支的内容。',
      greeting: '早安，{userName}, 今天又是充满活力的一天!',
      weatherDesc: '今日多云转晴，20℃ - 25℃!',
      projectCount: '项目数',
      todo: '待办',
      message: '消息',
      downloadCount: '下载量',
      registerCount: '注册量',
      schedule: '作息安排',
      study: '学习',
      work: '工作',
      rest: '休息',
      entertainment: '娱乐',
      visitCount: '访问量',
      turnover: '成交额',
      dealCount: '成交量',
      projectNews: {
        title: '项目动态',
        moreNews: '更多动态',
        desc1: 'Soybean 在2021年5月28日创建了开源项目 soybean-admin!',
        desc2: 'Yanbowe 向 soybean-admin 提交了一个bug，多标签栏不会自适应。',
        desc3: 'Soybean 准备为 soybean-admin 的发布做充分的准备工作!',
        desc4: 'Soybean 正在忙于为soybean-admin写项目说明文档！',
        desc5: 'Soybean 刚才把工作台页面随便写了一些，凑合能看了！'
      },
      creativity: '创意'
    },
    homeBiz: {
      quotaTitle: '我的套餐与额度',
      unlimited: '管理员账号不限制调用次数',
      packageName: '当前套餐',
      totalQuota: '总额度',
      usedQuota: '已使用',
      remainingQuota: '剩余',
      simulateExperience: '模拟一次体验链路调用',
      simulateOk: '已记录一次调用',
      callTrend: '近 14 日调用趋势',
      callCount: '调用次数',
      usageDetail: '调用明细',
      logTime: '时间',
      channel: '渠道',
      ref: '备注',
      announcements: '官方公告'
    },
    packageMall: {
      currentPlan: '当前套餐',
      unlimited: '不限额度',
      remaining: '剩余/总额',
      quotaCalls: '含 {n} 次调用',
      mockBuy: '模拟购买（立即生效）',
      purchaseOk: '购买成功'
    },
    sysPackage: {
      code: '套餐编码',
      name: '名称',
      description: '说明',
      quotaCalls: '调用次数',
      sortNo: '排序',
      status: '上架',
      rules: {
        code: '请输入套餐编码',
        name: '请输入名称'
      }
    },
    sysAnnouncement: {
      title: '标题',
      content: '正文',
      sortNo: '排序',
      status: '发布',
      rules: {
        title: '请输入标题',
        content: '请输入正文'
      }
    },
    voiceRecognition: {
      hint: '选择一条未绑定的设备音频，为其分配音色名称；绑定后该记录永久保留。',
      select: '选择',
      macId: '设备 MAC',
      audioLen: '时长',
      context: '文本内容',
      speakerName: '音色名称',
      audio: '音频',
      play: '播放',
      stop: '停止',
      createTime: '创建时间',
      expireTime: '过期时间',
      perpetual: '永久',
      assignSpeaker: '分配音色',
      useExisting: '选择已有音色',
      createNew: '新建音色',
      pickSpeaker: '请选择音色',
      speakerNamePlaceholder: '请输入音色名称',
      selectRowFirst: '请先选择一条记录',
      alreadyBound: '该记录已绑定音色',
      bindSuccess: '绑定成功',
      enterSpeakerName: '请输入音色名称',
      playFailed: '音频播放失败',
      boundOther: '已绑定',
      personalContent: '个性化内容',
      personalTitle: '音色个性化内容',
      personalHint: '为每个音色名称配置多条个性化说明，识别到该说话人后会拼入系统提示；每条不超过 20 字。输入后点击「添加一条」，或直接保存（未添加的输入会自动加入）。',
      pickSpeakerForPersonal: '选择音色',
      personalLinePlaceholder: '输入一条个性化内容（最多 20 字）',
      addPersonalLine: '添加一条',
      personalLineTooLong: '单条不能超过 20 字',
      personalLineEmpty: '请输入内容后再添加',
      personalSaveOk: '个性化内容已保存',
      personalNoSpeaker: '请先在「分配音色」中创建至少一个音色',
      personalSaveEmpty: '请至少添加一条个性化内容后再保存',
      personalConfiguredTitle: '已配置的音色',
      personalConfiguredLines: '已配置条目',
      personalNotConfigured: '尚未配置',
      personalLineCount: '{count} 条'
    },
    sysRuntimeConfig: {
      pageTitle: '算法运行时配置',
      pageHint:
        '该配置用于热更新 xiaozhi-go-core 的运行时参数，仅修改进程内存快照，不会写入 config.json，重启进程后失效。',
      reload: '重新拉取',
      submitSection: '保存当前分区',
      submitAll: '保存全部修改',
      reset: '重置为已生效',
      resetSection: '重置当前分区',
      diffOnly: '仅提交修改的分区',
      diffOnlyHint: '已检测到 {n} 个分区有改动',
      noChanges: '没有改动',
      jsonInvalid: 'JSON 格式不合法',
      jsonInvalidSection: '{section} 的 JSON 不合法：{msg}',
      saveOkSection: '{section} 已生效',
      saveOkAll: '全部分区已生效',
      sections: {
        gateway: '网关 gateway',
        ota: 'OTA',
        llm: 'LLM 大模型',
        tts: 'TTS 语音合成',
        vad: 'VAD 静音检测',
        asr: 'ASR 语音识别',
        voiceprint: '声纹 voiceprint',
        chatbot_admin: '设备音频上报 chatbot_admin',
        audio_pipeline: '音频流水线 audio_pipeline'
      },
      viewJson: 'JSON',
      viewForm: '表单',
      formatJson: '格式化',
      form: {
        addField: '添加字段',
        addItem: '添加项'
      }
    },
    sysUser: {
      username: '用户名',
      password: '密码',
      passwordOptional: '留空表示不修改密码',
      nickname: '昵称',
      status: '启用',
      isAdmin: '管理员',
      roles: '角色',
      rolesPlaceholder: '请选择角色',
      cannotDeleteSelf: '不能删除当前登录用户',
      quota: '额度(已用/总额)',
      defaultPackage: '初始套餐',
      defaultPackagePlaceholder: '可选，优先于体验次数',
      trialQuota: '体验调用次数',
      rules: {
        username: '请输入用户名',
        password: '请输入初始密码'
      }
    },
    sysRole: {
      code: '编码',
      name: '名称',
      status: '启用',
      routeCount: '路由数',
      routes: '路由权限',
      routesTitle: '配置该角色的路由',
      rules: {
        code: '请输入角色编码',
        name: '请输入角色名称'
      }
    },
    sysRoute: {
      hint: '新增菜单后请在「角色管理」中为角色分配路由；组件需与 Elegant 视图一致，例如 layout.base$view.sys-user。',
      root: '（顶级）',
      parent: '父级',
      name: '路由名',
      path: '路径',
      component: '组件',
      title: '标题',
      i18nKey: 'i18n 键',
      icon: '图标',
      order: '排序',
      hideInMenu: '菜单中隐藏',
      keepAlive: '缓存页面',
      status: '启用',
      rules: {
        name: '请输入路由名称',
        path: '请输入路径',
        component: '请输入组件路径'
      }
    }
  },
  form: {
    required: '不能为空',
    userName: {
      required: '请输入用户名',
      invalid: '用户名格式不正确'
    },
    phone: {
      required: '请输入手机号',
      invalid: '手机号格式不正确'
    },
    pwd: {
      required: '请输入密码',
      invalid: '密码格式不正确，6-18位字符，包含字母、数字、下划线'
    },
    confirmPwd: {
      required: '请输入确认密码',
      invalid: '两次输入密码不一致'
    },
    code: {
      required: '请输入验证码',
      invalid: '验证码格式不正确'
    },
    email: {
      required: '请输入邮箱',
      invalid: '邮箱格式不正确'
    }
  },
  dropdown: {
    closeCurrent: '关闭',
    closeOther: '关闭其它',
    closeLeft: '关闭左侧',
    closeRight: '关闭右侧',
    closeAll: '关闭所有',
    pin: '固定标签',
    unpin: '取消固定'
  },
  icon: {
    themeConfig: '主题配置',
    themeSchema: '主题模式',
    lang: '切换语言',
    fullscreen: '全屏',
    fullscreenExit: '退出全屏',
    reload: '刷新页面',
    collapse: '折叠菜单',
    expand: '展开菜单',
    pin: '固定',
    unpin: '取消固定'
  },
  datatable: {
    itemCount: '共 {total} 条',
    fixed: {
      left: '左固定',
      right: '右固定',
      unFixed: '取消固定'
    }
  }
};

export default local;

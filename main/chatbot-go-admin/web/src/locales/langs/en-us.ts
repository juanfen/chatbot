const local: App.I18n.Schema = {
  system: {
    title: 'SoybeanAdmin',
    updateTitle: 'System Version Update Notification',
    updateContent: 'A new version of the system has been detected. Do you want to refresh the page immediately?',
    updateConfirm: 'Refresh immediately',
    updateCancel: 'Later'
  },
  common: {
    action: 'Action',
    add: 'Add',
    addSuccess: 'Add Success',
    backToHome: 'Back to home',
    batchDelete: 'Batch Delete',
    cancel: 'Cancel',
    close: 'Close',
    check: 'Check',
    selectAll: 'Select All',
    expandColumn: 'Expand Column',
    columnSetting: 'Column Setting',
    config: 'Config',
    confirm: 'Confirm',
    delete: 'Delete',
    deleteSuccess: 'Delete Success',
    confirmDelete: 'Are you sure you want to delete?',
    edit: 'Edit',
    warning: 'Warning',
    error: 'Error',
    index: 'Index',
    keywordSearch: 'Please enter keyword',
    logout: 'Logout',
    logoutConfirm: 'Are you sure you want to log out?',
    lookForward: 'Coming soon',
    modify: 'Modify',
    modifySuccess: 'Modify Success',
    noData: 'No Data',
    operate: 'Operate',
    pleaseCheckValue: 'Please check whether the value is valid',
    refresh: 'Refresh',
    reset: 'Reset',
    search: 'Search',
    switch: 'Switch',
    tip: 'Tip',
    trigger: 'Trigger',
    update: 'Update',
    updateSuccess: 'Update Success',
    userCenter: 'User Center',
    yesOrNo: {
      yes: 'Yes',
      no: 'No'
    }
  },
  request: {
    logout: 'Logout user after request failed',
    logoutMsg: 'User status is invalid, please log in again',
    logoutWithModal: 'Pop up modal after request failed and then log out user',
    logoutWithModalMsg: 'User status is invalid, please log in again',
    refreshToken: 'The requested token has expired, refresh the token',
    tokenExpired: 'The requested token has expired'
  },
  theme: {
    themeDrawerTitle: 'Theme Configuration',
    tabs: {
      appearance: 'Appearance',
      layout: 'Layout',
      general: 'General',
      preset: 'Preset'
    },
    appearance: {
      themeSchema: {
        title: 'Theme Schema',
        light: 'Light',
        dark: 'Dark',
        auto: 'Follow System'
      },
      grayscale: 'Grayscale',
      colourWeakness: 'Colour Weakness',
      themeColor: {
        title: 'Theme Color',
        primary: 'Primary',
        info: 'Info',
        success: 'Success',
        warning: 'Warning',
        error: 'Error',
        followPrimary: 'Follow Primary'
      },
      themeRadius: {
        title: 'Theme Radius'
      },
      recommendColor: 'Apply Recommended Color Algorithm',
      recommendColorDesc: 'The recommended color algorithm refers to',
      preset: {
        title: 'Theme Presets',
        apply: 'Apply',
        applySuccess: 'Preset applied successfully',
        default: {
          name: 'Default Preset',
          desc: 'Default theme preset with balanced settings'
        },
        dark: {
          name: 'Dark Preset',
          desc: 'Dark theme preset for night time usage'
        },
        compact: {
          name: 'Compact Preset',
          desc: 'Compact layout preset for small screens'
        },
        azir: {
          name: "Azir's Preset",
          desc: 'It is a cold and elegant preset that Azir likes'
        }
      }
    },
    layout: {
      layoutMode: {
        title: 'Layout Mode',
        vertical: 'Vertical Mode',
        horizontal: 'Horizontal Mode',
        'vertical-mix': 'Vertical Mix Mode',
        'vertical-hybrid-header-first': 'Left Hybrid Header-First',
        'top-hybrid-sidebar-first': 'Top-Hybrid Sidebar-First',
        'top-hybrid-header-first': 'Top-Hybrid Header-First',
        vertical_detail: 'Vertical menu layout, with the menu on the left and content on the right.',
        'vertical-mix_detail':
          'Vertical mix-menu layout, with the primary menu on the dark left side and the secondary menu on the lighter left side.',
        'vertical-hybrid-header-first_detail':
          'Left hybrid layout, with the primary menu at the top, the secondary menu on the dark left side, and the tertiary menu on the lighter left side.',
        horizontal_detail: 'Horizontal menu layout, with the menu at the top and content below.',
        'top-hybrid-sidebar-first_detail':
          'Top hybrid layout, with the primary menu on the left and the secondary menu at the top.',
        'top-hybrid-header-first_detail':
          'Top hybrid layout, with the primary menu at the top and the secondary menu on the left.'
      },
      tab: {
        title: 'Tab Settings',
        visible: 'Tab Visible',
        cache: 'Tag Bar Info Cache',
        cacheTip: 'Keep the tab bar information after leaving the page',
        height: 'Tab Height',
        mode: {
          title: 'Tab Mode',
          slider: 'Slider',
          chrome: 'Chrome',
          button: 'Button'
        },
        closeByMiddleClick: 'Close Tab by Middle Click',
        closeByMiddleClickTip: 'Enable closing tabs by clicking with the middle mouse button'
      },
      header: {
        title: 'Header Settings',
        height: 'Header Height',
        breadcrumb: {
          visible: 'Breadcrumb Visible',
          showIcon: 'Breadcrumb Icon Visible'
        }
      },
      sider: {
        title: 'Sider Settings',
        inverted: 'Dark Sider',
        width: 'Sider Width',
        collapsedWidth: 'Sider Collapsed Width',
        mixWidth: 'Mix Sider Width',
        mixCollapsedWidth: 'Mix Sider Collapse Width',
        mixChildMenuWidth: 'Mix Child Menu Width',
        autoSelectFirstMenu: 'Auto Select First Submenu',
        autoSelectFirstMenuTip:
          'When a first-level menu is clicked, the first submenu is automatically selected and navigated to the deepest level'
      },
      footer: {
        title: 'Footer Settings',
        visible: 'Footer Visible',
        fixed: 'Fixed Footer',
        height: 'Footer Height',
        right: 'Right Footer'
      },
      content: {
        title: 'Content Area Settings',
        scrollMode: {
          title: 'Scroll Mode',
          tip: 'The theme scroll only scrolls the main part, the outer scroll can carry the header and footer together',
          wrapper: 'Wrapper',
          content: 'Content'
        },
        page: {
          animate: 'Page Animate',
          mode: {
            title: 'Page Animate Mode',
            fade: 'Fade',
            'fade-slide': 'Slide',
            'fade-bottom': 'Fade Zoom',
            'fade-scale': 'Fade Scale',
            'zoom-fade': 'Zoom Fade',
            'zoom-out': 'Zoom Out',
            none: 'None'
          }
        },
        fixedHeaderAndTab: 'Fixed Header And Tab'
      }
    },
    general: {
      title: 'General Settings',
      watermark: {
        title: 'Watermark Settings',
        visible: 'Watermark Full Screen Visible',
        text: 'Custom Watermark Text',
        enableUserName: 'Enable User Name Watermark',
        enableTime: 'Show Current Time',
        timeFormat: 'Time Format'
      },
      multilingual: {
        title: 'Multilingual Settings',
        visible: 'Display multilingual button'
      },
      globalSearch: {
        title: 'Global Search Settings',
        visible: 'Display GlobalSearch button'
      }
    },
    configOperation: {
      copyConfig: 'Copy Config',
      copySuccessMsg: 'Copy Success, Please replace the variable "themeSettings" in "src/theme/settings.ts"',
      resetConfig: 'Reset Config',
      resetSuccessMsg: 'Reset Success'
    }
  },
  route: {
    login: 'Login',
    403: 'No Permission',
    404: 'Page Not Found',
    500: 'Server Error',
    'iframe-page': 'Iframe',
    home: 'Home',
    'package-mall': 'Packages',
    'sys-user': 'Users',
    'sys-role': 'Roles',
    'sys-route': 'Routes',
    'sys-package': 'Package plans',
    'sys-announcement': 'Announcements',
    'sys-runtime-config': 'Runtime config',
    'voice-recognition': 'Voice recognition'
  },
  page: {
    login: {
      common: {
        loginOrRegister: 'Login / Register',
        userNamePlaceholder: 'Please enter user name',
        phonePlaceholder: 'Please enter phone number',
        codePlaceholder: 'Please enter verification code',
        passwordPlaceholder: 'Please enter password',
        confirmPasswordPlaceholder: 'Please enter password again',
        codeLogin: 'Verification code login',
        confirm: 'Confirm',
        back: 'Back',
        validateSuccess: 'Verification passed',
        loginSuccess: 'Login successfully',
        welcomeBack: 'Welcome back, {userName} !'
      },
      pwdLogin: {
        title: 'Password Login',
        rememberMe: 'Remember me',
        forgetPassword: 'Forget password?',
        register: 'Register',
        otherAccountLogin: 'Other Account Login',
        otherLoginMode: 'Other Login Mode',
        superAdmin: 'Super Admin',
        admin: 'Admin',
        user: 'User'
      },
      codeLogin: {
        title: 'Verification Code Login',
        getCode: 'Get verification code',
        reGetCode: 'Reacquire after {time}s',
        sendCodeSuccess: 'Verification code sent successfully',
        imageCodePlaceholder: 'Please enter image verification code'
      },
      register: {
        title: 'Register',
        agreement: 'I have read and agree to',
        protocol: '《User Agreement》',
        policy: '《Privacy Policy》'
      },
      resetPwd: {
        title: 'Reset Password'
      },
      bindWeChat: {
        title: 'Bind WeChat'
      }
    },
    home: {
      branchDesc:
        'For the convenience of everyone in developing and updating the merge, we have streamlined the code of the main branch, only retaining the homepage menu, and the rest of the content has been moved to the example branch for maintenance. The preview address displays the content of the example branch.',
      greeting: 'Good morning, {userName}, today is another day full of vitality!',
      weatherDesc: 'Today is cloudy to clear, 20℃ - 25℃!',
      projectCount: 'Project Count',
      todo: 'Todo',
      message: 'Message',
      downloadCount: 'Download Count',
      registerCount: 'Register Count',
      schedule: 'Work and rest Schedule',
      study: 'Study',
      work: 'Work',
      rest: 'Rest',
      entertainment: 'Entertainment',
      visitCount: 'Visit Count',
      turnover: 'Turnover',
      dealCount: 'Deal Count',
      projectNews: {
        title: 'Project News',
        moreNews: 'More News',
        desc1: 'Soybean created the open source project soybean-admin on May 28, 2021!',
        desc2: 'Yanbowe submitted a bug to soybean-admin, the multi-tab bar will not adapt.',
        desc3: 'Soybean is ready to do sufficient preparation for the release of soybean-admin!',
        desc4: 'Soybean is busy writing project documentation for soybean-admin!',
        desc5: 'Soybean just wrote some of the workbench pages casually, and it was enough to see!'
      },
      creativity: 'Creativity'
    },
    homeBiz: {
      quotaTitle: 'Plan & quota',
      unlimited: 'Administrators have unlimited calls',
      packageName: 'Current plan',
      totalQuota: 'Total quota',
      usedQuota: 'Used',
      remainingQuota: 'Remaining',
      simulateExperience: 'Simulate one experience call',
      simulateOk: 'Call recorded',
      callTrend: 'Calls in the last 14 days',
      callCount: 'Calls',
      usageDetail: 'Usage log',
      logTime: 'Time',
      channel: 'Channel',
      ref: 'Note',
      announcements: 'Announcements'
    },
    packageMall: {
      currentPlan: 'Current plan',
      unlimited: 'Unlimited',
      remaining: 'Remaining / total',
      quotaCalls: '{n} calls included',
      mockBuy: 'Mock purchase (instant)',
      purchaseOk: 'Purchase successful'
    },
    sysPackage: {
      code: 'Code',
      name: 'Name',
      description: 'Description',
      quotaCalls: 'Call quota',
      sortNo: 'Sort',
      status: 'Active',
      rules: {
        code: 'Please enter package code',
        name: 'Please enter name'
      }
    },
    sysAnnouncement: {
      title: 'Title',
      content: 'Body',
      sortNo: 'Sort',
      status: 'Published',
      rules: {
        title: 'Please enter title',
        content: 'Please enter content'
      }
    },
    voiceRecognition: {
      hint: 'Select an unbound device audio row and assign a speaker name. Bound rows are kept permanently.',
      select: 'Select',
      macId: 'MAC',
      audioLen: 'Length',
      context: 'Text',
      speakerName: 'Speaker',
      audio: 'Audio',
      play: 'Play',
      stop: 'Stop',
      createTime: 'Created',
      expireTime: 'Expires',
      perpetual: 'Permanent',
      assignSpeaker: 'Assign speaker',
      useExisting: 'Existing speaker',
      createNew: 'New speaker',
      pickSpeaker: 'Select speaker',
      speakerNamePlaceholder: 'Speaker name',
      selectRowFirst: 'Select a row first',
      alreadyBound: 'Already bound',
      bindSuccess: 'Bound successfully',
      enterSpeakerName: 'Enter speaker name',
      playFailed: 'Playback failed',
      boundOther: 'Bound',
      personalContent: 'Personalization',
      personalTitle: 'Speaker personalization',
      personalHint: 'Add short lines per speaker; injected into system prompt when recognized. Max 20 chars per line. Click Add line, or save directly (pending input is included).',
      pickSpeakerForPersonal: 'Select speaker',
      personalLinePlaceholder: 'One line (max 20 chars)',
      addPersonalLine: 'Add line',
      personalLineTooLong: 'Each line must be at most 20 characters',
      personalLineEmpty: 'Enter text before adding',
      personalSaveOk: 'Personalization saved',
      personalNoSpeaker: 'Create at least one speaker via Assign speaker first',
      personalSaveEmpty: 'Add at least one line before saving',
      personalConfiguredTitle: 'Configured speakers',
      personalConfiguredLines: 'Saved lines',
      personalNotConfigured: 'Not configured',
      personalLineCount: '{count} lines'
    },
    sysRuntimeConfig: {
      pageTitle: 'Runtime config',
      pageHint:
        'Hot-update xiaozhi-go-core runtime parameters. Only the in-memory snapshot is updated; config.json is not modified and changes are lost on restart.',
      reload: 'Reload',
      submitSection: 'Save this section',
      submitAll: 'Save all changes',
      reset: 'Reset to live',
      resetSection: 'Reset this section',
      diffOnly: 'Submit only changed sections',
      diffOnlyHint: '{n} section(s) modified',
      noChanges: 'No changes',
      jsonInvalid: 'Invalid JSON',
      jsonInvalidSection: 'Invalid JSON in {section}: {msg}',
      saveOkSection: '{section} applied',
      saveOkAll: 'All sections applied',
      sections: {
        gateway: 'Gateway',
        ota: 'OTA',
        llm: 'LLM',
        tts: 'TTS',
        vad: 'VAD',
        asr: 'ASR',
        voiceprint: 'Voiceprint',
        chatbot_admin: 'Device audio upload',
        audio_pipeline: 'Audio pipeline'
      },
      viewJson: 'JSON',
      viewForm: 'Form',
      formatJson: 'Format',
      form: {
        addField: 'Add field',
        addItem: 'Add item'
      }
    },
    sysUser: {
      username: 'Username',
      password: 'Password',
      passwordOptional: 'Leave blank to keep unchanged',
      nickname: 'Nickname',
      status: 'Enabled',
      isAdmin: 'Administrator',
      roles: 'Roles',
      rolesPlaceholder: 'Select roles',
      cannotDeleteSelf: 'Cannot delete the signed-in user',
      quota: 'Quota (used/total)',
      defaultPackage: 'Initial package',
      defaultPackagePlaceholder: 'Optional; overrides trial count when set',
      trialQuota: 'Trial call quota',
      rules: {
        username: 'Please enter username',
        password: 'Please enter initial password'
      }
    },
    sysRole: {
      code: 'Code',
      name: 'Name',
      status: 'Enabled',
      routeCount: 'Routes',
      routes: 'Route permissions',
      routesTitle: 'Configure routes for this role',
      rules: {
        code: 'Please enter role code',
        name: 'Please enter role name'
      }
    },
    sysRoute: {
      hint: 'After adding a menu, assign routes to roles in Role management; component must match Elegant views, e.g. layout.base$view.sys-user.',
      root: '(Top level)',
      parent: 'Parent',
      name: 'Route name',
      path: 'Path',
      component: 'Component',
      title: 'Title',
      i18nKey: 'i18n Key',
      icon: 'Icon',
      order: 'Order',
      hideInMenu: 'Hide in menu',
      keepAlive: 'Keep alive',
      status: 'Enabled',
      rules: {
        name: 'Please enter route name',
        path: 'Please enter path',
        component: 'Please enter component key'
      }
    }
  },
  form: {
    required: 'Cannot be empty',
    userName: {
      required: 'Please enter user name',
      invalid: 'User name format is incorrect'
    },
    phone: {
      required: 'Please enter phone number',
      invalid: 'Phone number format is incorrect'
    },
    pwd: {
      required: 'Please enter password',
      invalid: '6-18 characters, including letters, numbers, and underscores'
    },
    confirmPwd: {
      required: 'Please enter password again',
      invalid: 'The two passwords are inconsistent'
    },
    code: {
      required: 'Please enter verification code',
      invalid: 'Verification code format is incorrect'
    },
    email: {
      required: 'Please enter email',
      invalid: 'Email format is incorrect'
    }
  },
  dropdown: {
    closeCurrent: 'Close Current',
    closeOther: 'Close Other',
    closeLeft: 'Close Left',
    closeRight: 'Close Right',
    closeAll: 'Close All',
    pin: 'Pin Tab',
    unpin: 'Unpin Tab'
  },
  icon: {
    themeConfig: 'Theme Configuration',
    themeSchema: 'Theme Schema',
    lang: 'Switch Language',
    fullscreen: 'Fullscreen',
    fullscreenExit: 'Exit Fullscreen',
    reload: 'Reload Page',
    collapse: 'Collapse Menu',
    expand: 'Expand Menu',
    pin: 'Pin',
    unpin: 'Unpin'
  },
  datatable: {
    itemCount: 'Total {total} items',
    fixed: {
      left: 'Left Fixed',
      right: 'Right Fixed',
      unFixed: 'Unfixed'
    }
  }
};

export default local;

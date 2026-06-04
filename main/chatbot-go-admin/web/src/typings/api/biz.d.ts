declare namespace Api {
  namespace Biz {
    interface QuotaSummary {
      isUnlimited: boolean;
      packageId: string;
      packageTitle: string;
      totalQuota: number;
      usedQuota: number;
      remainingQuota: number;
    }

    interface UsageTrendPoint {
      date: string;
      count: number;
    }

    interface UsageLogItem {
      id: string;
      channel: string;
      refText: string;
      createTime: string;
    }

    interface UsageList {
      total: number;
      list: UsageLogItem[];
    }

    interface Announcement {
      id: string;
      title: string;
      content: string;
      publishTime: string | null;
    }

    interface PackageCatalogItem {
      id: string;
      code: string;
      name: string;
      description: string;
      quotaCalls: number;
    }

    interface PurchaseResult {
      ok: boolean;
      packageTitle: string;
      totalQuota: number;
      usedQuota: number;
      remainingQuota: number;
      message: string;
    }

    interface SysPackage {
      id: string;
      code: string;
      name: string;
      description: string;
      quotaCalls: number;
      sortNo: number;
      status: number;
    }

    interface SysAnnouncement {
      id: string;
      title: string;
      content: string;
      sortNo: number;
      status: number;
      publishTime: string | null;
    }

    interface DeviceAudioAdminItem {
      id: string;
      macId: string;
      audioLen: string;
      context: string;
      speakerId?: string;
      speakerName?: string;
      bound: boolean;
      perpetual: boolean;
      createTime: string;
      expireTime: string;
      deleteTime: string;
    }

    interface DeviceAudioAdminList {
      total: number;
      list: DeviceAudioAdminItem[];
    }

    interface UserSpeaker {
      speakerId: string;
      speakerName: string;
      personalLines?: string[];
      createTime: string;
    }

    interface SpeakerPersonalLinesResp {
      speakerId: string;
      speakerName: string;
      personalLines: string[];
    }

    interface DeviceAudioBindResult {
      deviceAudioId: string;
      speakerId: string;
      speakerName: string;
    }
  }
}

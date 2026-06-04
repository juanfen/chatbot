# Android Flutter 蓝牙 Wi‑Fi 配网业务说明

本文档从原 uni-app 微信小程序中的配网实现抽象而来，供 **仅在 Flutter（Android）侧** 实现同等配网能力时使用；**不要求继续维护小程序**。若与硬件固件约定不一致，以硬件协议文档为准，并更新本文第 9 节「合同清单」。

---

## 1. 产品目标与边界

**目标**：用户在 Android 手机上，通过 **BLE** 把 **路由器 SSID + 密码** 发给智能硬件，并触发设备去连该 Wi‑Fi；可选查询连接结果。

**边界**：

- 手机不负责替设备连路由器，只负责 **下发凭证 + 控制命令**。
- **Wi‑Fi 热点列表**来自 **Android 系统 Wi‑Fi 扫描**，不是 BLE 扫出来的。
- 下发格式为 **二进制帧**（按 BluFi 思路的简化实现），**不是 JSON**。

---

## 2. 用户侧主流程（状态机）

1. **进入配网**  
   - 检查/申请：蓝牙已开、蓝牙权限、定位权限（Android 12+ 的扫描/连接权限与「用于 Wi‑Fi/BLE 扫描的定位」按系统要求处理）。  
   - 可选：提示用户靠近设备、设备处于配网模式。

2. **扫描 BLE 设备**  
   - 开始扫描后，持续合并结果：同一 `deviceId` 只保留一条；`name` 为空时展示占位名（如「未命名设备」）。  
   - 提供「停止扫描」或超时自动停止（原小程序约 60 秒停扫，产品可配置）。

3. **用户选择设备 → 建立 BLE 连接**  
   - 对选中项发起 GATT 连接。  
   - 连接成功后：  
     - 记录 `deviceId`。  
     - **将报文序号 `sequence` 重置为 -1**（见第 5 节：第一次发帧前会变成 0）。  
     - 延迟约 **1 秒** 再进入「选 Wi‑Fi」界面（给连接与协议栈稳定时间，可配置）。

4. **发现 GATT：服务 + 特征**  
   - 枚举服务，取 **主服务（isPrimary == true）** 的 `uuid` 作为后续读写上下文。  
   - 在该服务下枚举特征：  
     - **写特征**：`properties.write == true` 中选取 **第一个**（若硬件有多个可写特征，应改为硬件文档指定的 UUID）。  
     - **通知特征**：`properties.notify == true` 中选取 **第一个**（同上，理想情况是写死 UUID）。  
   - 对通知特征：`setNotifyValue(true)`，并订阅其数据流；收到的数据交给 **回包解析**（第 7 节）。

5. **选择 Wi‑Fi + 输入密码**  
   - 展示系统扫描到的 SSID 列表（可排序：信号强度降序）。  
   - 用户必选一条 SSID，并输入密码（可为空仅当路由器开放网络——产品自行决定是否允许空密码）。  
   - 点击「开始配网」后执行 **下发序列**（第 6 节），期间可置按钮为 loading，禁止重复点击。

6. **下发序列（严格顺序、失败即停并提示）**  
   顺序固定为：  
   **发送 SSID 帧 → 成功后再发密码帧 → 成功后再发 CONNECT_WIFI → 可选延迟后发「查状态」帧。**  
   任一步 `write` 失败：提示失败，不自动执行后续步（除非产品定义重试策略）。

7. **结束与断开**  
   - 成功：根据 notify 解析结果（第 7 节）提示「已连上 / 密码错误 / 超时」等。  
   - 用户离开页面：停止扫描、取消 notify、`disconnect`、释放监听。

---

## 3. Android 权限与系统行为（业务要求）

- **蓝牙**：扫描、连接（Android 12+ 区分 `BLUETOOTH_SCAN` / `BLUETOOTH_CONNECT`）。  
- **定位**：扫描 BLE 与读取 Wi‑Fi 扫描结果时，多数机型仍要求 **精确/粗略定位权限** 且 **系统定位开关打开**。  
- **Wi‑Fi**：使用系统 API 触发扫描并读取结果列表；列表可能不含隐藏 SSID，需产品说明或允许 **手动输入 SSID**（推荐作为兜底）。

---

## 4. Wi‑Fi 列表业务规则

- 列表数据字段至少包含：`SSID`、`BSSID`（作列表 key）、`signalStrength` 或 level（用于排序与展示）。  
- 过滤：`SSID` 为空的条目不展示。  
- 排序：信号强到弱。  
- 监听扫描结果：**避免重复注册多个监听**导致内存/回调堆积。

---

## 5. 报文序号 `sequence`

- 在 **BLE 连接成功** 那一刻将内部计数器设为 **-1**。  
- **每构造并发送一帧**前，执行 **先自增再使用**：第一次发送时序号为 **0**，之后每帧 +1。  
- 序号占 **1 字节**（uint8），溢出后按 0–255 循环（与硬件约定一致即可）。

---

## 6. 下行帧格式（手机 → 设备）

所有下行数据均写入 **「写特征」**，结构为：

| 偏移 | 长度 | 含义 |
|------|------|------|
| 0 | 1 | `type` 字节，见下 |
| 1 | 1 | `frameCtrl` 固定 **0x04**（DIRECTION_OUTPUT，bit2=1） |
| 2 | 1 | `sequence`，本帧使用的序号（自增后的值） |
| 3 | 1 | `len`，payload 长度；无 payload 时为 **0** |
| 4 | `len` | payload；无 payload 时本帧总长为 **4** |

**`type` 字节计算**：

- 低 2 位：`pkgType`（0 = Ctrl，1 = Data）。  
- 高 6 位：`subtype & 0x3F`。  
- 公式：`type = ((subtype & 0x3F) << 2) | (pkgType & 0x03)`。

### 6.1 SSID（Data）

- `pkgType = 1`（Data）  
- `subtype = 0x02`  
- `payload` = SSID 的 **UTF-8 字节序列**。

### 6.2 Password（Data）

- `pkgType = 1`  
- `subtype = 0x03`  
- `payload` = 密码的 **UTF-8 字节序列**。

### 6.3 CONNECT_WIFI（Ctrl）

- `pkgType = 0`（Ctrl）  
- `subtype = 0x03`  
- **无 payload**，`len = 0`，总长 4 字节。  
- 业务含义：命令设备用已下发的 SSID/密码去连接路由器。  
- UI：提示「配网指令已发送，请等待…」；可在 **约 3 秒** 后自动发「查状态」（可调）。

### 6.4 查询 Wi‑Fi 状态（Ctrl）

- `pkgType = 0`  
- `subtype = 0x05`  
- **无 payload**，`len = 0`。  
- 业务含义：向设备查询当前 Wi‑Fi 连接相关状态（具体语义由固件定义）。

### 6.5 分包与加密

- 原小程序实现 **未体现** BluFi 标准里的加密协商、分片、ACK 等复杂步骤；若硬件同样只实现上述简化子集，Flutter **照搬即可**。  
- 若硬件文档要求 **安全会话、分片、CRC、ACK**，则须在协议层按硬件文档扩展。

---

## 7. 上行数据（设备 → 手机，Notify）

- 数据来源：**notify 特征** 的 value 流。  
- 产品化时应实现：  
  - 识别帧头/类型/序号/长度（若与下行同构，则对称解析）。  
  - 按 subtype 区分 **连接成功、密码错误、找不到 SSID、DHCP 成功/失败** 等，**具体枚举以硬件协议为准**。  
- 业务建议：  
  - 为「查状态」设置 **超时**（如 5–10 秒），无回包则提示「设备无响应，请重试」。  
  - CONNECT 之后不要立刻认为成功，应以 notify 或二次查状态为准。

---

## 8. 错误与重试（建议业务规则）

- 蓝牙未开、权限拒绝：阻断流程并引导去设置页。  
- 扫描不到目标设备：提示确认设备配网模式、距离、是否被其它手机占用连接。  
- GATT 找不到主服务或写/notify 特征：提示固件版本或设备型号不匹配。  
- **写特征失败**：记录错误码；可选 **同一帧重试 1–3 次**（间隔 100–300 ms），仍失败则中止序列。  
- **连接断开**：中止配网序列，回到设备列表或提示重连。

---

## 9. 与硬件联调的合同清单（交付前必须对齐）

1. 主服务 UUID、写特征 UUID、notify 特征 UUID（若不能「取第一个可写/第一个 notify」，必须写死）。  
2. 下行帧是否 **完全** 按第 6 节（尤其 `frameCtrl=0x04`、`type` 合成方式）。  
3. SSID/密码是否 **必须为 UTF-8**；最大长度；是否支持 5G / 隐藏 SSID。  
4. notify 回包格式及每种状态的 subtype / payload。  
5. 是否需要先握手/版本号/厂商自定义命令再发 SSID（原小程序无此步）。

---

## 10. Flutter 工程内建议的职责划分

1. **BleProvisioningRepository**：扫描、连接、发现服务特征、写字节、notify 流。  
2. **BlufiFrameCodec**（命名可改）：`buildFrame(pkgType, subtype, payload?)`、`nextSequence()`、字符串转 UTF-8 payload。  
3. **ProvisioningSession**：编排 `sendSsid → sendPassword → connectWifi → getStatus`，持有超时与取消。  
4. **WifiScanRepository**：Android 扫描列表 + 兜底手动输入 SSID。  
5. **UI**：设备列表、配网弹窗、结果页。

---

## 附录 A：配网流程文字总览（无代码）

1. 手机蓝牙可用 + 权限就绪。  
2. BLE 扫描发现设备列表，用户选一个连接。  
3. 连接成功后枚举 GATT：主服务 + 写特征 + notify 特征，打开 notify。  
4. 用系统能力列出周边 Wi‑Fi，用户选 SSID 并输入密码。  
5. 按固定顺序通过 **写特征** 发送四步二进制帧：SSID → 密码 → CONNECT_WIFI →（可选）查状态。  
6. 设备固件在收到数据后自行连路由器；手机根据 notify（及超时）给用户结果反馈。

---

## 附录 B：与原小程序实现的对应关系（便于对照）

| 能力 | 原小程序侧 |
|------|------------|
| BLE 初始化/扫描/连接 | `uni.openBluetoothAdapter`、`startBluetoothDevicesDiscovery`、`createBLEConnection` |
| GATT | `getBLEDeviceServices`、`getBLEDeviceCharacteristics` |
| 通知 | `notifyBLECharacteristicValueChange`、`onBLECharacteristicValueChange` |
| 下发 | `writeBLECharacteristicValue` |
| Wi‑Fi 列表 | 微信 `wx.startWifi` / `getWifiList`；Flutter 改用 Android Wi‑Fi API |
| 帧构造 | `buildBlufiFrame`、`stringToUint8Array`（UTF-8） |

原实现中 **notify 回包解析未完成**，Flutter 独立交付时应按第 7、9 节补全。

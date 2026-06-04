import BlockingQueue from './utils/BlockingQueue.js';
import { log } from './utils/logger.js';

// 音频流播放上下文类
export class StreamingContext {
    constructor(opusDecoder, audioContext, sampleRate, channels, minAudioDuration) {
        this.opusDecoder = opusDecoder;
        this.audioContext = audioContext;

        // 音频参数
        this.sampleRate = sampleRate;
        this.channels = channels;
        this.minAudioDuration = minAudioDuration;

        // 初始化队列和状态
        this.queue = [];          // 已解码的PCM队列。正在播放
        this.activeQueue = new BlockingQueue(); // 已解码的PCM队列。准备播放
        this.pendingAudioBufferQueue = [];  // 待处理的缓存队列
        this.audioBufferQueue = new BlockingQueue();  // 缓存队列
        this.playing = false;     // 是否正在播放
        this.endOfStream = false; // 是否收到结束信号
        this.source = null;       // 当前音频源
        this.totalSamples = 0;    // 累积的总样本数
        this.lastPlayTime = 0;    // 上次播放的时间戳
        this.nextPlayTime = 0;    // WebAudio 时间轴上的下一段播放时间
        this.playLoopStarted = false;
        this.decodeLoopStarted = false;
    }

    // 缓存音频数组
    pushAudioBuffer(item) {
        this.audioBufferQueue.enqueue(...item);
    }

    // 获取需要处理缓存队列，单线程：在audioBufferQueue一直更新的状态下不会出现安全问题
    async getPendingAudioBufferQueue() {
        // 原子交换 + 清空
        [this.pendingAudioBufferQueue, this.audioBufferQueue] = [await this.audioBufferQueue.dequeue(), new BlockingQueue()];
    }

    // 获取正在播放已解码的PCM队列，单线程：在activeQueue一直更新的状态下不会出现安全问题
    async getQueue(minSamples) {
        let TepArray = [];
        const num = minSamples - this.queue.length > 0 ? minSamples - this.queue.length : 1;
        // 原子交换 + 清空
        [TepArray, this.activeQueue] = [await this.activeQueue.dequeue(num), new BlockingQueue()];
        this.queue.push(...TepArray);
    }

    // 将Int16音频数据转换为Float32音频数据
    convertInt16ToFloat32(int16Data) {
        const float32Data = new Float32Array(int16Data.length);
        for (let i = 0; i < int16Data.length; i++) {
            // 将[-32768,32767]范围转换为[-1,1]
            float32Data[i] = int16Data[i] / (int16Data[i] < 0 ? 0x8000 : 0x7FFF);
        }
        return float32Data;
    }

    // 将Opus数据解码为PCM
    async decodeOpusFrames() {
        if (this.decodeLoopStarted) {
            return;
        }
        this.decodeLoopStarted = true;
        if (!this.opusDecoder) {
            log('Opus解码器未初始化，无法解码', 'error');
            return;
        } else {
            log('Opus解码器启动', 'info');
        }

        while (true) {
            const frames = await this.audioBufferQueue.dequeue(1);
            for (const frame of frames) {
                try {
                    // 使用Opus解码器解码
                    const frameData = this.opusDecoder.decode(frame);
                    if (frameData && frameData.length > 0) {
                        // 转换为Float32
                        const floatData = this.convertInt16ToFloat32(frameData);
                        this.activeQueue.enqueue(floatData);
                        this.totalSamples += floatData.length;
                    }
                } catch (error) {
                    log("Opus解码失败: " + error.message, 'error');
                }
            }
        }
    }

    // 开始播放音频
    async startPlaying() {
        if (this.playLoopStarted) {
            return;
        }
        this.playLoopStarted = true;
        while (true) {
            const chunks = await this.activeQueue.dequeue(1);
            for (const chunk of chunks) {
                this.scheduleChunk(chunk);
            }
        }
    }

    scheduleChunk(chunk) {
        if (!chunk || chunk.length === 0) return;

        const audioBuffer = this.audioContext.createBuffer(this.channels, chunk.length, this.sampleRate);
        audioBuffer.copyToChannel(chunk, 0);

        const source = this.audioContext.createBufferSource();
        source.buffer = audioBuffer;

        const gainNode = this.audioContext.createGain();
        const duration = audioBuffer.duration;
        const now = this.audioContext.currentTime;
        const startAt = Math.max(now + 0.02, this.nextPlayTime || now + 0.02);

        // 对非常短的 Opus 帧只做极短淡入淡出，避免每帧 20ms fade 吃掉主体声音。
        const fadeDuration = Math.min(0.005, duration / 4);
        gainNode.gain.setValueAtTime(0, startAt);
        gainNode.gain.linearRampToValueAtTime(1, startAt + fadeDuration);
        if (duration > fadeDuration * 2) {
            gainNode.gain.setValueAtTime(1, startAt + duration - fadeDuration);
            gainNode.gain.linearRampToValueAtTime(0, startAt + duration);
        }

        source.connect(gainNode);
        gainNode.connect(this.audioContext.destination);
        source.start(startAt);

        this.lastPlayTime = startAt;
        this.nextPlayTime = startAt + duration;
        log(`排程播放 ${chunk.length} 个样本，约 ${(duration).toFixed(3)} 秒，startAt=${startAt.toFixed(3)}`, 'debug');
    }
}

// 创建streamingContext实例的工厂函数
export function createStreamingContext(opusDecoder, audioContext, sampleRate, channels, minAudioDuration) {
    return new StreamingContext(opusDecoder, audioContext, sampleRate, channels, minAudioDuration);
}
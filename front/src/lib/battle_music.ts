const DEFAULT_CROSSFADE_MS = 4200;
const DEFAULT_STOP_FADE_MS = 420;

export const BATTLE_MUSIC_TRACKS = [
  "/audio/battle/track_1.mp3",
  "/audio/battle/track_2.mp3",
  "/audio/battle/track_3.mp3",
  "/audio/battle/track_4.mp3",
  "/audio/battle/track_5.mp3",
  "/audio/battle/track_6.mp3",
  "/audio/battle/track_7.mp3",
  "/audio/battle/track_8.mp3",
  "/audio/battle/track_9.mp3",
  "/audio/battle/track_10.mp3",
];

type BattleMusicOptions = {
  volume: number;
  crossfadeMs?: number;
};

type FadeState = {
  frameId: number;
};

function shuffleTracks(tracks: string[]) {
  const shuffled = [...tracks];
  for (let index = shuffled.length - 1; index > 0; index -= 1) {
    const swapIndex = Math.floor(Math.random() * (index + 1));
    [shuffled[index], shuffled[swapIndex]] = [shuffled[swapIndex], shuffled[index]];
  }
  return shuffled;
}

function easeOutCubic(progress: number) {
  return 1 - Math.pow(1 - progress, 3);
}

export class BattleMusicManager {
  private readonly tracks: string[];
  private readonly channels: HTMLAudioElement[];
  private readonly fadeStates = new Map<HTMLAudioElement, FadeState>();
  private readonly crossfadeMs: number;
  private targetVolume: number;
  private bag: string[] = [];
  private activeChannelIndex = 0;
  private lastTrack = "";
  private monitorFrameId: number | null = null;
  private transitionInProgress = false;
  private stopped = true;
  private unlocked = false;

  constructor(tracks: string[], options: BattleMusicOptions) {
    this.tracks = tracks;
    this.targetVolume = options.volume;
    this.crossfadeMs = options.crossfadeMs ?? DEFAULT_CROSSFADE_MS;
    this.channels = [this.createChannel(), this.createChannel()];
  }

  setVolume(volume: number) {
    this.targetVolume = volume;
    this.channels[this.activeChannelIndex].volume = Math.min(this.channels[this.activeChannelIndex].volume, volume);
  }

  async unlock() {
    if (this.unlocked || this.tracks.length === 0) {
      return;
    }

    const channel = this.channels[0];
    const previousSrc = channel.currentSrc || channel.src;
    const previousTime = channel.currentTime;
    const previousVolume = channel.volume;

    channel.src = previousSrc || this.tracks[0];
    channel.volume = 0;

    try {
      await channel.play();
      channel.pause();
      channel.currentTime = previousSrc ? previousTime : 0;
      channel.volume = previousVolume;
      this.unlocked = true;
    } catch {
      channel.volume = previousVolume;
    }
  }

  start() {
    if (this.tracks.length === 0) {
      return;
    }

    this.stopped = false;
    const activeChannel = this.channels[this.activeChannelIndex];

    if (!activeChannel.src || activeChannel.ended) {
      this.loadTrack(activeChannel, this.takeNextTrack());
    }

    if (activeChannel.paused) {
      activeChannel.volume = 0;
      void activeChannel.play().catch(() => undefined);
    }

    this.fadeTo(activeChannel, this.targetVolume, this.crossfadeMs);
    this.startMonitor();
  }

  stop(durationMs = DEFAULT_STOP_FADE_MS) {
    this.stopped = true;
    this.stopMonitor();
    this.transitionInProgress = false;

    this.channels.forEach((channel) => {
      this.fadeTo(channel, 0, durationMs, () => {
        if (!this.stopped) {
          return;
        }

        channel.pause();
        channel.currentTime = 0;
      });
    });
  }

  destroy() {
    this.stop(0);
    this.fadeStates.forEach((fadeState) => window.cancelAnimationFrame(fadeState.frameId));
    this.fadeStates.clear();
    this.channels.forEach((channel) => {
      channel.src = "";
      channel.load();
    });
  }

  private createChannel() {
    const channel = new Audio();
    channel.preload = "auto";
    channel.loop = false;
    channel.volume = 0;
    channel.addEventListener("ended", () => {
      if (this.channels[this.activeChannelIndex] === channel) {
        this.startNextTrack(1200);
      }
    });
    return channel;
  }

  private loadTrack(channel: HTMLAudioElement, track: string) {
    channel.src = track;
    channel.currentTime = 0;
    channel.load();
  }

  private takeNextTrack() {
    if (this.bag.length === 0) {
      this.bag = shuffleTracks(this.tracks);
      if (this.tracks.length > 1 && this.bag[0] === this.lastTrack) {
        const swapIndex = this.bag.findIndex((track) => track !== this.lastTrack);
        if (swapIndex > 0) {
          [this.bag[0], this.bag[swapIndex]] = [this.bag[swapIndex], this.bag[0]];
        }
      }
    }

    const track = this.bag.shift() ?? this.tracks[0];
    this.lastTrack = track;
    return track;
  }

  private startMonitor() {
    this.stopMonitor();

    const tick = () => {
      if (this.stopped) {
        this.monitorFrameId = null;
        return;
      }

      const activeChannel = this.channels[this.activeChannelIndex];
      const duration = activeChannel.duration;
      if (
        !this.transitionInProgress &&
        Number.isFinite(duration) &&
        duration > 0 &&
        duration - activeChannel.currentTime <= this.crossfadeMs / 1000
      ) {
        this.startNextTrack(this.crossfadeMs);
      }

      this.monitorFrameId = window.requestAnimationFrame(tick);
    };

    this.monitorFrameId = window.requestAnimationFrame(tick);
  }

  private stopMonitor() {
    if (this.monitorFrameId === null) {
      return;
    }

    window.cancelAnimationFrame(this.monitorFrameId);
    this.monitorFrameId = null;
  }

  private startNextTrack(fadeMs: number) {
    if (this.transitionInProgress || this.stopped || this.tracks.length === 0) {
      return;
    }

    this.transitionInProgress = true;
    const previousChannel = this.channels[this.activeChannelIndex];
    const nextChannelIndex = this.activeChannelIndex === 0 ? 1 : 0;
    const nextChannel = this.channels[nextChannelIndex];

    this.loadTrack(nextChannel, this.takeNextTrack());
    nextChannel.volume = 0;
    void nextChannel.play().catch(() => undefined);

    this.fadeTo(previousChannel, 0, fadeMs, () => {
      previousChannel.pause();
      previousChannel.currentTime = 0;
    });
    this.fadeTo(nextChannel, this.targetVolume, fadeMs, () => {
      this.transitionInProgress = false;
    });
    this.activeChannelIndex = nextChannelIndex;
  }

  private fadeTo(audio: HTMLAudioElement, targetVolume: number, durationMs: number, onDone?: () => void) {
    const currentFade = this.fadeStates.get(audio);
    if (currentFade) {
      window.cancelAnimationFrame(currentFade.frameId);
      this.fadeStates.delete(audio);
    }

    const startVolume = audio.volume;
    if (durationMs <= 0 || Math.abs(startVolume - targetVolume) < 0.005) {
      audio.volume = targetVolume;
      onDone?.();
      return;
    }

    const startedAt = performance.now();
    const tick = (now: number) => {
      const progress = Math.min((now - startedAt) / durationMs, 1);
      const eased = easeOutCubic(progress);
      audio.volume = startVolume + (targetVolume - startVolume) * eased;

      if (progress < 1) {
        const frameId = window.requestAnimationFrame(tick);
        this.fadeStates.set(audio, { frameId });
        return;
      }

      audio.volume = targetVolume;
      this.fadeStates.delete(audio);
      onDone?.();
    };

    const frameId = window.requestAnimationFrame(tick);
    this.fadeStates.set(audio, { frameId });
  }
}

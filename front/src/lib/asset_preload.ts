const warmedAssetUrls = new Set<string>();
const inFlightAssetUrls = new Map<string, Promise<void>>();

type WarmAssetOptions = {
  concurrency?: number;
  signal?: AbortSignal;
  onProgress?: (loadedCount: number) => void;
  batchDelayMs?: number;
};

type BackgroundWarmOptions = WarmAssetOptions & {
  initialDelayMs?: number;
};

export function uniqueAssetUrls(urls: string[]) {
  return Array.from(new Set(urls.filter(Boolean)));
}

function wait(ms: number, signal?: AbortSignal) {
  if (ms <= 0 || signal?.aborted) {
    return Promise.resolve();
  }

  return new Promise<void>((resolve) => {
    const timeoutId = window.setTimeout(resolve, ms);
    signal?.addEventListener(
      "abort",
      () => {
        window.clearTimeout(timeoutId);
        resolve();
      },
      { once: true },
    );
  });
}

function preloadImage(url: string): Promise<void> {
  if (warmedAssetUrls.has(url)) {
    return Promise.resolve();
  }

  const inFlight = inFlightAssetUrls.get(url);
  if (inFlight) {
    return inFlight;
  }

  const promise = new Promise<void>((resolve) => {
    const image = new Image();
    let settled = false;

    const finish = () => {
      if (settled) {
        return;
      }
      settled = true;
      warmedAssetUrls.add(url);
      inFlightAssetUrls.delete(url);
      resolve();
    };

    image.onload = finish;
    image.onerror = finish;
    image.decoding = "async";
    image.src = url;

    if (image.complete) {
      finish();
    }
  });

  inFlightAssetUrls.set(url, promise);
  return promise;
}

function preloadAudio(url: string): Promise<void> {
  if (warmedAssetUrls.has(url)) {
    return Promise.resolve();
  }

  const inFlight = inFlightAssetUrls.get(url);
  if (inFlight) {
    return inFlight;
  }

  const promise = new Promise<void>((resolve) => {
    const audio = new Audio(url);
    let settled = false;

    const cleanup = () => {
      audio.removeEventListener("canplaythrough", finish);
      audio.removeEventListener("loadeddata", finish);
      audio.removeEventListener("error", finish);
    };

    const finish = () => {
      if (settled) {
        return;
      }
      settled = true;
      cleanup();
      warmedAssetUrls.add(url);
      inFlightAssetUrls.delete(url);
      resolve();
    };

    audio.preload = "auto";
    audio.addEventListener("canplaythrough", finish);
    audio.addEventListener("loadeddata", finish);
    audio.addEventListener("error", finish);
    audio.load();
  });

  inFlightAssetUrls.set(url, promise);
  return promise;
}

export function preloadAsset(url: string): Promise<void> {
  return /\.(mp3|wav|ogg)(?:$|\?)/i.test(url) ? preloadAudio(url) : preloadImage(url);
}

export async function warmAssetUrls(urls: string[], options: WarmAssetOptions = {}) {
  const uniqueUrls = uniqueAssetUrls(urls);
  const concurrency = Math.max(1, options.concurrency ?? 4);
  const batchDelayMs = Math.max(0, options.batchDelayMs ?? 0);
  let loadedCount = 0;
  let nextIndex = 0;

  options.onProgress?.(0);

  async function worker() {
    while (nextIndex < uniqueUrls.length && !options.signal?.aborted) {
      const index = nextIndex;
      nextIndex += 1;
      await preloadAsset(uniqueUrls[index]);
      loadedCount += 1;
      options.onProgress?.(loadedCount);
      await wait(batchDelayMs, options.signal);
    }
  }

  await Promise.all(Array.from({ length: Math.min(concurrency, uniqueUrls.length) }, () => worker()));
}

export function warmAssetUrlsInBackground(urls: string[], options: BackgroundWarmOptions = {}) {
  const controller = new AbortController();
  const initialDelayMs = Math.max(0, options.initialDelayMs ?? 500);

  const timeoutId = window.setTimeout(() => {
    void warmAssetUrls(urls, {
      ...options,
      signal: controller.signal,
    }).catch(() => undefined);
  }, initialDelayMs);

  return () => {
    window.clearTimeout(timeoutId);
    controller.abort();
  };
}

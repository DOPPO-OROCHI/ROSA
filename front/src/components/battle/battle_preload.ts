import { useEffect, useMemo, useRef, useState } from "react";
import { resolveCardAssetVariantSrc, resolveHeroAssetVariantSrc } from "../../lib/api";
import type { MaskedBattleMatchState } from "./types";

const warmedAssetUrls = new Set<string>();
const inFlightAssetUrls = new Map<string, Promise<void>>();

type BattlePreloadState = {
  visible: boolean;
  progress: number;
  actualProgress: number;
  loadedCount: number;
  totalCount: number;
  label: string;
};

function unique<T>(items: T[]) {
  return Array.from(new Set(items));
}

function collectBattleAssetUrls(match: MaskedBattleMatchState): string[] {
  const urls: string[] = [];

  match.players.forEach((player) => {
    if (!player) {
      return;
    }

    if (player.hero_code) {
      urls.push(resolveHeroAssetVariantSrc(player.hero_code, "battle_icon"));
    }

    player.table.forEach((unit) => {
      if (!unit?.template_id) {
        return;
      }

      urls.push(resolveCardAssetVariantSrc("battle", unit.template_id, "view"));
      urls.push(resolveCardAssetVariantSrc("battle", unit.template_id, "on_table"));
      urls.push(resolveCardAssetVariantSrc("battle", unit.template_id, "full_art"));
    });

    [player.hand ?? [], player.deck ?? [], player.discard ?? []].forEach((zone) => {
      zone.forEach((card) => {
        if (!card.template_id || (card.kind !== "battle" && card.kind !== "buff")) {
          return;
        }

        urls.push(resolveCardAssetVariantSrc(card.kind, card.template_id, "view"));
        urls.push(resolveCardAssetVariantSrc(card.kind, card.template_id, "on_table"));
        urls.push(resolveCardAssetVariantSrc(card.kind, card.template_id, "full_art"));
      });
    });
  });

  return unique(urls);
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

function warmUrls(urls: string[], onProgress: (loadedCount: number) => void) {
  if (urls.length === 0) {
    onProgress(0);
    return Promise.resolve();
  }

  let loadedCount = 0;
  onProgress(0);

  return Promise.all(
    urls.map((url) =>
      preloadImage(url).then(() => {
        loadedCount += 1;
        onProgress(loadedCount);
      }),
    ),
  ).then(() => undefined);
}

export function useBattlePreload(match: MaskedBattleMatchState | null): BattlePreloadState {
  const [visible, setVisible] = useState(true);
  const [progress, setProgress] = useState(0);
  const [actualProgress, setActualProgress] = useState(0);
  const [loadedCount, setLoadedCount] = useState(0);
  const [totalCount, setTotalCount] = useState(0);
  const startedMatchIdRef = useRef<number | null>(null);
  const backgroundWarmedUrlsRef = useRef<Set<string>>(new Set());
  const actualProgressRef = useRef(0);
  const currentMatchId = match?.match_id ?? null;

  const urls = useMemo(() => (match ? collectBattleAssetUrls(match) : []), [match]);
  const urlsRef = useRef<string[]>([]);

  useEffect(() => {
    urlsRef.current = urls;
  }, [urls]);

  useEffect(() => {
    if (!match) {
      return;
    }

    const freshUrls = urls.filter((url) => !backgroundWarmedUrlsRef.current.has(url));
    if (freshUrls.length === 0) {
      return;
    }

    freshUrls.forEach((url) => {
      backgroundWarmedUrlsRef.current.add(url);
    });

    void Promise.all(freshUrls.map((url) => preloadImage(url)));
  }, [match, urls]);

  useEffect(() => {
    actualProgressRef.current = actualProgress;
  }, [actualProgress]);

  useEffect(() => {
    if (!currentMatchId) {
      return;
    }

    if (startedMatchIdRef.current === currentMatchId) {
      return;
    }

    startedMatchIdRef.current = currentMatchId;
    setVisible(true);
    setProgress(0);
    setActualProgress(0);
    setLoadedCount(0);
    setTotalCount(urlsRef.current.length);

    const startTime = Date.now();
    const minimumDurationMs = 5000;
    let disposed = false;

    const progressInterval = window.setInterval(() => {
      if (disposed) {
        return;
      }
      const elapsed = Date.now() - startTime;
      const timeProgress = Math.min(elapsed / minimumDurationMs, 1);
      setProgress((current) => Math.max(current, actualProgressRef.current, timeProgress));
    }, 60);

    void warmUrls(urlsRef.current, (nextLoadedCount) => {
      if (disposed) {
        return;
      }

      const total = urlsRef.current.length;
      const nextActualProgress = total > 0 ? nextLoadedCount / total : 1;
      setLoadedCount(nextLoadedCount);
      setActualProgress(nextActualProgress);
      setProgress((current) => Math.max(current, nextActualProgress));
    });

    const finishTimeout = window.setTimeout(() => {
      if (disposed) {
        return;
      }
      setProgress(1);
      setVisible(false);
    }, minimumDurationMs);

    return () => {
      disposed = true;
      window.clearInterval(progressInterval);
      window.clearTimeout(finishTimeout);
    };
  }, [currentMatchId]);

  const label =
    totalCount > 0
      ? `Прогреваем ассеты матча ${Math.min(loadedCount, totalCount)}/${totalCount}`
      : "Подготавливаем поле битвы";

  return {
    visible,
    progress,
    actualProgress,
    loadedCount,
    totalCount,
    label,
  };
}

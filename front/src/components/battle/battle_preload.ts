import { useEffect, useMemo, useRef, useState } from "react";
import { resolveCardAssetVariantSrc, resolveHeroAssetVariantSrc } from "../../lib/api";
import { preloadAsset, uniqueAssetUrls, warmAssetUrls } from "../../lib/asset_preload";
import type { MaskedBattleMatchState } from "./types";
import type { DeckEntry } from "../../types";

const STATIC_BATTLE_ASSET_URLS = [
  "/assets/ui/pictures/boards/battle/image.png",
  "/assets/ui/pictures/backgrounds/battle/image.png",
  "/assets/ui/sounds/ui/click.mp3",
  "/assets/ui/sounds/music/battle.mp3",
  "/assets/ui/sounds/combat/impact.mp3",
];

type BattlePreloadState = {
  visible: boolean;
  completed: boolean;
  progress: number;
  actualProgress: number;
  loadedCount: number;
  totalCount: number;
  label: string;
};

function addCardAssetUrls(urls: string[], kind: "battle" | "buff", templateId: string) {
  urls.push(resolveCardAssetVariantSrc(kind, templateId, "view"));
  urls.push(resolveCardAssetVariantSrc(kind, templateId, "full_art"));

  if (kind === "battle") {
    urls.push(resolveCardAssetVariantSrc(kind, templateId, "on_table"));
    urls.push(`/assets/cards/battle/${templateId}/sfx/summon/sound.mp3`);
    urls.push(`/assets/cards/battle/${templateId}/sfx/death/sound.mp3`);
    urls.push(`/assets/cards/battle/${templateId}/sfx/spell/sound.mp3`);
  }
}

function collectBattleAssetUrls(match: MaskedBattleMatchState, deckEntries: DeckEntry[]): string[] {
  const urls: string[] = [...STATIC_BATTLE_ASSET_URLS];

  deckEntries.forEach((entry) => {
    if (entry.count <= 0) {
      return;
    }
    addCardAssetUrls(urls, entry.kind, entry.template_id);
  });

  match.players.forEach((player) => {
    if (!player) {
      return;
    }

    if (player.hero_code) {
      urls.push(resolveHeroAssetVariantSrc(player.hero_code, "view"));
      urls.push(resolveHeroAssetVariantSrc(player.hero_code, "full_art"));
      urls.push(resolveHeroAssetVariantSrc(player.hero_code, "battle_icon"));
    }

    player.table.forEach((unit) => {
      if (!unit?.template_id) {
        return;
      }

      addCardAssetUrls(urls, "battle", unit.template_id);
    });

    [player.hand ?? [], player.deck ?? [], player.discard ?? [], player.graveyard ?? []].forEach((zone) => {
      zone.forEach((card) => {
        if (!card.template_id || (card.kind !== "battle" && card.kind !== "buff")) {
          return;
        }

        addCardAssetUrls(urls, card.kind, card.template_id);
      });
    });
  });

  return uniqueAssetUrls(urls);
}

export function useBattlePreload(match: MaskedBattleMatchState | null, deckEntries: DeckEntry[] = []): BattlePreloadState {
  const [visible, setVisible] = useState(true);
  const [completed, setCompleted] = useState(false);
  const [progress, setProgress] = useState(0);
  const [actualProgress, setActualProgress] = useState(0);
  const [loadedCount, setLoadedCount] = useState(0);
  const [totalCount, setTotalCount] = useState(0);
  const startedMatchIdRef = useRef<number | null>(null);
  const backgroundWarmedUrlsRef = useRef<Set<string>>(new Set());
  const currentMatchId = match?.match_id ?? null;

  const urls = useMemo(() => (match ? collectBattleAssetUrls(match, deckEntries) : []), [deckEntries, match]);
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

    void warmAssetUrls(freshUrls, { concurrency: 2, batchDelayMs: 60 });
  }, [match, urls]);

  useEffect(() => {
    if (!currentMatchId) {
      return;
    }

    if (startedMatchIdRef.current === currentMatchId) {
      return;
    }

    startedMatchIdRef.current = currentMatchId;
    setVisible(true);
    setCompleted(false);
    setProgress(0);
    setActualProgress(0);
    setLoadedCount(0);
    setTotalCount(urlsRef.current.length);

    let disposed = false;

    void warmAssetUrls(urlsRef.current, {
      concurrency: 4,
      onProgress: (nextLoadedCount) => {
        if (disposed) {
          return;
        }

        const total = urlsRef.current.length;
        const nextActualProgress = total > 0 ? nextLoadedCount / total : 1;
        setLoadedCount(nextLoadedCount);
        setActualProgress(nextActualProgress);
        setProgress(nextActualProgress);
      },
    }).then(() => {
      if (disposed) {
        return;
      }
      setProgress(1);
      setCompleted(true);
      setVisible(false);
    });

    return () => {
      disposed = true;
    };
  }, [currentMatchId]);

  const label =
    totalCount > 0
      ? `Прогреваем ассеты матча ${Math.min(loadedCount, totalCount)}/${totalCount}`
      : "Подготавливаем поле битвы";

  return {
    visible,
    completed,
    progress,
    actualProgress,
    loadedCount,
    totalCount,
    label,
  };
}

import { useEffect, useRef } from "react";
import type { MaskedBattleMatchState, BattleEvent, BattleUnitState } from "./types";

export type BattleCardSfxKind = "summon" | "death" | "spell";

export const COMMON_ATTACK_HIT_SFX_SRC = "/assets/ui/sounds/combat/impact.mp3";

const audioCache = new Map<string, HTMLAudioElement>();
const missingAudioSources = new Set<string>();

function getEventSequenceId(event: BattleEvent, fallback: string) {
  return event.id ?? event.event_id ?? fallback;
}

function findUnit(match: MaskedBattleMatchState, instanceId: string): BattleUnitState | null {
  for (const player of match.players) {
    for (const unit of player?.table ?? []) {
      if (unit?.instance_id === instanceId) {
        return unit;
      }
    }
  }
  return null;
}

function resolveEventTemplateId(match: MaskedBattleMatchState, event: BattleEvent): string {
  return event.source_card_template_id
    ?? event.source_template_id
    ?? (event.source_instance_id ? findUnit(match, event.source_instance_id)?.template_id : "")
    ?? "";
}

function createAudioSource(src: string) {
  const audio = new Audio(src);
  audio.preload = "auto";
  audio.addEventListener("error", () => {
    missingAudioSources.add(src);
    audioCache.delete(src);
  });
  return audio;
}

export function playAudioSource(src: string, volume = 1) {
  if (!src || missingAudioSources.has(src)) {
    return;
  }

  let audio = audioCache.get(src);
  if (!audio) {
    audio = createAudioSource(src);
    audioCache.set(src, audio);
  }

  const instance = audio.cloneNode() as HTMLAudioElement;
  instance.volume = volume;
  instance.currentTime = 0;
  instance.addEventListener("error", () => {
    missingAudioSources.add(src);
    audioCache.delete(src);
  }, { once: true });
  void instance.play().catch(() => undefined);
}

export function resolveBattleCardSfxSrc(templateId: string, kind: BattleCardSfxKind) {
  if (!templateId) {
    return "";
  }
  return `/assets/cards/battle/${templateId}/sfx/${kind}/sound.mp3`;
}

export function playBattleCardSfx(templateId: string, kind: BattleCardSfxKind, volume = 1) {
  const src = resolveBattleCardSfxSrc(templateId, kind);
  if (!src) {
    return;
  }
  playAudioSource(src, volume);
}

export function useBattleEventSfx(match: MaskedBattleMatchState | null) {
  const processedEventIdsRef = useRef(new Set<string>());

  useEffect(() => {
    if (!match?.events?.length) {
      return;
    }

    match.events.forEach((event, eventIndex) => {
      const targetFingerprint = (event.targets ?? [])
        .map((target) => `${target.instance_id ?? "none"}:${target.amount ?? 0}:${target.new_hp ?? "na"}:${target.died ? 1 : 0}`)
        .join("|");
      const eventId = getEventSequenceId(
        event,
        `${eventIndex}-${event.type}-${event.source_instance_id ?? "none"}-${event.source_template_id ?? "none"}-${targetFingerprint}`,
      );
      if (processedEventIdsRef.current.has(eventId)) {
        return;
      }
      processedEventIdsRef.current.add(eventId);

      const isHeroSource = event.source_kind === "hero" || (event.source_instance_id ?? "").startsWith("hero:");
      if (isHeroSource) {
        return;
      }

      const templateId = resolveEventTemplateId(match, event);
      if (!templateId) {
        return;
      }

      if (event.type === "card_skill") {
        playBattleCardSfx(templateId, "spell", 0.76);
      }
    });
  }, [match]);
}

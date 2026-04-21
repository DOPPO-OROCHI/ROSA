import { useEffect, useRef, useState } from "react";
import { COMMON_ATTACK_HIT_SFX_SRC, playAudioSource } from "./sfx";
import type { MaskedBattleMatchState } from "./types";

const HIT_SFX_STAGGER_MS = 45;
const HIT_SFX_ATTACK_DELAY_MS = 120;

export function useOnHitEffects(match: MaskedBattleMatchState | null) {
  const [hitTokens, setHitTokens] = useState<Record<string, number>>({});
  const lastProcessedVersionRef = useRef(0);
  const pendingSoundTimeoutsRef = useRef<number[]>([]);
  const processedEventIdsRef = useRef(new Set<string>());

  useEffect(() => {
    return () => {
      pendingSoundTimeoutsRef.current.forEach((timeoutId) => window.clearTimeout(timeoutId));
      pendingSoundTimeoutsRef.current = [];
    };
  }, []);

  function playAttackImpactSound(hitCount: number, initialDelayMs = 0) {
    if (hitCount <= 0) {
      return;
    }

    for (let index = 0; index < hitCount; index += 1) {
      const timeoutId = window.setTimeout(() => {
        pendingSoundTimeoutsRef.current = pendingSoundTimeoutsRef.current.filter((entry) => entry !== timeoutId);
        playAudioSource(COMMON_ATTACK_HIT_SFX_SRC, 0.72);
      }, initialDelayMs + index * HIT_SFX_STAGGER_MS);
      pendingSoundTimeoutsRef.current.push(timeoutId);
    }
  }

  useEffect(() => {
    if (!match || match.version <= lastProcessedVersionRef.current) {
      return;
    }

    lastProcessedVersionRef.current = match.version;

    const nextHits = new Map<string, number>();
    match.events?.forEach((event, eventIndex) => {
      const targetFingerprint = (event.targets ?? [])
        .map((target) => `${target.instance_id ?? "none"}:${target.amount ?? 0}:${target.new_hp ?? "na"}:${target.died ? 1 : 0}`)
        .join("|");
      const eventId = event.id
        ?? event.event_id
        ?? `${eventIndex}-${event.type}-${event.source_instance_id ?? "none"}-${event.source_template_id ?? "none"}-${targetFingerprint}`;
      if (processedEventIdsRef.current.has(eventId)) {
        return;
      }

      const isAttackEvent = event.type.toLowerCase().includes("attack");
      const isHealEvent = event.type.toLowerCase().includes("heal");
      if (isHealEvent || !isAttackEvent) {
        return;
      }

      processedEventIdsRef.current.add(eventId);

      event.targets?.forEach((target) => {
        if (!target.instance_id || !target.amount || target.amount <= 0) {
          return;
        }
        nextHits.set(target.instance_id, (nextHits.get(target.instance_id) ?? 0) + 1);
      });
    });

    if (nextHits.size === 0) {
      return;
    }

    setHitTokens((current) => {
      const next = { ...current };
      let totalHits = 0;
      nextHits.forEach((count, instanceId) => {
        next[instanceId] = (next[instanceId] ?? 0) + count;
        totalHits += count;
      });
      playAttackImpactSound(totalHits, HIT_SFX_ATTACK_DELAY_MS);
      return next;
    });
  }, [match]);

  function triggerHit(instanceId: string, count = 1) {
    if (!instanceId || count <= 0) {
      return;
    }

    setHitTokens((current) => ({
      ...current,
      [instanceId]: (current[instanceId] ?? 0) + count,
    }));
  }

  return {
    hitTokens,
    triggerHit,
  };
}

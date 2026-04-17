import { useEffect, useRef, useState } from "react";
import type { MaskedBattleMatchState } from "./types";

export function useOnHitEffects(match: MaskedBattleMatchState | null) {
  const [hitTokens, setHitTokens] = useState<Record<string, number>>({});
  const lastProcessedVersionRef = useRef(0);

  useEffect(() => {
    if (!match || match.version <= lastProcessedVersionRef.current) {
      return;
    }

    lastProcessedVersionRef.current = match.version;

    const nextHits = new Map<string, number>();
    match.events?.forEach((event) => {
      const isAttackEvent = event.type.toLowerCase().includes("attack");
      const isHealEvent = event.type.toLowerCase().includes("heal");
      if (isHealEvent || !isAttackEvent) {
        return;
      }

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
      nextHits.forEach((count, instanceId) => {
        next[instanceId] = (next[instanceId] ?? 0) + count;
      });
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

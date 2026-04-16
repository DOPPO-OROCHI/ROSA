import { useEffect, useRef, useState } from "react";
import type { MaskedBattleMatchState } from "./types";

export function useOnHitEffects(match: MaskedBattleMatchState | null) {
  const [hitUnitIds, setHitUnitIds] = useState<string[]>([]);
  const lastProcessedVersionRef = useRef(0);

  useEffect(() => {
    if (!match || match.version <= lastProcessedVersionRef.current) {
      return;
    }

    lastProcessedVersionRef.current = match.version;

    const nextIds = new Set<string>();
    match.events?.forEach((event) => {
      const isHealEvent = event.type.toLowerCase().includes("heal");
      if (isHealEvent) {
        return;
      }

      event.targets?.forEach((target) => {
        if (!target.instance_id || !target.amount || target.amount <= 0) {
          return;
        }
        if (target.instance_id.startsWith("hero:")) {
          return;
        }
        nextIds.add(target.instance_id);
      });
    });

    if (nextIds.size === 0) {
      return;
    }

    const ids = Array.from(nextIds);
    setHitUnitIds((current) => Array.from(new Set([...current, ...ids])));

    const timeoutId = window.setTimeout(() => {
      setHitUnitIds((current) => current.filter((id) => !nextIds.has(id)));
    }, 320);

    return () => window.clearTimeout(timeoutId);
  }, [match]);

  return {
    hitUnitIds,
  };
}

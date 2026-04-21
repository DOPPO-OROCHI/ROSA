import { useEffect, useMemo, useRef, useState } from "react";
import type { BattleEvent, BattleEventTarget, BattleUnitState, MaskedBattleMatchState } from "../types";
import { getUnitSkill, resolveProjectileRule } from "./projectileConfig";
import type {
  ProjectileImpact,
  ProjectileRule,
  ProjectileSpread,
  ProjectileRuntimeState,
  ProjectileSequence,
  ProjectileSnapshot,
  ProjectileStep,
  UnitRect,
} from "./types";

const PROJECTILE_DURATION_MS = 180;
const PROJECTILE_IMPACT_MS = 80;
const PROJECTILE_SPREAD_MS = 40;
const PROJECTILE_QUEUE_GAP_MS = 40;

type Params = {
  match: MaskedBattleMatchState | null;
  playerIndex: number;
  getUnitRect: (instanceId: string) => UnitRect | null;
};

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

function uniqueTargetIds(targets: BattleEventTarget[] | undefined): string[] {
  const seen = new Set<string>();
  const ids: string[] = [];

  (targets ?? []).forEach((target) => {
    const instanceId = target.instance_id ?? "";
    if (!instanceId || seen.has(instanceId)) {
      return;
    }
    seen.add(instanceId);
    ids.push(instanceId);
  });

  return ids;
}

function getEventSequenceId(event: BattleEvent, fallback: string): string {
  return event.id ?? event.event_id ?? fallback;
}

function buildImpact(
  id: string,
  tone: ProjectileRule["tone"],
  targetRect: UnitRect,
): ProjectileImpact {
  return {
    id,
    tone,
    centerX: targetRect.centerX,
    centerY: targetRect.centerY,
    width: 42,
    height: 42,
    flashLeft: targetRect.left,
    flashTop: targetRect.top,
    flashWidth: targetRect.width,
    flashHeight: targetRect.height,
  };
}

function buildStep(
  id: string,
  tone: ProjectileRule["tone"],
  sourceRect: UnitRect,
  targetRect: UnitRect,
  targetInstanceId: string,
  spreadImpacts?: ProjectileImpact[],
): ProjectileStep {
  return {
    id,
    tone,
    fromX: sourceRect.centerX,
    fromY: sourceRect.centerY,
    toX: targetRect.centerX,
    toY: targetRect.centerY,
    targetInstanceId,
    flashLeft: targetRect.left,
    flashTop: targetRect.top,
    flashWidth: targetRect.width,
    flashHeight: targetRect.height,
    spreadImpacts,
  };
}

function buildSpread(
  id: string,
  tone: ProjectileRule["tone"],
  sourceRect: UnitRect,
  targetRect: UnitRect,
): ProjectileSpread {
  return {
    id,
    tone,
    fromX: sourceRect.centerX,
    fromY: sourceRect.centerY,
    toX: targetRect.centerX,
    toY: targetRect.centerY,
    dx: targetRect.centerX - sourceRect.centerX,
    dy: targetRect.centerY - sourceRect.centerY,
    width: 0,
    height: 8,
  };
}

function buildSequenceFromEvent(
  match: MaskedBattleMatchState,
  event: BattleEvent,
  eventIndex: number,
  playerIndex: number,
  getUnitRect: (instanceId: string) => UnitRect | null,
): ProjectileSequence | null {
  if (event.type !== "card_skill") {
    return null;
  }
  if (event.visible_for_player_index != null && event.visible_for_player_index !== playerIndex) {
    return null;
  }
  if (!event.source_instance_id) {
    return null;
  }

  const sourceRect = getUnitRect(event.source_instance_id);
  const sourceUnit = findUnit(match, event.source_instance_id);
  if (!sourceRect || !sourceUnit) {
    return null;
  }

  const skill = getUnitSkill(sourceUnit);
  const rule = resolveProjectileRule(skill, event);
  if (rule.mode === "none") {
    return null;
  }

  const targetIds = uniqueTargetIds(event.targets);
  if (targetIds.length === 0) {
    return null;
  }

  const sequenceId = getEventSequenceId(event, `${match.version}-${eventIndex}-${event.source_instance_id}`);

  if (rule.mode === "single") {
    const primaryTargetId = targetIds[0];
    const primaryRect = getUnitRect(primaryTargetId);
    if (!primaryRect) {
      return null;
    }

    return {
      id: sequenceId,
      steps: [buildStep(`${sequenceId}-0`, rule.tone, sourceRect, primaryRect, primaryTargetId)],
    };
  }

  if (rule.mode === "splash") {
    const primaryTargetId = targetIds[0];
    const primaryRect = getUnitRect(primaryTargetId);
    if (!primaryRect) {
      return null;
    }

    const spreadImpacts = targetIds
      .slice(1)
      .map((targetId, index) => {
        const rect = getUnitRect(targetId);
        if (!rect) {
          return null;
        }
        return buildImpact(`${sequenceId}-spread-${index}`, rule.tone, rect);
      })
      .filter((impact): impact is ProjectileImpact => impact !== null);

    return {
      id: sequenceId,
      steps: [buildStep(`${sequenceId}-0`, rule.tone, sourceRect, primaryRect, primaryTargetId, spreadImpacts)],
    };
  }

  const steps = targetIds
    .map((targetId, index) => {
      const rect = getUnitRect(targetId);
      if (!rect) {
        return null;
      }
      return buildStep(`${sequenceId}-${index}`, rule.tone, sourceRect, rect, targetId);
    })
    .filter((step): step is ProjectileStep => step !== null);

  if (steps.length === 0) {
    return null;
  }

  return {
    id: sequenceId,
    steps,
  };
}

export function useProjectileRuntime({ match, playerIndex, getUnitRect }: Params): ProjectileRuntimeState {
  const [projectiles, setProjectiles] = useState<ProjectileSnapshot[]>([]);
  const [impacts, setImpacts] = useState<ProjectileImpact[]>([]);
  const [spreads, setSpreads] = useState<ProjectileSpread[]>([]);
  const queueRef = useRef<ProjectileSequence[]>([]);
  const activeSequenceRef = useRef(false);
  const processedEventIdsRef = useRef(new Set<string>());
  const timeoutIdsRef = useRef<number[]>([]);

  function clearTimer(id: number) {
    timeoutIdsRef.current = timeoutIdsRef.current.filter((value) => value !== id);
    window.clearTimeout(id);
  }

  function schedule(callback: () => void, delayMs: number) {
    const timeoutId = window.setTimeout(() => {
      clearTimer(timeoutId);
      callback();
    }, delayMs);
    timeoutIdsRef.current.push(timeoutId);
  }

  function spawnImpact(impact: ProjectileImpact) {
    setImpacts((current) => [...current, impact]);
    schedule(() => {
      setImpacts((current) => current.filter((entry) => entry.id !== impact.id));
    }, PROJECTILE_IMPACT_MS);
  }

  function spawnSpread(spread: ProjectileSpread) {
    setSpreads((current) => [...current, spread]);
    schedule(() => {
      setSpreads((current) => current.filter((entry) => entry.id !== spread.id));
    }, PROJECTILE_SPREAD_MS);
  }

  function runNextSequence() {
    if (activeSequenceRef.current) {
      return;
    }

    const nextSequence = queueRef.current.shift();
    if (!nextSequence) {
      return;
    }

    activeSequenceRef.current = true;
    const totalDuration = nextSequence.steps.length === 0
      ? 0
      : PROJECTILE_DURATION_MS + PROJECTILE_IMPACT_MS + PROJECTILE_SPREAD_MS + (nextSequence.steps.length - 1) * PROJECTILE_QUEUE_GAP_MS;

    nextSequence.steps.forEach((step, index) => {
      schedule(() => {
        const projectile: ProjectileSnapshot = {
          id: step.id,
          tone: step.tone,
          fromX: step.fromX,
          fromY: step.fromY,
          toX: step.toX,
          toY: step.toY,
          dx: step.toX - step.fromX,
          dy: step.toY - step.fromY,
          width: 52,
          height: 12,
        };

        setProjectiles((current) => [...current, projectile]);

        schedule(() => {
          setProjectiles((current) => current.filter((entry) => entry.id !== projectile.id));
          spawnImpact(buildImpact(`${step.id}-impact`, step.tone, {
            left: step.flashLeft,
            top: step.flashTop,
            width: step.flashWidth,
            height: step.flashHeight,
            centerX: step.toX,
            centerY: step.toY,
          }));

          const primaryRect: UnitRect = {
            left: step.flashLeft,
            top: step.flashTop,
            width: step.flashWidth,
            height: step.flashHeight,
            centerX: step.toX,
            centerY: step.toY,
          };

          (step.spreadImpacts ?? []).forEach((impact, spreadIndex) => {
            spawnSpread(buildSpread(`${step.id}-spread-beam-${spreadIndex}`, step.tone, primaryRect, {
              left: impact.flashLeft,
              top: impact.flashTop,
              width: impact.flashWidth,
              height: impact.flashHeight,
              centerX: impact.centerX,
              centerY: impact.centerY,
            }));
            spawnImpact({ ...impact, id: `${impact.id}-${spreadIndex}` });
          });
        }, PROJECTILE_DURATION_MS);
      }, index * PROJECTILE_QUEUE_GAP_MS);
    });

    schedule(() => {
      activeSequenceRef.current = false;
      runNextSequence();
    }, totalDuration);
  }

  useEffect(() => {
    if (!match?.events?.length) {
      return;
    }

    const nextSequences = match.events
      .map((event, eventIndex) => buildSequenceFromEvent(match, event, eventIndex, playerIndex, getUnitRect))
      .filter((sequence): sequence is ProjectileSequence => sequence !== null)
      .filter((sequence) => {
        if (processedEventIdsRef.current.has(sequence.id)) {
          return false;
        }
        processedEventIdsRef.current.add(sequence.id);
        return true;
      });

    if (nextSequences.length === 0) {
      return;
    }

    queueRef.current.push(...nextSequences);
    runNextSequence();
  }, [getUnitRect, match, playerIndex]);

  useEffect(() => {
    return () => {
      timeoutIdsRef.current.forEach((timeoutId) => window.clearTimeout(timeoutId));
    };
  }, []);

  return useMemo(
    () => ({
      projectiles,
      impacts,
      spreads,
    }),
    [impacts, projectiles, spreads],
  );
}

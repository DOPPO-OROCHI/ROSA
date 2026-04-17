import { useEffect, useMemo, useState } from "react";
import type { BattleUnitState, MaskedBattleMatchState } from "../types";
import { getUnitSkill } from "./utils";

type UnitRect = {
  left: number;
  top: number;
  width: number;
  height: number;
  centerX: number;
  centerY: number;
};

type SkillProjectileKind = "damage" | "heal" | "buff" | "debuff";

type SkillProjectileState = {
  id: string;
  fromX: number;
  fromY: number;
  toX: number;
  toY: number;
  targetInstanceId: string;
  kind: SkillProjectileKind;
  durationMs: number;
  arcHeight: number;
};

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

function getProjectileKind(kind: string | undefined): SkillProjectileKind {
  switch (kind) {
    case "heal":
      return "heal";
    case "buff":
    case "summon":
      return "buff";
    case "debuff":
      return "debuff";
    default:
      return "damage";
  }
}

export function useSkillProjectiles({ match, playerIndex, getUnitRect }: Params) {
  const [projectiles, setProjectiles] = useState<SkillProjectileState[]>([]);

  useEffect(() => {
    if (!match?.events?.length) {
      return;
    }

    const nextProjectiles: SkillProjectileState[] = [];

    match.events.forEach((event, eventIndex) => {
      if (event.type !== "card_skill") {
        return;
      }
      if (event.visible_for_player_index != null && event.visible_for_player_index !== playerIndex) {
        return;
      }
      if (!event.source_instance_id) {
        return;
      }

      const sourceRect = getUnitRect(event.source_instance_id);
      const sourceUnit = findUnit(match, event.source_instance_id);
      if (!sourceRect || !sourceUnit) {
        return;
      }

      const kind = getProjectileKind(getUnitSkill(sourceUnit)?.kind);
      if (kind !== "damage" && kind !== "debuff") {
        return;
      }

      (event.targets ?? []).forEach((target, targetIndex) => {
        if (!target.instance_id) {
          return;
        }
        const targetRect = getUnitRect(target.instance_id);
        if (!targetRect) {
          return;
        }

        const dx = targetRect.centerX - sourceRect.centerX;
        const dy = targetRect.centerY - sourceRect.centerY;
        const distance = Math.hypot(dx, dy);

        nextProjectiles.push({
          id: `${match.version}-${eventIndex}-${targetIndex}-${target.instance_id}`,
          fromX: sourceRect.centerX,
          fromY: sourceRect.centerY,
          toX: targetRect.centerX,
          toY: targetRect.centerY,
          targetInstanceId: target.instance_id,
          kind,
          durationMs: Math.max(320, Math.min(520, Math.round(distance * 0.62))),
          arcHeight: Math.max(18, Math.min(54, distance * 0.22)),
        });
      });
    });

    if (nextProjectiles.length === 0) {
      return;
    }

    setProjectiles((current) => [...current, ...nextProjectiles]);
  }, [getUnitRect, match, playerIndex]);

  return {
    projectiles,
    removeProjectile(id: string) {
      setProjectiles((current) => current.filter((entry) => entry.id !== id));
    },
  };
}

type SkillProjectilesProps = {
  projectiles: SkillProjectileState[];
  onImpact?: (targetInstanceId: string, kind: SkillProjectileKind) => void;
  onDone?: (id: string) => void;
};

type SkillProjectileNodeProps = {
  projectile: SkillProjectileState;
  onImpact?: (targetInstanceId: string, kind: SkillProjectileKind) => void;
  onDone?: (id: string) => void;
};

function SkillProjectileNode({ projectile, onImpact, onDone }: SkillProjectileNodeProps) {
  const [progress, setProgress] = useState(0);

  useEffect(() => {
    let rafId = 0;
    let impactSent = false;
    const startedAt = performance.now();

    function frame(now: number) {
      const elapsed = now - startedAt;
      const nextProgress = Math.min(1, elapsed / projectile.durationMs);
      setProgress(nextProgress);

      if (!impactSent && nextProgress >= 0.96) {
        impactSent = true;
        onImpact?.(projectile.targetInstanceId, projectile.kind);
      }

      if (nextProgress >= 1) {
        onDone?.(projectile.id);
        return;
      }

      rafId = window.requestAnimationFrame(frame);
    }

    rafId = window.requestAnimationFrame(frame);
    return () => window.cancelAnimationFrame(rafId);
  }, [onDone, onImpact, projectile]);

  const x = projectile.fromX + (projectile.toX - projectile.fromX) * progress;
  const yLinear = projectile.fromY + (projectile.toY - projectile.fromY) * progress;
  const arcOffset = Math.sin(progress * Math.PI) * projectile.arcHeight;
  const y = yLinear - arcOffset;
  const scale = 0.72 + Math.sin(progress * Math.PI * 0.9) * 0.34;
  const opacity = progress < 0.08 ? progress / 0.08 : progress > 0.92 ? (1 - progress) / 0.08 : 1;

  return (
    <div
      className={`battle-skill-projectile battle-skill-projectile--${projectile.kind}`}
      style={{
        left: `${x}px`,
        top: `${y}px`,
        transform: `translate(-50%, -50%) scale(${scale})`,
        opacity,
      }}
    />
  );
}

export function SkillProjectiles({ projectiles, onImpact, onDone }: SkillProjectilesProps) {
  const visibleProjectiles = useMemo(() => projectiles.filter((projectile) => projectile.kind === "damage" || projectile.kind === "debuff"), [projectiles]);

  if (visibleProjectiles.length === 0) {
    return null;
  }

  return (
    <div className="battle-skill-projectiles-layer" aria-hidden="true">
      {visibleProjectiles.map((projectile) => (
        <SkillProjectileNode
          key={projectile.id}
          projectile={projectile}
          onImpact={onImpact}
          onDone={onDone}
        />
      ))}
    </div>
  );
}

export type { SkillProjectileKind };

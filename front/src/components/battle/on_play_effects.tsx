import { useEffect, useRef, useState, type CSSProperties } from "react";
import { resolveCardAssetVariantSrc } from "../../lib/api";
import { getBoardAttackDisplayKind, getBoardAttackDisplayValue } from "./card_attack";
import { getBoardSkillLabel } from "./CARD_SKILLS";
import { playBattleCardSfx } from "./sfx";
import type { BattleUnitState } from "./types";

type UnitRect = {
  left: number;
  top: number;
  width: number;
  height: number;
  centerX: number;
  centerY: number;
};

type PlayAnimationState = {
  id: string;
  unit: BattleUnitState;
  from: {
    left: number;
    top: number;
    width: number;
    height: number;
  };
  dx: number;
  dy: number;
};

type UseOnPlayEffectsParams = {
  version: number;
  playerTable: Array<BattleUnitState | null>;
  enemyTable: Array<BattleUnitState | null>;
  shellWidth: number;
  shellHeight: number;
  getUnitRect: (instanceId: string) => UnitRect | null;
};

function getOriginForSlot(
  side: "player" | "enemy",
  slotIndex: number,
  shellWidth: number,
  shellHeight: number,
  target: UnitRect,
) {
  const fromLeft = slotIndex <= 1;
  const fromRight = slotIndex >= 3;

  if (side === "enemy") {
    if (fromLeft) {
      return { left: -target.width * 0.8, top: -target.height * 1.1 };
    }
    if (fromRight) {
      return { left: shellWidth - target.width * 0.2, top: -target.height * 1.1 };
    }
    return { left: shellWidth * 0.5 - target.width * 0.5, top: -target.height * 1.2 };
  }

  if (fromLeft) {
    return { left: -target.width * 0.8, top: shellHeight - target.height * 0.1 };
  }
  if (fromRight) {
    return { left: shellWidth - target.width * 0.2, top: shellHeight - target.height * 0.1 };
  }
  return { left: shellWidth * 0.5 - target.width * 0.5, top: shellHeight - target.height * 0.02 };
}

export function useOnPlayEffects({
  version,
  playerTable,
  enemyTable,
  shellWidth,
  shellHeight,
  getUnitRect,
}: UseOnPlayEffectsParams) {
  const [animations, setAnimations] = useState<PlayAnimationState[]>([]);
  const prevTablesRef = useRef<{ player: Array<BattleUnitState | null>; enemy: Array<BattleUnitState | null> } | null>(null);
  const lastVersionRef = useRef(0);

  useEffect(() => {
    if (!version || version <= lastVersionRef.current) {
      return;
    }

    lastVersionRef.current = version;

    const prev = prevTablesRef.current;
    prevTablesRef.current = { player: playerTable, enemy: enemyTable };

    if (!prev || !shellWidth || !shellHeight) {
      return;
    }

    const detected: PlayAnimationState[] = [];

    ([
      ["enemy", prev.enemy, enemyTable],
      ["player", prev.player, playerTable],
    ] as const).forEach(([side, before, after]) => {
      for (let slotIndex = 0; slotIndex < after.length; slotIndex += 1) {
        const prevUnit = before[slotIndex];
        const nextUnit = after[slotIndex];
        if (prevUnit || !nextUnit) {
          continue;
        }

        const target = getUnitRect(nextUnit.instance_id);
        if (!target) {
          continue;
        }

        const origin = getOriginForSlot(side, slotIndex, shellWidth, shellHeight, target);
        detected.push({
          id: `${version}-${side}-${slotIndex}-${nextUnit.instance_id}`,
          unit: nextUnit,
          from: {
            left: origin.left,
            top: origin.top,
            width: target.width,
            height: target.height,
          },
          dx: target.centerX - (origin.left + target.width / 2),
          dy: target.centerY - (origin.top + target.height / 2),
        });
      }
    });

    if (detected.length === 0) {
      return;
    }

    detected.forEach((entry) => {
      if (entry.unit.card_type !== "hero") {
        playBattleCardSfx(entry.unit.template_id, "summon", 0.78);
      }
    });

    setAnimations((current) => [...current, ...detected]);
  }, [enemyTable, getUnitRect, playerTable, shellHeight, shellWidth, version]);

  function removeAnimation(id: string) {
    setAnimations((current) => current.filter((entry) => entry.id !== id));
  }

  return {
    animations,
    removeAnimation,
  };
}

type PlayAnimationsProps = {
  animations: PlayAnimationState[];
  onDone: (id: string) => void;
};

export function BattlePlayAnimations({ animations, onDone }: PlayAnimationsProps) {
  if (animations.length === 0) {
    return null;
  }

  return (
    <div className="battle-play-animation-layer" aria-hidden="true">
      {animations.map((entry) => {
        const skillLabel = getBoardSkillLabel(entry.unit);
        const primaryValue = getBoardAttackDisplayValue(entry.unit);
        const primaryKind = getBoardAttackDisplayKind(entry.unit);

        return (
          <div
            key={entry.id}
            className="battle-play-animation"
            style={
              {
                left: `${entry.from.left}px`,
                top: `${entry.from.top}px`,
                width: `${entry.from.width}px`,
                height: `${entry.from.height}px`,
                "--battle-play-dx": `${entry.dx}px`,
                "--battle-play-dy": `${entry.dy}px`,
              } as CSSProperties
            }
            onAnimationEnd={() => onDone(entry.id)}
          >
            <img
              className="battle-play-animation__art"
              src={resolveCardAssetVariantSrc("battle", entry.unit.template_id, "on_table")}
              alt={entry.unit.template_id}
              onError={(event) => {
                const target = event.currentTarget;
                if (target.dataset.fallbackApplied === "1") {
                  return;
                }
                target.dataset.fallbackApplied = "1";
                target.src = resolveCardAssetVariantSrc("battle", entry.unit.template_id, "view");
              }}
            />
            <span className={`battle-board-slot__attack battle-board-slot__attack--${primaryKind}`}>{primaryValue}</span>
            <span className="battle-board-slot__cooldown">{entry.unit.cooldown}</span>
            <span className="battle-board-slot__skill-label">{skillLabel}</span>
            <span className="battle-board-slot__hp">{entry.unit.hp}</span>
          </div>
        );
      })}
    </div>
  );
}

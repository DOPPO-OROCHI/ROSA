import { useEffect, useRef, type CSSProperties } from "react";
import { resolveCardAssetVariantSrc } from "../../lib/api";
import { playBattleCardSfx } from "./sfx";
import type { BattleUnitState, MaskedBattleMatchState } from "./types";

type UnitRect = {
  left: number;
  top: number;
  width: number;
  height: number;
  centerX: number;
  centerY: number;
};

export type DeathAnimationState = {
  id: string;
  unit: BattleUnitState;
  rect: {
    left: number;
    top: number;
    width: number;
    height: number;
  };
};

type FragmentConfig = {
  id: string;
  clipPath: string;
  dx: number;
  dy: number;
  rotate: number;
  scale: number;
};

type CollectDeathAnimationsParams = {
  prevMatch: MaskedBattleMatchState | null;
  nextMatch: MaskedBattleMatchState;
  getUnitRect: (instanceId: string) => UnitRect | null;
};

export function collectDeathAnimations({
  prevMatch,
  nextMatch,
  getUnitRect,
}: CollectDeathAnimationsParams): DeathAnimationState[] {
  if (!prevMatch || !nextMatch.events?.length) {
    return [];
  }

  const prevUnits = new Map<string, BattleUnitState>();
  prevMatch.players.forEach((player) => {
    player?.table.forEach((unit) => {
      if (unit?.instance_id) {
        prevUnits.set(unit.instance_id, unit);
      }
    });
  });

  const deaths: DeathAnimationState[] = [];

  nextMatch.events.forEach((event, eventIndex) => {
    event.targets?.forEach((target, targetIndex) => {
      if (!target.died || !target.instance_id || target.instance_id.startsWith("hero:")) {
        return;
      }

      const unit = prevUnits.get(target.instance_id);
      if (!unit) {
        return;
      }

      const rect = getUnitRect(target.instance_id);
      if (!rect) {
        return;
      }

      deaths.push({
        id: `${nextMatch.version}-${eventIndex}-${targetIndex}-${target.instance_id}`,
        unit,
        rect: {
          left: rect.left,
          top: rect.top,
          width: rect.width,
          height: rect.height,
        },
      });
    });
  });

  return deaths;
}

function createSeededRandom(seedSource: string) {
  let seed = 2166136261;
  for (let index = 0; index < seedSource.length; index += 1) {
    seed ^= seedSource.charCodeAt(index);
    seed = Math.imul(seed, 16777619);
  }

  return () => {
    seed += 0x6d2b79f5;
    let value = seed;
    value = Math.imul(value ^ (value >>> 15), value | 1);
    value ^= value + Math.imul(value ^ (value >>> 7), value | 61);
    return ((value ^ (value >>> 14)) >>> 0) / 4294967296;
  };
}

function buildShatterFragments(seedSource: string): FragmentConfig[] {
  const random = createSeededRandom(seedSource);

  const left = 18 + Math.round(random() * 10);
  const right = 82 - Math.round(random() * 10);
  const upperMid = 28 + Math.round(random() * 10);
  const lowerMid = 60 + Math.round(random() * 12);
  const centerX = 50 + Math.round((random() - 0.5) * 10);
  const centerY = 48 + Math.round((random() - 0.5) * 10);

  const clipPaths = [
    `polygon(0% 0%, ${left}% 0%, ${centerX}% ${upperMid}%, 0% ${lowerMid}%)`,
    `polygon(${left}% 0%, ${right}% 0%, ${centerX}% ${centerY}%, ${centerX - 8}% ${upperMid + 10}%)`,
    `polygon(${right}% 0%, 100% 0%, 100% ${lowerMid}%, ${centerX}% ${centerY}%)`,
    `polygon(0% ${lowerMid}%, ${centerX - 6}% ${centerY + 6}%, ${left + 4}% 100%, 0% 100%)`,
    `polygon(${centerX - 8}% ${centerY + 4}%, ${centerX + 8}% ${centerY + 2}%, ${right - 6}% 100%, ${left + 8}% 100%)`,
    `polygon(${centerX}% ${centerY}%, 100% ${lowerMid}%, 100% 100%, ${right - 8}% 100%)`,
  ];

  const directions = [
    { dx: -42 - random() * 34, dy: -58 - random() * 20 },
    { dx: -12 + random() * 18, dy: -72 - random() * 16 },
    { dx: 38 + random() * 34, dy: -54 - random() * 18 },
    { dx: -34 - random() * 26, dy: 18 + random() * 26 },
    { dx: -4 + random() * 12, dy: 42 + random() * 28 },
    { dx: 34 + random() * 26, dy: 20 + random() * 26 },
  ];

  return clipPaths.map((clipPath, index) => ({
    id: `${seedSource}-${index}`,
    clipPath,
    dx: directions[index]?.dx ?? 0,
    dy: directions[index]?.dy ?? 0,
    rotate: -28 + random() * 56,
    scale: 0.82 + random() * 0.26,
  }));
}

type Props = {
  animations: DeathAnimationState[];
  onDone: (id: string) => void;
};

export function BattleDeathAnimations({ animations, onDone }: Props) {
  const playedAnimationIdsRef = useRef(new Set<string>());

  useEffect(() => {
    animations.forEach((entry) => {
      if (playedAnimationIdsRef.current.has(entry.id)) {
        return;
      }
      playedAnimationIdsRef.current.add(entry.id);
      if (entry.unit.card_type !== "hero") {
        playBattleCardSfx(entry.unit.template_id, "death", 0.82);
      }
    });
  }, [animations]);

  if (animations.length === 0) {
    return null;
  }

  return (
    <div className="battle-death-animation-layer" aria-hidden="true">
      {animations.map((entry) => {
        const fragments = buildShatterFragments(entry.id);
        const artSrc = resolveCardAssetVariantSrc("battle", entry.unit.template_id, "on_table");

        return (
          <div
            key={entry.id}
            className="battle-death-animation"
            style={
              {
                left: `${entry.rect.left}px`,
                top: `${entry.rect.top}px`,
                width: `${entry.rect.width}px`,
                height: `${entry.rect.height}px`,
              } as CSSProperties
            }
            onAnimationEnd={() => onDone(entry.id)}
          >
            <div className="battle-death-animation__glow" />
            <div className="battle-death-animation__base">
              <img
                className="battle-death-animation__art"
                src={artSrc}
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
            </div>
            <div className="battle-death-animation__flash" />
            {fragments.map((fragment) => (
              <div
                key={fragment.id}
                className="battle-death-animation__fragment"
                style={
                  {
                    clipPath: fragment.clipPath,
                    backgroundImage: `url("${artSrc}")`,
                    "--battle-death-fragment-dx": `${fragment.dx}px`,
                    "--battle-death-fragment-dy": `${fragment.dy}px`,
                    "--battle-death-fragment-rotate": `${fragment.rotate}deg`,
                    "--battle-death-fragment-scale": fragment.scale,
                  } as CSSProperties
                }
              />
            ))}
          </div>
        );
      })}
    </div>
  );
}

import { useEffect, useRef } from "react";
import { resolveHeroAssetVariantSrc } from "../../lib/api";
import type { MaskedBattlePlayerState } from "./types";

type Props = {
  player: MaskedBattlePlayerState;
  maxHp: number;
  side: "player" | "enemy";
  isActive?: boolean;
  attackTarget?: boolean;
  heroInstanceId?: string;
  hitToken?: number;
  attackAnimation?: {
    dx: number;
    dy: number;
  } | null;
  onClick?: () => void;
};

const HERO_MANA_CAP = 10;
const ARC_SIZE = 156;
const ARC_RADIUS = 60;
const ARC_LENGTH = Math.PI * ARC_RADIUS;
const TAU = Math.PI * 2;
const ARC_GAP = 0.028;
const MANA_SEGMENT_GAP = 0.06;
const ARC_DIVIDER_INNER = ARC_RADIUS - 8;
const ARC_DIVIDER_OUTER = ARC_RADIUS + 3;

function clamp01(value: number) {
  return Math.max(0, Math.min(1, value));
}

function polarToCartesian(cx: number, cy: number, radius: number, angle: number) {
  return {
    x: cx + radius * Math.cos(angle),
    y: cy + radius * Math.sin(angle),
  };
}

function describeArcPath(startAngle: number, endAngle: number) {
  const center = ARC_SIZE / 2;
  const start = polarToCartesian(center, center, ARC_RADIUS, startAngle);
  const end = polarToCartesian(center, center, ARC_RADIUS, endAngle);
  const largeArcFlag = endAngle - startAngle <= Math.PI ? 0 : 1;
  return `M ${start.x} ${start.y} A ${ARC_RADIUS} ${ARC_RADIUS} 0 ${largeArcFlag} 1 ${end.x} ${end.y}`;
}

function buildManaSegments(startAngle: number, endAngle: number, segments: number) {
  if (segments <= 0) {
    return [];
  }

  const totalSpan = endAngle - startAngle;
  const totalGap = MANA_SEGMENT_GAP * Math.max(0, segments - 1);
  const segmentSpan = (totalSpan - totalGap) / segments;

  return Array.from({ length: segments }, (_, index) => {
    const segmentStart = startAngle + index * (segmentSpan + MANA_SEGMENT_GAP);
    const segmentEnd = segmentStart + segmentSpan;
    return describeArcPath(segmentStart, segmentEnd);
  });
}

export function CharacterBlock({
  player,
  maxHp,
  side,
  isActive = false,
  attackTarget = false,
  heroInstanceId = "",
  hitToken = 0,
  attackAnimation = null,
  onClick,
}: Props) {
  const hpRatio = clamp01(player.hero_hp / Math.max(1, maxHp));
  const hpOnTop = side === "player";
  const manaMax = Math.max(player.mana, Math.min(Math.max(player.turns, 1), HERO_MANA_CAP));
  const topStart = Math.PI + ARC_GAP;
  const topEnd = TAU - ARC_GAP;
  const bottomStart = ARC_GAP;
  const bottomEnd = Math.PI - ARC_GAP;
  const hpPath = hpOnTop ? describeArcPath(topStart, topEnd) : describeArcPath(bottomStart, bottomEnd);
  const hpRatioStyle = { strokeDasharray: `${ARC_LENGTH * hpRatio} ${ARC_LENGTH}` };
  const manaSegments = hpOnTop
    ? buildManaSegments(bottomStart, bottomEnd, manaMax)
    : buildManaSegments(topStart, topEnd, manaMax);
  const activeManaCount = Math.max(0, Math.min(player.mana, manaMax));
  const avatarRef = useRef<HTMLButtonElement | null>(null);

  useEffect(() => {
    if (!hitToken || !avatarRef.current) {
      return;
    }

    const node = avatarRef.current;
    node.classList.remove("battle-character__avatar--hit");
    void node.offsetWidth;
    node.classList.add("battle-character__avatar--hit");

    const timeoutId = window.setTimeout(() => {
      node.classList.remove("battle-character__avatar--hit");
    }, 320);

    return () => window.clearTimeout(timeoutId);
  }, [hitToken]);

  return (
    <section className={`battle-character battle-character--${side}`}>
      <button
        ref={avatarRef}
        type="button"
        className={`battle-character__avatar ${isActive ? "battle-character__avatar--active" : ""} ${attackTarget ? "battle-character__avatar--attack-target" : ""} ${attackAnimation ? "battle-character__avatar--attacking" : ""}`}
        data-unit-instance-id={heroInstanceId}
        data-hero-side={side}
        data-hero-code={player.hero_code}
        onClick={onClick}
        disabled={!onClick}
      >
        <svg className="battle-character__ring" viewBox={`0 0 ${ARC_SIZE} ${ARC_SIZE}`} aria-hidden="true">
          <path className="battle-character__arc-track" d={hpPath} pathLength={ARC_LENGTH} />
          <path className="battle-character__arc battle-character__arc--hp" d={hpPath} pathLength={ARC_LENGTH} style={hpRatioStyle} />
          {manaSegments.map((path, index) => (
            <path
              key={`mana-track-${index}`}
              className="battle-character__arc-track battle-character__arc-track--mana"
              d={path}
            />
          ))}
          {manaSegments.slice(0, activeManaCount).map((path, index) => (
            <path
              key={`mana-fill-${index}`}
              className="battle-character__arc battle-character__arc--mana battle-character__arc--mana-segment"
              d={path}
            />
          ))}
          <line
            className="battle-character__arc-divider"
            x1={ARC_SIZE / 2 - ARC_DIVIDER_INNER}
            y1={ARC_SIZE / 2}
            x2={ARC_SIZE / 2 - ARC_DIVIDER_OUTER}
            y2={ARC_SIZE / 2}
          />
          <line
            className="battle-character__arc-divider"
            x1={ARC_SIZE / 2 + ARC_DIVIDER_INNER}
            y1={ARC_SIZE / 2}
            x2={ARC_SIZE / 2 + ARC_DIVIDER_OUTER}
            y2={ARC_SIZE / 2}
          />
        </svg>
        <img src={resolveHeroAssetVariantSrc(player.hero_code, "battle_icon")} alt={player.hero_code} />
      </button>
    </section>
  );
}

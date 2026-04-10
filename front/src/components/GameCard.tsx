import { resolveCardFallbackSrc, resolveImageSrc } from "../assets";
import "../game-card.css";

export type GameCardMode = "viewer" | "hand" | "catalog";

export type GameCardData = {
  kind: "battle" | "buff";
  name: string;
  description: string;
  imageKey: string;
  race?: string;
  mana?: number;
  attack?: number;
  hp?: number;
  cooldown?: number;
  skillCooldown?: number;
  buffType?: string;
  buffValue?: number;
  duration?: number;
};

type Props = {
  data: GameCardData;
  mode: GameCardMode;
  className?: string;
};

export function GameCard({ data, mode, className }: Props) {
  const attackValue = data.kind === "battle" ? data.attack ?? 0 : data.buffValue ?? 0;
  const hpValue = data.kind === "battle" ? data.hp ?? 0 : data.duration ?? 0;
  const raceLabel = data.race || (data.kind === "battle" ? "НЕИЗВЕСТНО" : "ЭФФЕКТ");

  return (
    <div className={`game-card game-card--${mode} ${className ?? ""}`.trim()}>
      <img
        className="game-card__image"
        src={resolveImageSrc(data.imageKey)}
        alt={data.name}
        loading="lazy"
        onError={(event) => {
          const target = event.currentTarget;
          if (target.dataset.fallbackApplied === "1") {
            return;
          }
          target.dataset.fallbackApplied = "1";
          target.src = resolveCardFallbackSrc();
        }}
      />

      <div className="game-card__overlay">
        <span className="game-card__stat game-card__stat--mana">{data.mana ?? 0}</span>
        <span className="game-card__stat game-card__stat--attack">{attackValue}</span>
        <span className="game-card__stat game-card__stat--hp">{hpValue}</span>

        <div className="game-card__desc">{data.description}</div>
        {data.kind === "battle" && (
          <div className="game-card__cd">ATK CD {data.cooldown ?? 0} | SKILL CD {data.skillCooldown ?? 0}</div>
        )}
        <div className="game-card__name">{data.name}</div>
        <div className="game-card__race">{raceLabel}</div>
      </div>
    </div>
  );
}


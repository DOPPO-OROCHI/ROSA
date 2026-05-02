import { resolveCardImageSrc } from "../../lib/api";
import type { BattleCardInMatch } from "./types";
import type { CSSProperties, MouseEvent } from "react";

function getHandNameFitClass(name: string): string {
  if (name.length >= 18) {
    return "battle-hand-card__name--long";
  }
  if (name.length >= 11) {
    return "battle-hand-card__name--medium";
  }
  return "battle-hand-card__name--short";
}

type Props = {
  card: BattleCardInMatch;
  selected?: boolean;
  count?: number;
  stackIndex?: number;
  stacked?: boolean;
  fanAngle?: number;
  fanDrop?: number;
  offsetX?: number;
  onOpen?: (cardRect: DOMRect) => void;
};

export function HandCard({
  card,
  selected = false,
  count = 1,
  stackIndex = 0,
  stacked = false,
  fanAngle = 0,
  fanDrop = 0,
  offsetX = 0,
  onOpen,
}: Props) {
  const isBattleCard = card.kind === "battle";
  const lift = selected ? -14 : -4 + fanDrop;
  const style: CSSProperties = {
    zIndex: stackIndex + 1,
    left: `calc(50% + ${offsetX}px)`,
    transform: `translateX(-50%) translateY(${lift}px) rotate(${fanAngle}deg)`,
  };

  return (
    <button
      type="button"
      className={`battle-hand-card ${selected ? "battle-hand-card--selected" : ""} ${stacked ? "battle-hand-card--stacked" : ""}`}
      style={style}
      onClick={(event: MouseEvent<HTMLButtonElement>) => onOpen?.(event.currentTarget.getBoundingClientRect())}
    >
      <div className="inventory-card__frame">
        <img
          className="inventory-card__art"
          src={resolveCardImageSrc(card.kind as "battle" | "buff", card.template_id, card.image_key)}
          alt={card.name}
          loading="lazy"
          onError={(event) => {
            const target = event.currentTarget;
            if (target.dataset.fallbackApplied === "1") {
              return;
            }
            target.dataset.fallbackApplied = "1";
            target.src = card.image_key
              ? `/assets/${card.image_key.replace(/^\/+/, "").replace(/\/+/g, "/")}.png`
              : "/assets/placeholders/card_image.svg";
          }}
        />
        <div className="inventory-card__stat-slot inventory-card__stat-slot--mana">
          <span className="inventory-card__stat">{card.mana_cost}</span>
        </div>
        {isBattleCard ? (
          <>
            <div className="inventory-card__stat-slot inventory-card__stat-slot--attack">
              <span className="inventory-card__stat">{card.attack}</span>
            </div>
            <div className="inventory-card__stat-slot inventory-card__stat-slot--hp">
              <span className="inventory-card__stat">{card.health_points}</span>
            </div>
          </>
        ) : null}
        <span className={`battle-hand-card__name ${getHandNameFitClass(card.name)}`}>{card.name}</span>
        {count > 1 ? <span className="battle-hand-card__count">x{count}</span> : null}
      </div>
    </button>
  );
}

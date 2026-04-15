import { resolveCardImageSrc } from "../../lib/api";
import type { BattleCardInMatch } from "./types";

type Props = {
  card: BattleCardInMatch;
  selected?: boolean;
};

export function HandCard({ card, selected = false }: Props) {
  const isBattleCard = card.kind === "battle";

  return (
    <button type="button" className={`battle-hand-card ${selected ? "battle-hand-card--selected" : ""}`}>
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
      </div>
    </button>
  );
}

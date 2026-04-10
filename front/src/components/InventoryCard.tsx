import { resolveCardImageSrc } from "../lib/api";
import type { BattleCard } from "../types";

type Props = {
  card: BattleCard;
  size?: "deck" | "grid";
};

export function InventoryCard({ card, size = "grid" }: Props) {
  return (
    <article className={`inventory-card inventory-card--${size}`}>
      <div className="inventory-card__frame">
        <img
          className="inventory-card__art"
          src={resolveCardImageSrc(card.kind, card.template_id, card.image_key)}
          alt={card.name}
          loading="lazy"
        />
        <div className="inventory-card__stat-slot inventory-card__stat-slot--mana">
          <span className="inventory-card__stat">{card.mana_cost}</span>
        </div>
        <div className="inventory-card__stat-slot inventory-card__stat-slot--attack">
          <span className="inventory-card__stat">{card.attack}</span>
        </div>
        <div className="inventory-card__stat-slot inventory-card__stat-slot--hp">
          <span className="inventory-card__stat">{card.health_points}</span>
        </div>
        <div className="inventory-card__text-slot">
          <div className="inventory-card__name">{card.name}</div>
          <div className="inventory-card__meta">
            x{card.copies} / max {card.max_copies}
          </div>
        </div>
      </div>
    </article>
  );
}

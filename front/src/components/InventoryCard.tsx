import { resolveCardImageSrc } from "../lib/api";
import type { BattleCard } from "../types";

type Props = {
  card: BattleCard;
  size?: "deck" | "grid";
  description?: string;
  countBadge?: number | null;
  onOpen?: (() => void) | null;
  onAdd?: (() => void) | null;
  onRemove?: (() => void) | null;
  addDisabled?: boolean;
  removeDisabled?: boolean;
};

export function InventoryCard({
  card,
  size = "grid",
  description,
  countBadge = null,
  onOpen = null,
  onAdd = null,
  onRemove = null,
  addDisabled = false,
  removeDisabled = false,
}: Props) {
  return (
    <article
      className={`inventory-card inventory-card--${size} ${onOpen ? "inventory-card--interactive" : ""}`}
      onClick={onOpen ?? undefined}
    >
      <div className="inventory-card__frame">
        <img
          className="inventory-card__art"
          src={resolveCardImageSrc(card.kind, card.template_id, card.image_key)}
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
        <div className="inventory-card__stat-slot inventory-card__stat-slot--attack">
          <span className="inventory-card__stat">{card.attack}</span>
        </div>
        <div className="inventory-card__stat-slot inventory-card__stat-slot--hp">
          <span className="inventory-card__stat">{card.health_points}</span>
        </div>
        {countBadge && countBadge > 1 ? <div className="inventory-card__count-badge">x{countBadge}</div> : null}
        {onRemove ? (
          <button
            type="button"
            className="inventory-card__deck-action inventory-card__deck-action--remove"
            onClick={(event) => {
              event.stopPropagation();
              onRemove();
            }}
            disabled={removeDisabled}
            aria-label={`Remove ${card.name} from deck`}
          >
            x
          </button>
        ) : null}
        {onAdd ? (
          <button
            type="button"
            className="inventory-card__catalog-action"
            onClick={(event) => {
              event.stopPropagation();
              onAdd();
            }}
            disabled={addDisabled}
            aria-label={`Add ${card.name} to deck`}
          >
            +
          </button>
        ) : null}
        <div className="inventory-card__name">{card.name}</div>
        <div className="inventory-card__text-slot">
          <div className="inventory-card__description">{description ?? card.description}</div>
        </div>
      </div>
    </article>
  );
}

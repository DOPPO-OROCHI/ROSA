import { GameCard } from "./GameCard";
import styles from "./inventar_cardsview.module.css";

export type InventoryCatalogKind = "battle" | "buff";
export type InventoryCatalogSort = "mana" | "attack" | "hp" | "tank";

export type InventoryCardItem = {
  kind: "battle" | "buff";
  template_id: string;
  name: string;
  description: string;
  mana_cost: number;
  image_key: string;
  asset_base_key: string;
  max_copies: number;
  copies: number;
  card_type?: string;
  attack?: number;
  health_points?: number;
  cooldown?: number;
  skill_cooldown?: number;
  buff_type?: string;
  buff_value?: number;
  duration?: number;
};

type Props = {
  catalogKind: InventoryCatalogKind;
  catalogSort: InventoryCatalogSort;
  catalogPage: number;
  catalogPages: number;
  catalogPageItems: InventoryCardItem[];
  deckCountMap: Map<string, number>;
  getTone: (assetBaseKey?: string) => string;
  resolveImageKey: (card: InventoryCardItem) => string;
  raceLabel: (cardType?: string) => string;
  onCatalogKindChange: (kind: InventoryCatalogKind) => void;
  onCatalogSortChange: (sort: InventoryCatalogSort) => void;
  onCardPreview: (card: InventoryCardItem, imageKey: string) => void;
  onAddCard: (kind: InventoryCatalogKind, templateId: string) => void;
  onPrevPage: () => void;
  onNextPage: () => void;
};

export function InventarCardsView(props: Props) {
  return (
    <div className={`panel inventory-panel catalog-panel ${styles.root}`}>
      <div className="catalog-toolbar">
        <div className="catalog-kind-switch">
          <button
            className={props.catalogKind === "battle" ? "nav-pill active" : "nav-pill"}
            onClick={() => props.onCatalogKindChange("battle")}
          >
            Battle Cards
          </button>
          <button
            className={props.catalogKind === "buff" ? "nav-pill active" : "nav-pill"}
            onClick={() => props.onCatalogKindChange("buff")}
          >
            Buff Cards
          </button>
        </div>
        <label className="catalog-sort">
          <span>Sort</span>
          <select
            value={props.catalogSort}
            onChange={(event) => props.onCatalogSortChange(event.target.value as InventoryCatalogSort)}
          >
            <option value="mana">Mana</option>
            <option value="attack">Attack</option>
            <option value="hp">HP</option>
            <option value="tank">Tank / Non-tank</option>
          </select>
        </label>
      </div>

      <div className="catalog-grid">
        {props.catalogPageItems.map((card) => {
          const imageKey = props.resolveImageKey(card);
          const templateKey = `${card.kind}:${card.template_id}`;
          const deckCount = props.deckCountMap.get(templateKey) ?? 0;
          const addLimit = Math.min(card.max_copies, card.copies);
          const exhausted = deckCount >= addLimit;
          return (
            <article
              key={templateKey}
              className={`asset-card tone-${props.getTone(card.asset_base_key)} clickable ${exhausted ? "exhausted" : ""}`}
              onClick={() => props.onCardPreview(card, imageKey)}
            >
              <div className="asset-frame">
                <GameCard
                  mode="catalog"
                  data={{
                    kind: card.kind,
                    name: card.name,
                    description: card.description,
                    imageKey,
                    race: card.kind === "battle" ? props.raceLabel(card.card_type) : "ЭФФЕКТ",
                    mana: card.mana_cost,
                    attack: card.kind === "battle" ? card.attack : undefined,
                    hp: card.kind === "battle" ? card.health_points : undefined,
                    cooldown: card.kind === "battle" ? card.cooldown : undefined,
                    skillCooldown: card.kind === "battle" ? card.skill_cooldown : undefined,
                    buffType: card.kind === "buff" ? card.buff_type : undefined,
                    buffValue: card.kind === "buff" ? card.buff_value : undefined,
                    duration: card.kind === "buff" ? card.duration : undefined,
                  }}
                />
                <button
                  className="asset-add"
                  disabled={exhausted}
                  onClick={(event) => {
                    event.stopPropagation();
                    if (exhausted) {
                      return;
                    }
                    props.onAddCard(card.kind, card.template_id);
                  }}
                >
                  +
                </button>
              </div>
            </article>
          );
        })}
      </div>

      <div className="catalog-pager">
        <button className="ghost-button" onClick={props.onPrevPage} disabled={props.catalogPage === 0}>
          {"<"}
        </button>
        <span>
          {props.catalogPage + 1} / {props.catalogPages}
        </span>
        <button
          className="ghost-button"
          onClick={props.onNextPage}
          disabled={props.catalogPage >= props.catalogPages - 1}
        >
          {">"}
        </button>
      </div>
    </div>
  );
}

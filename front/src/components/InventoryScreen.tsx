import { useState } from "react";
import { InventoryCard } from "./InventoryCard";
import type { BattleCard, BuffCard, DeckEntry } from "../types";

type Props = {
  deckEntries: DeckEntry[];
  battleCards: BattleCard[];
  buffCards: BuffCard[];
  onBack: () => void;
};

function flattenDeck(entries: DeckEntry[], battleCards: BattleCard[]): BattleCard[] {
  const battleMap = new Map(battleCards.map((card) => [card.template_id, card] as const));
  const cards: BattleCard[] = [];

  entries.forEach((entry) => {
    if (entry.kind !== "battle") {
      return;
    }
    const card = battleMap.get(entry.template_id);
    if (!card) {
      return;
    }
    for (let index = 0; index < entry.count; index += 1) {
      cards.push(card);
    }
  });

  return cards;
}

export function InventoryScreen({ deckEntries, battleCards, buffCards, onBack }: Props) {
  const deckCards = flattenDeck(deckEntries, battleCards);
  const previewDeckCards = deckCards.slice(0, 4);
  const [showFullDeck, setShowFullDeck] = useState(false);
  const [catalogKind, setCatalogKind] = useState<"battle" | "buff">("battle");
  const [battlePage, setBattlePage] = useState(0);
  const [buffPage, setBuffPage] = useState(0);

  const pageSize = 6;
  const page = catalogKind === "battle" ? battlePage : buffPage;
  const battleVisibleCards = battleCards.slice(battlePage * pageSize, battlePage * pageSize + pageSize);
  const buffVisibleCards = buffCards.slice(buffPage * pageSize, buffPage * pageSize + pageSize);
  const totalCards = catalogKind === "battle" ? battleCards.length : buffCards.length;
  const pageCount = Math.max(1, Math.ceil(totalCards / pageSize));

  function setPage(nextPage: number) {
    if (catalogKind === "battle") {
      setBattlePage(nextPage);
      return;
    }
    setBuffPage(nextPage);
  }

  return (
    <section className="inventory-screen surface">
      <div className="video-stage" aria-hidden="true">
        <div className="video-stage__glow" />
        <div className="video-stage__label">shared video zone</div>
      </div>

      <div className="inventory-screen__content">
        <header className="inventory-topbar">
          <button type="button" className="top-slot inventory-back" onClick={onBack}>
            Back
          </button>
        </header>

        <section className="inventory-deck surface-shell">
          <div className="inventory-deck__slots">
            {Array.from({ length: 4 }).map((_, index) => (
              <div key={index} className="inventory-deck__slot">
                {previewDeckCards[index] ? <InventoryCard card={previewDeckCards[index]} size="deck" /> : null}
              </div>
            ))}
          </div>
          <button type="button" className="inventory-toggle" onClick={() => setShowFullDeck((value) => !value)}>
            {showFullDeck ? "Спрятать" : "Показать все"}
          </button>
          {showFullDeck ? (
            <div className="inventory-deck__all">
              {deckCards.map((card, index) => (
                <div key={`${card.template_id}:${index}`} className="inventory-deck__all-slot">
                  <InventoryCard card={card} size="deck" />
                </div>
              ))}
            </div>
          ) : null}
        </section>

        <section className="inventory-catalog surface-shell">
          <div className="inventory-catalog__tabs">
            <button
              type="button"
              className={`inventory-tab ${catalogKind === "battle" ? "inventory-tab--active" : ""}`}
              onClick={() => setCatalogKind("battle")}
            >
              Battle
            </button>
            <button
              type="button"
              className={`inventory-tab ${catalogKind === "buff" ? "inventory-tab--active" : ""}`}
              onClick={() => setCatalogKind("buff")}
            >
              Buff
            </button>
          </div>

          {catalogKind === "battle" ? (
            <div className="inventory-catalog__grid">
              {battleVisibleCards.map((card) => (
                <InventoryCard key={card.template_id} card={card} />
              ))}
            </div>
          ) : (
            <div className="inventory-catalog__empty">
              {buffVisibleCards.length > 0 ? "Buff cards layer reserved for future cards." : "Buff cards are not available yet."}
            </div>
          )}

          <div className="inventory-pager">
            <button type="button" className="inventory-pager__button" onClick={() => setPage(Math.max(0, page - 1))}>
              Prev
            </button>
            <span>
              {page + 1} / {pageCount}
            </span>
            <button
              type="button"
              className="inventory-pager__button"
              onClick={() => setPage(Math.min(pageCount - 1, page + 1))}
            >
              Next
            </button>
          </div>
        </section>

        <section className="inventory-shop surface-shell">Shop Placeholder</section>
      </div>
    </section>
  );
}

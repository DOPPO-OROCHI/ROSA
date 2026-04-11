import { useEffect, useMemo, useRef, useState } from "react";
import { request } from "../lib/api";
import { InventoryCard } from "./InventoryCard";
import type { BattleCard, BuffCard, DeckEntry } from "../types";

type Props = {
  draftDeckEntries: DeckEntry[];
  savedDeckEntries: DeckEntry[];
  battleCards: BattleCard[];
  buffCards: BuffCard[];
  onBack: () => void;
  onDraftDeckChange: (entries: DeckEntry[]) => void;
  onDeckSaved: (entries: DeckEntry[]) => void;
};

type DeckStack = {
  card: BattleCard;
  count: number;
};

function groupDeck(entries: DeckEntry[], battleCards: BattleCard[]): DeckStack[] {
  const battleMap = new Map(battleCards.map((card) => [card.template_id, card] as const));

  return entries
    .filter((entry) => entry.kind === "battle" && entry.count > 0)
    .map((entry) => {
      const card = battleMap.get(entry.template_id);
      if (!card) {
        return null;
      }

      return {
        card,
        count: entry.count,
      } satisfies DeckStack;
    })
    .filter((stack): stack is DeckStack => stack !== null);
}

export function InventoryScreen({
  draftDeckEntries,
  savedDeckEntries,
  battleCards,
  buffCards,
  onBack,
  onDraftDeckChange,
  onDeckSaved,
}: Props) {
  const [showFullDeck, setShowFullDeck] = useState(false);
  const [catalogKind, setCatalogKind] = useState<"battle" | "buff">("battle");
  const [battlePage, setBattlePage] = useState(0);
  const [buffPage, setBuffPage] = useState(0);
  const [deckSyncState, setDeckSyncState] = useState<"idle" | "syncing" | "saved" | "error">("idle");
  const [deckMessage, setDeckMessage] = useState("");
  const lastSavedDeckRef = useRef(JSON.stringify(savedDeckEntries));

  const pageSize = 6;
  const page = catalogKind === "battle" ? battlePage : buffPage;
  const battleVisibleCards = battleCards.slice(battlePage * pageSize, battlePage * pageSize + pageSize);
  const buffVisibleCards = buffCards.slice(buffPage * pageSize, buffPage * pageSize + pageSize);
  const totalCards = catalogKind === "battle" ? battleCards.length : buffCards.length;
  const pageCount = Math.max(1, Math.ceil(totalCards / pageSize));
  const groupedDeck = useMemo(() => groupDeck(draftDeckEntries, battleCards), [draftDeckEntries, battleCards]);
  const previewDeck = groupedDeck.slice(0, 4);
  const deckCardCount = useMemo(
    () => draftDeckEntries.reduce((total, entry) => total + entry.count, 0),
    [draftDeckEntries],
  );
  const deckCountMap = useMemo(
    () => new Map(groupedDeck.map((stack) => [stack.card.template_id, stack.count] as const)),
    [groupedDeck],
  );

  useEffect(() => {
    lastSavedDeckRef.current = JSON.stringify(savedDeckEntries);
  }, [savedDeckEntries]);

  useEffect(() => {
    const serialized = JSON.stringify(draftDeckEntries);
    if (serialized === lastSavedDeckRef.current) {
      setDeckSyncState("idle");
      setDeckMessage(deckCardCount === 20 ? "Deck ready" : `Deck draft: ${deckCardCount} / 20`);
      return;
    }

    if (deckCardCount !== 20) {
      setDeckSyncState("idle");
      setDeckMessage(`Deck draft: ${deckCardCount} / 20`);
      return;
    }

    let cancelled = false;

    async function saveDeck() {
      setDeckSyncState("syncing");
      setDeckMessage("Saving deck...");
      try {
        await request("/deck", {
          method: "POST",
          body: JSON.stringify({ entries: draftDeckEntries }),
        });
        if (cancelled) {
          return;
        }
        lastSavedDeckRef.current = serialized;
        onDeckSaved(draftDeckEntries);
        setDeckSyncState("saved");
        setDeckMessage("Deck saved");
      } catch {
        if (cancelled) {
          return;
        }
        setDeckSyncState("error");
        setDeckMessage("Deck save failed");
      }
    }

    void saveDeck();

    return () => {
      cancelled = true;
    };
  }, [draftDeckEntries, deckCardCount, onDeckSaved]);

  function setPage(nextPage: number) {
    if (catalogKind === "battle") {
      setBattlePage(nextPage);
      return;
    }
    setBuffPage(nextPage);
  }

  function updateDraftDeck(templateId: string, delta: number) {
    onDraftDeckChange(
      ((current: DeckEntry[]) => {
        const index = current.findIndex((entry) => entry.kind === "battle" && entry.template_id === templateId);
        if (index === -1) {
          if (delta <= 0) {
            return current;
          }
          return [...current, { kind: "battle", template_id: templateId, count: delta }];
        }

        const next = [...current];
        const target = next[index];
        const nextCount = target.count + delta;

        if (nextCount <= 0) {
          next.splice(index, 1);
          return next;
        }

        next[index] = { ...target, count: nextCount };
        return next;
      })(draftDeckEntries),
    );
  }

  function canAddCard(card: BattleCard) {
    const inDeck = deckCountMap.get(card.template_id) ?? 0;
    return deckCardCount < 20 && inDeck < card.copies && inDeck < card.max_copies;
  }

  function removeCard(templateId: string) {
    updateDraftDeck(templateId, -1);
  }

  function addCard(templateId: string) {
    const card = battleCards.find((item) => item.template_id === templateId);
    if (!card || !canAddCard(card)) {
      return;
    }
    updateDraftDeck(templateId, 1);
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
          <div className="inventory-deck__summary">
            <span className="inventory-deck__count">{deckCardCount} / 20 cards</span>
            <span className={`inventory-deck__status inventory-deck__status--${deckSyncState}`}>{deckMessage || "Deck ready"}</span>
          </div>
          <div className="inventory-deck__slots">
            {Array.from({ length: 4 }).map((_, index) => (
              <div key={index} className="inventory-deck__slot">
                {previewDeck[index] ? (
                  <InventoryCard
                    card={previewDeck[index].card}
                    size="deck"
                    countBadge={previewDeck[index].count}
                    onRemove={() => removeCard(previewDeck[index].card.template_id)}
                    description={previewDeck[index].card.description}
                  />
                ) : null}
              </div>
            ))}
          </div>
          <button type="button" className="inventory-toggle" onClick={() => setShowFullDeck(true)}>
            Показать все
          </button>
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
                <InventoryCard
                  key={card.template_id}
                  card={card}
                  description={card.description}
                  countBadge={deckCountMap.get(card.template_id) ?? null}
                  onAdd={() => addCard(card.template_id)}
                  addDisabled={!canAddCard(card)}
                />
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

      {showFullDeck ? (
        <div className="overlay">
          <section className="inventory-deck-modal surface">
            <div className="inventory-deck-modal__grid">
              {groupedDeck.map((stack) => (
                <div key={stack.card.template_id} className="inventory-deck__all-slot">
                  <InventoryCard
                    card={stack.card}
                    size="deck"
                    countBadge={stack.count}
                    onRemove={() => removeCard(stack.card.template_id)}
                    description={stack.card.description}
                  />
                </div>
              ))}
            </div>
            <button type="button" className="inventory-toggle inventory-toggle--modal" onClick={() => setShowFullDeck(false)}>
              Скрыть
            </button>
          </section>
        </div>
      ) : null}
    </section>
  );
}

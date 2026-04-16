import { useEffect, useRef, useState } from "react";
import { HandCard } from "./HandCard";
import type { BattleCardInMatch } from "./types";

type Props = {
  hand: BattleCardInMatch[];
  selectedCardId?: string;
  onPreview?: (card: BattleCardInMatch | null, originRect?: DOMRect) => void;
};

export function HandPanel({ hand, selectedCardId = "", onPreview }: Props) {
  const [panelWidth, setPanelWidth] = useState(0);
  const panelRef = useRef<HTMLElement | null>(null);
  const groupedHand = hand.reduce<Array<{ card: BattleCardInMatch; count: number }>>((acc, card) => {
    const existing = acc.find((entry) => entry.card.template_id === card.template_id);
    if (existing) {
      existing.count += 1;
      return acc;
    }

    acc.push({ card, count: 1 });
    return acc;
  }, []);
  const cardWidth = 96;
  const horizontalPadding = 20;
  const availableWidth = Math.max(0, panelWidth - horizontalPadding);

  useEffect(() => {
    const node = panelRef.current;
    if (!node) {
      return;
    }

    const updateWidth = () => {
      setPanelWidth(node.clientWidth);
    };

    updateWidth();

    const observer = new ResizeObserver(() => {
      updateWidth();
    });
    observer.observe(node);

    return () => observer.disconnect();
  }, []);

  const center = (groupedHand.length - 1) / 2;
  const spacing =
    groupedHand.length > 1
      ? Math.max(28, Math.min(cardWidth * 0.82, (availableWidth - cardWidth) / (groupedHand.length - 1)))
      : 0;

  return (
    <section
      ref={panelRef}
      className="battle-hand-panel"
    >
      <div className="battle-hand-panel__cards">
        {groupedHand.map(({ card, count }, index) => {
        const distanceFromCenter = index - center;
        const fanAngle = distanceFromCenter * (groupedHand.length > 1 ? 2.15 : 0);
        const fanDrop = Math.abs(distanceFromCenter) * 1.35;
        const offsetX = distanceFromCenter * spacing;

        return (
          <HandCard
            key={card.instance_id}
            card={card}
            selected={selectedCardId === card.instance_id}
            count={count}
            stackIndex={index}
            stacked={groupedHand.length > 4}
            fanAngle={fanAngle}
            fanDrop={fanDrop}
            offsetX={offsetX}
            onOpen={(originRect) => {
              if (selectedCardId === card.instance_id) {
                onPreview?.(null);
                return;
              }
              onPreview?.(card, originRect);
            }}
          />
        );
        })}
      </div>
      <div
        className="battle-hand-panel__hitbox"
        onClick={() => {
          onPreview?.(null);
        }}
      />
    </section>
  );
}

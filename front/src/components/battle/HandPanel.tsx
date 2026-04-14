import { useState } from "react";
import { HandCard } from "./HandCard";
import type { BattleCardInMatch } from "./types";

type Props = {
  hand: BattleCardInMatch[];
};

export function HandPanel({ hand }: Props) {
  const [selectedCardId, setSelectedCardId] = useState("");

  return (
    <section className="battle-hand-panel">
      {hand.map((card) => (
        <HandCard
          key={card.instance_id}
          card={card}
          selected={selectedCardId === card.instance_id}
        />
      ))}
      <div className="battle-hand-panel__hitbox" onClick={() => setSelectedCardId("")} />
    </section>
  );
}

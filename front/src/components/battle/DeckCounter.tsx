type Props = {
  count: number;
};

export function DeckCounter({ count }: Props) {
  return (
    <div className="battle-deck-counter">
      <span className="battle-deck-counter__label">DECK</span>
      <span className="battle-deck-counter__value">{count}</span>
    </div>
  );
}

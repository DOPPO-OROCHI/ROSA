type Props = {
  count: number;
};

export function GraveyardBlock({ count }: Props) {
  return (
    <button type="button" className="battle-graveyard" disabled>
      <span className="battle-graveyard__label">GRAVE</span>
      <span className="battle-graveyard__value">{count}</span>
    </button>
  );
}

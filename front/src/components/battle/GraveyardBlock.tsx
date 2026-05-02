type Props = {
  count: number;
  onOpen?: () => void;
};

export function GraveyardBlock({ count, onOpen }: Props) {
  return (
    <button type="button" className="battle-graveyard" onClick={onOpen} disabled={count <= 0}>
      <span className="battle-graveyard__label">GRAVE</span>
      <span className="battle-graveyard__value">{count}</span>
    </button>
  );
}

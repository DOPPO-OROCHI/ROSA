type Props = {
  disabled?: boolean;
  onEndTurn: () => void;
};

export function EndTurnButton({ disabled = false, onEndTurn }: Props) {
  return (
    <button type="button" className="battle-end-turn-button" onClick={onEndTurn} disabled={disabled}>
      ЗАКОНЧИТЬ ХОД
    </button>
  );
}

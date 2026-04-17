type Props = {
  attack: number;
  cooldown: number;
  selected?: boolean;
  disabled?: boolean;
  onClick?: () => void;
};

export function AttackBlock({ attack, cooldown, selected = false, disabled = false, onClick }: Props) {
  return (
    <button
      type="button"
      className={`battle-attack ${selected ? "battle-attack--selected" : ""}`}
      disabled={disabled}
      onClick={onClick}
    >
      <span className="battle-attack__title">ATK</span>
      <span className="battle-attack__value">{attack}</span>
      <span className="battle-attack__meta">CD {cooldown}</span>
    </button>
  );
}

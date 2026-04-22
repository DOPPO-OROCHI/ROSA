type Props = {
  cooldown: number;
  manaCost?: number;
  selected?: boolean;
  disabled?: boolean;
  onClick?: () => void;
};

export function AbilityBlock({ cooldown, manaCost = 0, selected = false, disabled = false, onClick }: Props) {
  return (
    <button
      type="button"
      className={`battle-ability ${selected ? "battle-attack--selected" : ""}`}
      disabled={disabled}
      onClick={onClick}
    >
      <span className="battle-ability__title">SKILL</span>
      <span className="battle-ability__meta">CD {cooldown}</span>
      <span className="battle-ability__meta">MP {manaCost}</span>
    </button>
  );
}

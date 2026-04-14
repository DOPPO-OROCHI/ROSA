type Props = {
  attack: number;
  cooldown: number;
};

export function AttackBlock({ attack, cooldown }: Props) {
  return (
    <div className="battle-attack">
      <span className="battle-attack__title">ATK</span>
      <span className="battle-attack__value">{attack}</span>
      <span className="battle-attack__meta">CD {cooldown}</span>
    </div>
  );
}

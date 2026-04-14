type Props = {
  cooldown: number;
  manaCost?: number;
};

export function AbilityBlock({ cooldown, manaCost = 0 }: Props) {
  return (
    <div className="battle-ability">
      <span className="battle-ability__title">SKILL</span>
      <span className="battle-ability__meta">CD {cooldown}</span>
      <span className="battle-ability__meta">MP {manaCost}</span>
    </div>
  );
}

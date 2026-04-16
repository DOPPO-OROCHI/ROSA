type FloatingNumber = {
  id: string;
  left: number;
  top: number;
  amount: number;
  kind: "damage" | "heal";
};

type Props = {
  numbers: FloatingNumber[];
};

export function BattleFloatingNumbers({ numbers }: Props) {
  if (numbers.length === 0) {
    return null;
  }

  return (
    <div className="battle-floating-numbers-layer" aria-hidden="true">
      {numbers.map((entry) => (
        <div
          key={entry.id}
          className={`battle-floating-number battle-floating-number--${entry.kind}`}
          style={{
            left: `${entry.left}px`,
            top: `${entry.top}px`,
          }}
        >
          {entry.amount}
        </div>
      ))}
    </div>
  );
}

export type { FloatingNumber };

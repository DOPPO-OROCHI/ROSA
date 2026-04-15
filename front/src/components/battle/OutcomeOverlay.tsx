type Props = {
  label: string;
  tone: "victory" | "defeat" | "draw";
};

export function OutcomeOverlay({ label, tone }: Props) {
  return (
    <div className="battle-outcome-overlay">
      <div className={`battle-outcome-panel battle-outcome-panel--${tone}`}>{label}</div>
    </div>
  );
}

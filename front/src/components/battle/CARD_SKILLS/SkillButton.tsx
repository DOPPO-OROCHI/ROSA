import type { BattleUnitState } from "../types";
import { getBoardSkillLabel } from "./utils";

type Props = {
  unit: BattleUnitState;
  active?: boolean;
  disabled?: boolean;
  onClick?: () => void;
};

export function SkillButton({ unit, active = false, disabled = false, onClick }: Props) {
  return (
    <button
      type="button"
      className={`battle-board-slot__skill-button ${active ? "battle-board-slot__skill-button--active" : ""}`}
      onClick={(event) => {
        event.stopPropagation();
        onClick?.();
      }}
      disabled={disabled || !onClick}
      aria-label={getBoardSkillLabel(unit)}
    >
      <span className="battle-board-slot__skill-label">{getBoardSkillLabel(unit)}</span>
    </button>
  );
}


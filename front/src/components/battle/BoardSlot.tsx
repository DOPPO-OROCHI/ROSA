import { resolveCardAssetVariantSrc } from "../../lib/api";
import { getBoardAttackDisplayKind, getBoardAttackDisplayValue } from "./card_attack";
import type { BattleUnitState } from "./types";

type Props = {
  unit: BattleUnitState | null;
  side: "player" | "enemy";
  playable?: boolean;
  selected?: boolean;
  attackTarget?: boolean;
  onClick?: () => void;
};

export function BoardSlot({ unit, side, playable = false, selected = false, attackTarget = false, onClick }: Props) {
  const skillLabel = unit ? (unit.cooldown > 0 ? `CD ${unit.cooldown}` : "SKILL") : "";
  const primaryValue = unit ? getBoardAttackDisplayValue(unit) : 0;
  const primaryKind = unit ? getBoardAttackDisplayKind(unit) : "attack";

  return (
    <button
      type="button"
      className={`battle-board-slot battle-board-slot--${side} ${unit ? "battle-board-slot--filled" : ""} ${playable ? "battle-board-slot--playable" : ""} ${selected ? "battle-board-slot--selected" : ""} ${attackTarget ? "battle-board-slot--attack-target" : ""}`}
      onClick={onClick}
      disabled={!onClick}
    >
      {unit ? (
        <>
          <img
            className="battle-board-slot__art"
            src={resolveCardAssetVariantSrc("battle", unit.template_id, "on_table")}
            alt={unit.template_id}
            onError={(event) => {
              const target = event.currentTarget;
              if (target.dataset.fallbackApplied === "1") {
                return;
              }
              target.dataset.fallbackApplied = "1";
              target.src = resolveCardAssetVariantSrc("battle", unit.template_id, "view");
            }}
          />
          <span className={`battle-board-slot__attack battle-board-slot__attack--${primaryKind}`}>{primaryValue}</span>
          <span className="battle-board-slot__cooldown">{unit.cooldown}</span>
          <span className="battle-board-slot__skill-label">{skillLabel}</span>
          <span className="battle-board-slot__hp">{unit.hp}</span>
        </>
      ) : playable ? (
        <span className="battle-board-slot__plus">+</span>
      ) : null}
    </button>
  );
}

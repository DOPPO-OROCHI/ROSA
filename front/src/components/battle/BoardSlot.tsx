import { resolveCardAssetVariantSrc } from "../../lib/api";
import type { BattleUnitState } from "./types";

type Props = {
  unit: BattleUnitState | null;
  side: "player" | "enemy";
};

export function BoardSlot({ unit, side }: Props) {
  return (
    <div className={`battle-board-slot battle-board-slot--${side} ${unit ? "battle-board-slot--filled" : ""}`}>
      {unit ? (
        <>
          <img
            className="battle-board-slot__art"
            src={resolveCardAssetVariantSrc("battle", unit.template_id, "view")}
            alt={unit.template_id}
          />
          <span className="battle-board-slot__attack">{unit.attack}</span>
          <span className="battle-board-slot__hp">{unit.hp}</span>
        </>
      ) : null}
    </div>
  );
}

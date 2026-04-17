import { resolveCardAssetVariantSrc, resolveHeroAssetVariantSrc } from "../../lib/api";
import type { CSSProperties } from "react";
import { getBoardAttackDisplayKind, getBoardAttackDisplayValue } from "./card_attack";
import { getBoardSkillLabel } from "./CARD_SKILLS";
import type { BattleUnitState } from "./types";

export type CardAttackAnimationState = {
  attacker: BattleUnitState;
  from: {
    left: number;
    top: number;
    width: number;
    height: number;
  };
  dx: number;
  dy: number;
};

type Props = {
  state: CardAttackAnimationState;
  onDone: () => void;
};

export function CardAttackAnimation({ state, onDone }: Props) {
  const { attacker, from, dx, dy } = state;
  const isHero = attacker.card_type === "hero";
  const skillLabel = getBoardSkillLabel(attacker);
  const primaryValue = getBoardAttackDisplayValue(attacker);
  const primaryKind = getBoardAttackDisplayKind(attacker);

  return (
    <div className="battle-card-attack-animation-layer" aria-hidden="true">
      <div
        className="battle-card-attack-animation"
        style={
          {
            left: `${from.left}px`,
            top: `${from.top}px`,
            width: `${from.width}px`,
            height: `${from.height}px`,
            "--battle-attack-dx": `${dx}px`,
            "--battle-attack-dy": `${dy}px`,
          } as CSSProperties
        }
        onAnimationEnd={onDone}
      >
        <img
          className={`battle-card-attack-animation__art ${isHero ? "battle-card-attack-animation__art--hero" : ""}`}
          src={
            isHero
              ? resolveHeroAssetVariantSrc(attacker.template_id, "battle_icon")
              : resolveCardAssetVariantSrc("battle", attacker.template_id, "on_table")
          }
          alt={attacker.template_id}
          onError={(event) => {
            const target = event.currentTarget;
            if (target.dataset.fallbackApplied === "1") {
              return;
            }
            target.dataset.fallbackApplied = "1";
            target.src =
              isHero
                ? resolveHeroAssetVariantSrc(attacker.template_id, "battle_icon")
                : resolveCardAssetVariantSrc("battle", attacker.template_id, "view");
          }}
        />
        {isHero ? null : (
          <>
            <span className={`battle-board-slot__attack battle-board-slot__attack--${primaryKind}`}>{primaryValue}</span>
            <span className="battle-board-slot__cooldown">{attacker.cooldown}</span>
            <span className="battle-board-slot__skill-label">{skillLabel}</span>
            <span className="battle-board-slot__hp">{attacker.hp}</span>
          </>
        )}
      </div>
    </div>
  );
}

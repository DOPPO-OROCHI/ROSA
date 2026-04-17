import { useEffect, useRef } from "react";
import { resolveCardAssetVariantSrc } from "../../lib/api";
import { getBoardAttackDisplayKind, getBoardAttackDisplayValue } from "./card_attack";
import { SkillButton, getBoardSkillLabel, getUnitAuraState, isUnitStunned } from "./CARD_SKILLS";
import type { BattleUnitState } from "./types";
import type { SkillTargetTone } from "./CARD_SKILLS";

type Props = {
  unit: BattleUnitState | null;
  side: "player" | "enemy";
  playable?: boolean;
  selected?: boolean;
  skillSelected?: boolean;
  attackTarget?: boolean;
  skillTarget?: boolean;
  skillTargetTone?: SkillTargetTone | null;
  animating?: boolean;
  hitToken?: number;
  actionDisabled?: boolean;
  skillDisabled?: boolean;
  onClick?: () => void;
  onSkillClick?: () => void;
};

export function BoardSlot({
  unit,
  side,
  playable = false,
  selected = false,
  skillSelected = false,
  attackTarget = false,
  skillTarget = false,
  skillTargetTone = null,
  animating = false,
  hitToken = 0,
  actionDisabled = false,
  skillDisabled = false,
  onClick,
  onSkillClick,
}: Props) {
  const skillLabel = unit ? getBoardSkillLabel(unit) : "";
  const primaryValue = unit ? getBoardAttackDisplayValue(unit) : 0;
  const primaryKind = unit ? getBoardAttackDisplayKind(unit) : "attack";
  const auraState = getUnitAuraState(unit);
  const slotRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!hitToken || !slotRef.current) {
      return;
    }

    const node = slotRef.current;
    node.classList.remove("battle-board-slot--hit");
    void node.offsetWidth;
    node.classList.add("battle-board-slot--hit");

    const timeoutId = window.setTimeout(() => {
      node.classList.remove("battle-board-slot--hit");
    }, 320);

    return () => window.clearTimeout(timeoutId);
  }, [hitToken]);

  return (
    <div
      ref={slotRef}
      className={`battle-board-slot battle-board-slot--${side} ${unit ? "battle-board-slot--filled" : ""} ${playable ? "battle-board-slot--playable" : ""} ${selected ? "battle-board-slot--selected" : ""} ${skillSelected ? "battle-board-slot--skill-selected" : ""} ${attackTarget ? "battle-board-slot--attack-target" : ""} ${skillTarget ? `battle-board-slot--skill-target battle-board-slot--skill-target-${skillTargetTone ?? "damage"}` : ""} ${animating ? "battle-board-slot--animating" : ""} ${actionDisabled ? "battle-board-slot--disabled" : ""} ${unit && isUnitStunned(unit) ? "battle-board-slot--stunned" : ""} ${auraState !== "none" ? `battle-board-slot--aura-${auraState}` : ""}`}
      data-unit-instance-id={unit?.instance_id ?? ""}
    >
      {unit ? (
        <>
          {auraState === "buff" || auraState === "both" ? <span className={`battle-board-slot__aura battle-board-slot__aura--buff ${auraState === "both" ? "battle-board-slot__aura--buff-half" : ""}`} /> : null}
          {auraState === "debuff" || auraState === "both" ? <span className={`battle-board-slot__aura battle-board-slot__aura--debuff ${auraState === "both" ? "battle-board-slot__aura--debuff-half" : ""}`} /> : null}
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
          <span className="battle-board-slot__hp">{unit.hp}</span>
          {skillLabel ? (
            onSkillClick ? (
              <SkillButton unit={unit} active={skillSelected} disabled={actionDisabled || skillDisabled} onClick={onSkillClick} />
            ) : (
              <span className="battle-board-slot__skill-label">{skillLabel}</span>
            )
          ) : null}
          <button type="button" className="battle-board-slot__tap" onClick={onClick} disabled={!onClick || actionDisabled} aria-label={unit.template_id} />
        </>
      ) : playable ? (
        <>
          <span className="battle-board-slot__plus">+</span>
          <button type="button" className="battle-board-slot__tap" onClick={onClick} disabled={!onClick} aria-label="play-card-slot" />
        </>
      ) : null}
    </div>
  );
}

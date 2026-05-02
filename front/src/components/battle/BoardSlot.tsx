import { useEffect, useRef } from "react";
import { resolveCardAssetVariantSrc } from "../../lib/api";
import { getBoardAttackDisplayKind, getBoardAttackDisplayValue } from "./card_attack";
import { SkillButton, getBoardSkillLabel, getUnitShieldState, getUnitStatusEntries, isUnitStunned } from "./CARD_SKILLS";
import type { BattleUnitState } from "./types";
import type { SkillTargetTone } from "./CARD_SKILLS";

type Props = {
  unit: BattleUnitState | null;
  side: "player" | "enemy";
  effectSourceLabels?: Record<string, string>;
  playable?: boolean;
  selected?: boolean;
  skillSelected?: boolean;
  attackTarget?: boolean;
  attackReady?: boolean;
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
  effectSourceLabels = {},
  playable = false,
  selected = false,
  skillSelected = false,
  attackTarget = false,
  attackReady = false,
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
  const shieldState = getUnitShieldState(unit);
  const statusEntries = getUnitStatusEntries(unit, effectSourceLabels).filter((entry) => entry.effectType !== "shield" && entry.effectType !== "reflect_shield");
  const slotRef = useRef<HTMLDivElement | null>(null);

  function resolveStatusIconSrc(effectType: string): string {
    switch (effectType) {
      case "attack":
        return "/assets/ui/pictures/icons/status/buff_atk.png";
      case "hp":
        return "/assets/ui/pictures/icons/status/buff_hp.png";
      case "attack_and_hp":
        return "/assets/ui/pictures/icons/status/buff_atk_hp.png";
      case "attack_cooldown":
        return "/assets/ui/pictures/icons/status/atk_cd_up.png";
      case "cooldown_up":
        return "/assets/ui/pictures/icons/status/atk_cd_down.png";
      case "skill_cooldown":
        return "/assets/ui/pictures/icons/status/skill_cd_up.png";
      case "skill_cooldown_up":
        return "/assets/ui/pictures/icons/status/skill_cd_down.png";
      case "stun":
        return "/assets/ui/pictures/icons/status/stun.png";
      case "disarm":
        return "/assets/ui/pictures/icons/status/disarm.png";
      case "silence":
        return "/assets/ui/pictures/icons/status/silence.png";
      case "damage_over_time":
        return "/assets/ui/pictures/icons/status/dot.png";
      case "vulnerable":
        return "/assets/ui/pictures/icons/status/vulnerable.png";
      case "heal_per_turn":
        return "/assets/ui/pictures/icons/status/heal_over_time.png";
      case "no_heal":
        return "/assets/ui/pictures/icons/status/no_heal.png";
      default:
        return "";
    }
  }

  const shieldIconSrc = shieldState.hasReflectShield
    ? "/assets/ui/pictures/icons/status/reflect_shield.png"
    : shieldState.hasShield
      ? "/assets/ui/pictures/icons/status/shield.png"
      : "";

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
      className={`battle-board-slot battle-board-slot--${side} ${unit ? "battle-board-slot--filled" : ""} ${playable ? "battle-board-slot--playable" : ""} ${selected ? "battle-board-slot--selected" : ""} ${skillSelected ? "battle-board-slot--skill-selected" : ""} ${attackTarget ? "battle-board-slot--attack-target" : ""} ${attackReady ? "battle-board-slot--attack-ready" : ""} ${skillTarget ? `battle-board-slot--skill-target battle-board-slot--skill-target-${skillTargetTone ?? "damage"}` : ""} ${animating ? "battle-board-slot--animating" : ""} ${actionDisabled ? "battle-board-slot--disabled" : ""} ${unit && isUnitStunned(unit) ? "battle-board-slot--stunned" : ""}`}
      data-unit-instance-id={unit?.instance_id ?? ""}
    >
      {unit ? (
        <>
          {attackReady ? <span className="battle-board-slot__ready-ring" aria-hidden="true" /> : null}
          {shieldIconSrc ? (
            <div className={`battle-board-slot__shield-badge ${shieldState.hasReflectShield ? "battle-board-slot__shield-badge--reflect" : "battle-board-slot__shield-badge--plain"}`}>
              <img src={shieldIconSrc} alt={shieldState.hasReflectShield ? "Reflect shield" : "Shield"} className="battle-board-slot__shield-icon" />
              <span className="battle-board-slot__shield-badge-value">{shieldState.label}</span>
            </div>
          ) : null}
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
          {statusEntries.length > 0 ? (
            <div className="battle-board-slot__status-strip" aria-label={`Status effects on ${unit.template_id}`}>
              {statusEntries.map((entry) => {
                const iconSrc = resolveStatusIconSrc(entry.effectType);
                if (!iconSrc) {
                  return null;
                }

                return (
                  <div key={entry.key} className={`battle-board-slot__status-icon battle-board-slot__status-icon--${entry.tone}`} title={`${entry.sourceLabel} - ${entry.label} - ${entry.valueLabel} - ${entry.turnsLabel}`}>
                    <img src={iconSrc} alt={entry.label} className="battle-board-slot__status-icon-image" />
                    <span className="battle-board-slot__status-icon-value">{entry.valueLabel}</span>
                  </div>
                );
              })}
            </div>
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

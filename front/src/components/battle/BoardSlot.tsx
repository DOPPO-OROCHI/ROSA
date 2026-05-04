import { useEffect, useRef, useState, type CSSProperties } from "react";
import { resolveCardAssetVariantSrc } from "../../lib/api";
import { getBoardAttackDisplayKind, getBoardAttackDisplayValue } from "./card_attack";
import { SkillButton, getBoardSkillLabel, getUnitShieldState, getUnitStatusEntries, isUnitStunned } from "./CARD_SKILLS";
import type { BattleUnitState } from "./types";
import type { SkillTargetTone } from "./CARD_SKILLS";

type Props = {
  unit: BattleUnitState | null;
  side: "player" | "enemy";
  effectSourceLabels?: Record<string, string>;
  cardNameByTemplateId?: Record<string, string>;
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

type StatusEntry = ReturnType<typeof getUnitStatusEntries>[number];

export function BoardSlot({
  unit,
  side,
  effectSourceLabels = {},
  cardNameByTemplateId = {},
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
  const [statusPanelOpen, setStatusPanelOpen] = useState(false);

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

  useEffect(() => {
    if (!unit || statusEntries.length === 0) {
      setStatusPanelOpen(false);
    }
  }, [statusEntries.length, unit]);

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
          {unit.is_tank ? <span className="battle-board-slot__tank-label">ТАНК</span> : null}
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
            <div className="battle-board-slot__status-layer">
              <button
                type="button"
                className={`battle-board-slot__status-bar ${statusPanelOpen ? "battle-board-slot__status-bar--open" : ""}`}
                onClick={(event) => {
                  event.stopPropagation();
                  setStatusPanelOpen(true);
                }}
              >
                <span className="battle-board-slot__status-text">ЭФФЕКТЫ</span>
                <span className="battle-board-slot__status-count">{statusEntries.length}</span>
              </button>
            </div>
          ) : null}
          {statusPanelOpen ? <StatusPanel unit={unit} unitName={cardNameByTemplateId[unit.template_id] ?? unit.template_id} entries={statusEntries} onClose={() => setStatusPanelOpen(false)} /> : null}
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

function formatStatusTurns(turnsLabel: string) {
  return turnsLabel === "INF" ? "∞" : turnsLabel;
}

function formatStatusValue(valueLabel: string) {
  return valueLabel || "—";
}

function StatusPanel({ unit, unitName, entries, onClose }: { unit: BattleUnitState; unitName: string; entries: StatusEntry[]; onClose: () => void }) {
  return (
    <div
      className="battle-status-panel"
      role="dialog"
      aria-modal="true"
      aria-label={`Effects on ${unitName}`}
      onClick={(event) => {
        event.stopPropagation();
        onClose();
      }}
    >
      <div
        className="battle-status-panel__card"
        style={{ "--battle-status-panel-bg": `url("${resolveCardAssetVariantSrc("battle", unit.template_id, "view")}")` } as CSSProperties}
        onClick={(event) => {
          event.stopPropagation();
        }}
      >
        <button type="button" className="battle-status-panel__close" onClick={onClose} aria-label="Close effects">
          ×
        </button>
        <div className="battle-status-panel__header">
        <div className="battle-status-panel__title">ЭФФЕКТЫ</div>
        <div className="battle-status-panel__subtitle">{unitName}</div>
        </div>
        <div className="battle-status-panel__list">
          {entries.map((entry) => (
            <div key={entry.key} className={`battle-status-panel__row battle-status-panel__row--${entry.tone}`}>
              <div className="battle-status-panel__effect">
                <span className="battle-status-panel__effect-name">{entry.label}</span>
                <span className="battle-status-panel__effect-source">{entry.sourceLabel}</span>
              </div>
              <span className={`battle-status-panel__value battle-status-panel__value--${entry.tone}`}>{formatStatusValue(entry.valueLabel)}</span>
              <span className="battle-status-panel__turns">{formatStatusTurns(entry.turnsLabel)}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

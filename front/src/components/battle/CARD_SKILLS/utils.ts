import type { BattleSkillState, BattleUnitEffect, BattleUnitState, MaskedBattlePlayerState } from "../types";
import type { SkillTargetTone } from "./types";

const EFFECT_STUN = "stun";
const EFFECT_SILENCE = "silence";
const EFFECT_SHIELD = "shield";
const EFFECT_REFLECT_SHIELD = "reflect_shield";

function getMaxEffectTurns(effects: BattleUnitEffect[], effectType: string): number {
  return effects.reduce((max, effect) => {
    if (effect.effect_type !== effectType) {
      return max;
    }
    return Math.max(max, effect.turns_left ?? 0);
  }, 0);
}

export function getUnitSkill(unit: BattleUnitState | null | undefined): BattleSkillState | null {
  const raw = (unit as { skill?: Record<string, unknown> } | null | undefined)?.skill;
  if (!raw) {
    return null;
  }

  return {
    name: String(raw.name ?? raw.Name ?? ""),
    code: String(raw.code ?? raw.Code ?? ""),
    kind: String(raw.kind ?? raw.Kind ?? ""),
    target: String(raw.target ?? raw.Target ?? ""),
    power: Number(raw.power ?? raw.Power ?? 0),
    base_cooldown: Number(raw.base_cooldown ?? raw.BaseCooldown ?? 0),
    cooldown_left: Number(raw.cooldown_left ?? raw.CooldownLeft ?? 0),
    duration: Number(raw.duration ?? raw.Duration ?? 0),
    extra_value: Number(raw.extra_value ?? raw.ExtraValue ?? 0),
    buff_effect: String(raw.buff_effect ?? raw.BuffEffect ?? ""),
    debuff_effect: String(raw.debuff_effect ?? raw.DebuffEffect ?? ""),
    cleanse_mode: String(raw.cleanse_mode ?? raw.CleanseMode ?? ""),
    ignore_tank: Boolean(raw.ignore_tank ?? raw.IgnoreTank ?? false),
    apply_count: Number(raw.apply_count ?? raw.ApplyCount ?? raw.hit_count ?? raw.HitCount ?? 0),
  };
}

export function getStunTurns(unit: BattleUnitState | null | undefined): number {
  return unit ? getMaxEffectTurns(unit.effects ?? [], EFFECT_STUN) : 0;
}

export function getSilenceTurns(unit: BattleUnitState | null | undefined): number {
  return unit ? getMaxEffectTurns(unit.effects ?? [], EFFECT_SILENCE) : 0;
}

export function isUnitStunned(unit: BattleUnitState | null | undefined): boolean {
  return getStunTurns(unit) > 0;
}

export function isUnitSilenced(unit: BattleUnitState | null | undefined): boolean {
  return getSilenceTurns(unit) > 0;
}

export function hasBoardSkill(unit: BattleUnitState | null | undefined): boolean {
  return Boolean(unit?.has_skill && getUnitSkill(unit));
}

export function getBoardSkillCooldown(unit: BattleUnitState | null | undefined): number {
  return getUnitSkill(unit)?.cooldown_left ?? 0;
}

export function isBoardSkillReady(unit: BattleUnitState | null | undefined): boolean {
  return hasBoardSkill(unit) && getBoardSkillCooldown(unit) <= 0 && !isUnitSilenced(unit) && !isUnitStunned(unit);
}

export function getBoardSkillLabel(unit: BattleUnitState | null | undefined): string {
  if (!unit || !unit.has_skill || !unit.skill) {
    return "";
  }

  const silenceTurns = getSilenceTurns(unit);
  if (silenceTurns > 0) {
    return `SILENCE ${silenceTurns}`;
  }

  const cooldown = getBoardSkillCooldown(unit);
  if (cooldown > 0) {
    return `CD ${cooldown}`;
  }

  return "SKILL";
}

export function getSkillTargetTone(skill: BattleSkillState | null | undefined): SkillTargetTone {
  switch (skill?.kind) {
    case "heal":
      return "heal";
    case "buff":
    case "summon":
      return "buff";
    case "debuff":
      return "debuff";
    default:
      return "damage";
  }
}

export function skillNeedsManualTarget(skill: BattleSkillState | null | undefined): boolean {
  if (!skill) {
    return false;
  }

  return skill.target === "ally_single" || skill.target === "enemy_single" || skill.target === "enemy_splash";
}

export function canSkillTargetEnemyHero(skill: BattleSkillState | null | undefined): boolean {
  if (!skill) {
    return false;
  }
  if (skill.target !== "enemy_single") {
    return false;
  }

  return skill.kind === "damage" || skill.kind === "kill" || skill.kind === "explode" || skill.kind === "hybrid";
}

export function getEnemyTankTargets(enemy: MaskedBattlePlayerState | null): BattleUnitState[] {
  if (!enemy) {
    return [];
  }

  return enemy.table.filter((unit): unit is BattleUnitState => Boolean(unit?.is_tank));
}

export function getManualSkillTargets(
  caster: BattleUnitState,
  player: MaskedBattlePlayerState | null,
  enemy: MaskedBattlePlayerState | null,
): { ids: string[]; canTargetHero: boolean } {
  const skill = getUnitSkill(caster);
  if (!skill) {
    return { ids: [], canTargetHero: false };
  }

  if (skill.target === "ally_single") {
    return {
      ids: (player?.table ?? [])
        .filter((unit): unit is BattleUnitState => Boolean(unit && unit.instance_id !== caster.instance_id))
        .map((unit) => unit.instance_id),
      canTargetHero: false,
    };
  }

  if (skill.target === "enemy_single" || skill.target === "enemy_splash") {
    const tanks = !skill.ignore_tank ? getEnemyTankTargets(enemy) : [];
    const targets = tanks.length > 0
      ? tanks
      : (enemy?.table ?? []).filter((unit): unit is BattleUnitState => Boolean(unit));

    return {
      ids: targets.map((unit) => unit.instance_id),
      canTargetHero: tanks.length === 0 && canSkillTargetEnemyHero(skill),
    };
  }

  return { ids: [], canTargetHero: false };
}

export function getUnitAuraState(unit: BattleUnitState | null | undefined): "none" | "buff" | "debuff" | "both" {
  if (!unit) {
    return "none";
  }

  const hasBuff = (unit.effects ?? []).some((effect) => effect.polarity === "buff");
  const hasDebuff = (unit.effects ?? []).some((effect) => effect.polarity === "debuff");

  if (hasBuff && hasDebuff) {
    return "both";
  }
  if (hasBuff) {
    return "buff";
  }
  if (hasDebuff) {
    return "debuff";
  }

  return "none";
}

export function getUnitShieldState(unit: BattleUnitState | null | undefined): {
  totalShield: number;
  totalReflect: number;
  hasShield: boolean;
  hasReflectShield: boolean;
  label: string;
} {
  if (!unit) {
    return {
      totalShield: 0,
      totalReflect: 0,
      hasShield: false,
      hasReflectShield: false,
      label: "",
    };
  }

  let totalShield = 0;
  let totalReflect = 0;
  let hasShield = false;
  let hasReflectShield = false;

  (unit.effects ?? []).forEach((effect) => {
    if (effect.effect_type === EFFECT_SHIELD) {
      totalShield += Math.max(0, effect.value ?? 0);
      hasShield = true;
    }

    if (effect.effect_type === EFFECT_REFLECT_SHIELD) {
      totalShield += Math.max(0, effect.value ?? 0);
      totalReflect += Math.max(0, effect.extra_value ?? 0);
      hasReflectShield = true;
    }
  });

  return {
    totalShield,
    totalReflect,
    hasShield: totalShield > 0 && (hasShield || hasReflectShield),
    hasReflectShield: totalShield > 0 && hasReflectShield,
    label: totalReflect > 0 ? `${totalShield}+${totalReflect}` : totalShield > 0 ? `${totalShield}` : "",
  };
}

function getEffectDisplayName(effectType: string): string {
  switch (effectType) {
    case "shield":
      return "Shield";
    case "reflect_shield":
      return "Reflect Shield";
    case "stun":
      return "Stun";
    case "silence":
      return "Silence";
    case "attack":
      return "ATK";
    case "attack_and_hp":
      return "ATK + HP";
    case "hp":
      return "HP";
    case "attack_cooldown":
      return "ATK CD";
    case "skill_cooldown":
      return "SKILL CD";
    case "skill_power":
      return "Skill Power";
    case "heal_per_turn":
      return "Regeneration";
    case "splash":
      return "Splash";
    case "overdrive":
      return "Overdrive";
    case "multicast":
      return "Multicast";
    case "make_tank":
      return "Tank";
    case "vampiric_strike":
      return "Vampiric Strike";
    case "chain_attack":
      return "Chain Attack";
    case "damage_reduction":
      return "Damage Reduction";
    case "death_explosion":
      return "Death Explosion";
    case "death_mass_heal":
      return "Death Mass Heal";
    case "counterattack":
      return "Counterattack";
    case "life_on_hit":
      return "Life On Hit";
    case "bonus_after_attack":
      return "Bonus After Attack";
    case "attack_down":
      return "Attack Down";
    case "cooldown_up":
      return "Cooldown Up";
    case "skill_cooldown_up":
      return "Skill Cooldown Up";
    case "damage_over_time":
      return "Damage Over Time";
    case "no_heal":
      return "No Heal";
    case "vulnerable":
      return "Vulnerable";
    case "disarm":
      return "Disarm";
    default:
      return effectType
        .split("_")
        .filter(Boolean)
        .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
        .join(" ");
  }
}

function formatEffectValue(value: number): string {
  if (value > 0) {
    return `+${value}`;
  }
  return `${value}`;
}

function getEffectValueLabel(effect: BattleUnitEffect): string {
  if (effect.effect_type === EFFECT_REFLECT_SHIELD) {
    return `${formatEffectValue(effect.value ?? 0)}/${formatEffectValue(effect.extra_value ?? 0)}`;
  }

  if (effect.effect_type === "attack_and_hp") {
    return `${formatEffectValue(effect.value ?? 0)}/${formatEffectValue(effect.extra_value ?? 0)}`;
  }

  if ((effect.extra_value ?? 0) > 0) {
    return `${formatEffectValue(effect.value ?? 0)}/${formatEffectValue(effect.extra_value ?? 0)}`;
  }

  return formatEffectValue(effect.value ?? 0);
}

export function getUnitStatusEntries(
  unit: BattleUnitState | null | undefined,
  sourceLabels: Record<string, string> = {},
): Array<{
  key: string;
  effectType: string;
  sourceLabel: string;
  label: string;
  valueLabel: string;
  turnsLabel: string;
  tone: "buff" | "debuff" | "neutral";
}> {
  if (!unit) {
    return [];
  }

  const grouped = new Map<string, {
    sourceLabel: string;
    effectType: string;
    totalValue: number;
    totalExtraValue: number;
    turns: number[];
    tone: "buff" | "debuff" | "neutral";
  }>();

  (unit.effects ?? []).forEach((effect) => {
    const sourceLabel = sourceLabels[effect.source_instance_id] ?? "Unknown Source";
    const key = `${effect.source_instance_id}:${effect.effect_type}:${effect.polarity ?? ""}`;
    const current = grouped.get(key);
    const nextTone = effect.polarity === "buff" ? "buff" : effect.polarity === "debuff" ? "debuff" : "neutral";

    if (!current) {
      grouped.set(key, {
        sourceLabel,
        effectType: effect.effect_type,
        totalValue: effect.value ?? 0,
        totalExtraValue: effect.extra_value ?? 0,
        turns: [effect.turns_left ?? 0],
        tone: nextTone,
      });
      return;
    }

    current.totalValue += effect.value ?? 0;
    current.totalExtraValue += effect.extra_value ?? 0;
    current.turns.push(effect.turns_left ?? 0);
    if (current.tone === "neutral") {
      current.tone = nextTone;
    }
  });

  return Array.from(grouped.entries())
    .map(([key, group]) => {
      const hasInfinite = group.turns.some((turn) => turn <= 0);
      const maxTurns = group.turns.reduce((max, turn) => Math.max(max, turn), 0);
      const aggregatedEffect: BattleUnitEffect = {
        effect_type: group.effectType,
        turns_left: maxTurns,
        value: group.totalValue,
        extra_value: group.totalExtraValue,
        source_type: "",
        polarity: group.tone,
        source_instance_id: "",
        dispellable: true,
        targeting: "",
      };

      return {
        key: `${unit.instance_id}-${key}`,
        effectType: group.effectType,
        sourceLabel: group.sourceLabel,
        label: getEffectDisplayName(group.effectType),
        valueLabel: getEffectValueLabel(aggregatedEffect),
        turnsLabel: hasInfinite ? "INF" : `${maxTurns}T`,
        tone: group.tone,
      };
    })
    .sort((left, right) => left.sourceLabel.localeCompare(right.sourceLabel) || left.label.localeCompare(right.label));
}

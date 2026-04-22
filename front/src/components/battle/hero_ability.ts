import { useEffect, useMemo, useState } from "react";
import type { Hero } from "../../types";
import type { BattleUnitState, MaskedBattlePlayerState } from "./types";
import type { SkillTargetTone } from "./CARD_SKILLS";

type HeroAbilityMeta = {
  kind: string;
  target: string;
  ignoreTank?: boolean;
};

type Params = {
  player: MaskedBattlePlayerState | null;
  enemy: MaskedBattlePlayerState | null;
  playerHero: Hero | null;
  isPlayerTurn: boolean;
  busy: boolean;
  finished: boolean;
  onInfo: (message: string) => void;
  onCast: (target: BattleUnitState | null, attackHero: boolean) => Promise<void> | void;
};

function getHeroAbilityMeta(playerHero: Hero | null | undefined): HeroAbilityMeta | null {
  if (!playerHero?.ability?.code) {
    return null;
  }
  return {
    kind: playerHero.ability.kind,
    target: playerHero.ability.target,
    ignoreTank: playerHero.ability.ignore_tank,
  };
}

function getHeroAbilityTone(meta: HeroAbilityMeta | null): SkillTargetTone | null {
  if (!meta) {
    return null;
  }
  switch (meta.kind) {
    case "buff":
      return "buff";
    case "debuff":
      return "debuff";
    default:
      return "damage";
  }
}

function getEnemyTankTargets(enemy: MaskedBattlePlayerState | null): BattleUnitState[] {
  if (!enemy) {
    return [];
  }
  return enemy.table.filter((unit): unit is BattleUnitState => Boolean(unit?.is_tank));
}

function getHeroAbilityTargets(
  meta: HeroAbilityMeta,
  player: MaskedBattlePlayerState | null,
  enemy: MaskedBattlePlayerState | null,
): { ids: string[]; canTargetHero: boolean; needsManualTarget: boolean } {
  switch (meta.target) {
    case "ally_single":
    case "ally_splash":
      return {
        ids: (player?.table ?? []).filter((unit): unit is BattleUnitState => Boolean(unit)).map((unit) => unit.instance_id),
        canTargetHero: false,
        needsManualTarget: true,
      };
    case "ally_all":
      return { ids: [], canTargetHero: false, needsManualTarget: false };
    case "enemy_single":
    case "enemy_splash": {
      const tanks = !meta.ignoreTank ? getEnemyTankTargets(enemy) : [];
      const targets = tanks.length > 0
        ? tanks
        : (enemy?.table ?? []).filter((unit): unit is BattleUnitState => Boolean(unit));
      return {
        ids: targets.map((unit) => unit.instance_id),
        canTargetHero: tanks.length === 0 && meta.kind === "damage",
        needsManualTarget: true,
      };
    }
    case "enemy_all":
      return { ids: [], canTargetHero: false, needsManualTarget: false };
    default:
      return { ids: [], canTargetHero: false, needsManualTarget: false };
  }
}

export function useHeroAbility({ player, enemy, playerHero, isPlayerTurn, busy, finished, onInfo, onCast }: Params) {
  const [selected, setSelected] = useState(false);

  const meta = useMemo(() => getHeroAbilityMeta(playerHero), [playerHero]);
  const targetTone = useMemo(() => getHeroAbilityTone(meta), [meta]);

  const state = useMemo(() => {
    if (!meta) {
      return { ids: [], canTargetHero: false, needsManualTarget: false };
    }
    return getHeroAbilityTargets(meta, player, enemy);
  }, [enemy, meta, player]);

  useEffect(() => {
    if (!selected) {
      return;
    }
    if (!isPlayerTurn || busy || finished) {
      setSelected(false);
    }
  }, [busy, finished, isPlayerTurn, selected]);

  function clearSelection() {
    setSelected(false);
  }

  async function cast(target: BattleUnitState | null, attackHero: boolean) {
    await onCast(target, attackHero);
    clearSelection();
  }

  async function toggleSelection() {
    if (!player || !meta || !isPlayerTurn || busy || finished) {
      onInfo("СКИЛЛ НЕДОСТУПЕН");
      return;
    }
    if ((player.hero_ability_cooldown ?? 0) > 0) {
      onInfo(`КД СКИЛЛА - ${player.hero_ability_cooldown}`);
      return;
    }
    if ((player.hero_ability_mana_cost ?? 0) > player.mana) {
      onInfo("НЕ ХВАТАЕТ МАНЫ");
      return;
    }
    if (!state.needsManualTarget) {
      await cast(null, false);
      return;
    }
    if (state.ids.length === 0 && !state.canTargetHero) {
      onInfo("НЕТ ДОСТУПНЫХ ЦЕЛЕЙ");
      return;
    }
    setSelected((current) => !current);
  }

  async function tryCastOnUnit(unit: BattleUnitState) {
    if (!selected || !state.ids.includes(unit.instance_id)) {
      clearSelection();
      return;
    }
    await cast(unit, false);
  }

  async function tryCastOnHero() {
    if (!selected || !state.canTargetHero) {
      clearSelection();
      return;
    }
    await cast(null, true);
  }

  const infoMessage = useMemo(() => {
    if (!selected) {
      return "";
    }
    if ((player?.hero_ability_cooldown ?? 0) > 0) {
      return `КД СКИЛЛА - ${player?.hero_ability_cooldown ?? 0}`;
    }
    return "";
  }, [player?.hero_ability_cooldown, selected]);

  return {
    selected,
    meta,
    targetTone,
    targetIds: state.ids,
    canTargetHero: state.canTargetHero,
    infoMessage,
    clearSelection,
    toggleSelection,
    tryCastOnUnit,
    tryCastOnHero,
  };
}

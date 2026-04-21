import type { BattleEvent } from "../types";
import { getUnitSkill } from "../CARD_SKILLS";
import type { ProjectileRuleResolver, ProjectileTone } from "./types";

function toneFromSkillKind(kind: string | undefined): ProjectileTone {
  switch (kind) {
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

export const resolveProjectileRule: ProjectileRuleResolver = (skill, event: BattleEvent) => {
  const tone = toneFromSkillKind(skill?.kind);
  const target = skill?.target ?? "";
  const skillCode = skill?.code ?? "";

  if (target === "self" || target === "ally_adjacent" || target === "ally_all" || target === "enemy_all") {
    return { mode: "none", tone };
  }

  if (target === "enemy_splash") {
    return { mode: "splash", tone };
  }

  if (target === "enemy_random_multi" || skillCode.includes("multi")) {
    return { mode: "sequence", tone };
  }

  return { mode: "single", tone };
};

export { getUnitSkill };

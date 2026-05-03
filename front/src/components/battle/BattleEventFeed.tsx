import { useEffect, useRef, useState, type CSSProperties } from "react";
import { resolveCardAssetVariantSrc, resolveHeroAssetVariantSrc, resolveImageSrc } from "../../lib/api";
import type { Hero } from "../../types";
import type { BattleCardInMatch, BattleEvent, BattleEventTarget, BattleUnitState, MaskedBattleMatchState } from "./types";
import "./battle_event_feed.css";

const EVENT_LIFETIME_MS = 2500;
const ACTION_ATTACK_ICON = "/assets/ui/pictures/icons/status/disarm.png";
const ACTION_SKILL_FALLBACK_ICON = "/assets/ui/pictures/icons/status/skill_cd_down.png";
const ACTION_BUFF_ICON = "/assets/ui/pictures/icons/status/buff_atk_hp.png";
const ACTION_HEAL_ICON = "/assets/ui/pictures/icons/status/buff_hp.png";
const ACTION_DEATH_ICON = "/assets/ui/pictures/icons/status/dot.png";
const ACTION_SUMMON_ICON = "/assets/ui/pictures/icons/status/shield.png";

type FeedEntity = {
  id: string;
  name: string;
  src: string;
  kind: "card" | "hero";
};

type FeedItem = {
  id: string;
  type: string;
  source: FeedEntity | null;
  targets: FeedEntity[];
  actionSrc: string;
  actionLabel: string;
};

type Props = {
  match: MaskedBattleMatchState | null;
  heroes: Hero[];
};

function getEventSequenceId(event: BattleEvent, fallback: string) {
  return event.id ?? event.event_id ?? fallback;
}

function formatCodeName(value: string) {
  return value
    .split(/[_:\-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function findUnit(match: MaskedBattleMatchState, instanceId?: string): BattleUnitState | null {
  if (!instanceId) {
    return null;
  }

  for (const player of match.players) {
    for (const unit of player?.table ?? []) {
      if (unit?.instance_id === instanceId) {
        return unit;
      }
    }
  }

  return null;
}

function findCard(match: MaskedBattleMatchState, instanceId?: string): BattleCardInMatch | null {
  if (!instanceId) {
    return null;
  }

  for (const player of match.players) {
    const zones = [player?.hand, player?.discard, player?.graveyard, player?.deck];
    for (const zone of zones) {
      const card = zone?.find((entry) => entry.instance_id === instanceId);
      if (card) {
        return card;
      }
    }
  }

  return null;
}

function getHeroCodeByInstanceId(match: MaskedBattleMatchState, instanceId?: string) {
  const matchResult = instanceId?.match(/^hero:p([01])$/);
  if (!matchResult) {
    return "";
  }

  const playerIndex = Number(matchResult[1]);
  return match.players[playerIndex]?.hero_code ?? "";
}

function makeCardEntity(templateId: string, name?: string, kind: "battle" | "buff" = "battle"): FeedEntity | null {
  if (!templateId) {
    return null;
  }

  return {
    id: `card:${templateId}`,
    kind: "card",
    name: name || formatCodeName(templateId),
    src: resolveCardAssetVariantSrc(kind, templateId, "view"),
  };
}

function makeHeroEntity(heroCode: string, heroes: Hero[]): FeedEntity | null {
  if (!heroCode) {
    return null;
  }

  return {
    id: `hero:${heroCode}`,
    kind: "hero",
    name: heroes.find((hero) => hero.hero_code === heroCode)?.name || formatCodeName(heroCode),
    src: resolveHeroAssetVariantSrc(heroCode, "battle_icon"),
  };
}

function resolveSourceEntity(match: MaskedBattleMatchState, heroes: Hero[], event: BattleEvent): FeedEntity | null {
  const sourceHeroCode =
    event.source_hero_code ||
    ((event.source_kind === "hero" || event.source_instance_id?.startsWith("hero:"))
      ? getHeroCodeByInstanceId(match, event.source_instance_id)
      : "");

  if (sourceHeroCode) {
    return makeHeroEntity(sourceHeroCode, heroes);
  }

  const sourceUnit = findUnit(match, event.source_instance_id);
  if (sourceUnit) {
    return makeCardEntity(sourceUnit.template_id);
  }

  const sourceCard = findCard(match, event.source_instance_id);
  if (sourceCard) {
    return makeCardEntity(sourceCard.template_id, sourceCard.name, sourceCard.kind as "battle" | "buff");
  }

  if (event.source_card_template_id && event.source_kind === "card") {
    return makeCardEntity(event.source_card_template_id, undefined, "buff");
  }

  return makeCardEntity(event.source_template_id || event.source_card_template_id || "");
}

function resolveTargetEntity(match: MaskedBattleMatchState, heroes: Hero[], target: BattleEventTarget): FeedEntity | null {
  const targetHeroCode = getHeroCodeByInstanceId(match, target.instance_id);
  if (targetHeroCode) {
    return makeHeroEntity(targetHeroCode, heroes);
  }

  const unit = findUnit(match, target.instance_id);
  if (unit) {
    return makeCardEntity(unit.template_id);
  }

  const card = findCard(match, target.instance_id);
  if (card) {
    return makeCardEntity(card.template_id, card.name, card.kind as "battle" | "buff");
  }

  return makeCardEntity(target.template_id ?? "");
}

function resolveEffectIconSrc(effectType?: string) {
  switch (effectType) {
    case "attack":
    case "attack_down":
    case "skill_power":
    case "overdrive":
    case "vampiric_strike":
    case "chain_attack":
    case "bonus_after_attack":
      return "/assets/ui/pictures/icons/status/buff_atk.png";
    case "hp":
    case "heal_per_turn":
    case "life_on_hit":
    case "death_mass_heal":
    case "set_fixed_hp":
      return "/assets/ui/pictures/icons/status/buff_hp.png";
    case "attack_and_hp":
      return "/assets/ui/pictures/icons/status/buff_atk_hp.png";
    case "attack_cooldown":
      return "/assets/ui/pictures/icons/status/atk_cd_down.png";
    case "cooldown_up":
      return "/assets/ui/pictures/icons/status/atk_cd_up.png";
    case "skill_cooldown":
      return "/assets/ui/pictures/icons/status/skill_cd_down.png";
    case "skill_cooldown_up":
      return "/assets/ui/pictures/icons/status/skill_cd_up.png";
    case "damage_over_time":
    case "death_explosion":
      return "/assets/ui/pictures/icons/status/dot.png";
    case "no_heal":
      return "/assets/ui/pictures/icons/status/no_heal.png";
    case "reflect_shield":
      return "/assets/ui/pictures/icons/status/reflect_shield.png";
    case "shield":
    case "damage_reduction":
    case "make_tank":
      return "/assets/ui/pictures/icons/status/shield.png";
    case "silence":
      return "/assets/ui/pictures/icons/status/silence.png";
    case "stun":
      return "/assets/ui/pictures/icons/status/stun.png";
    case "vulnerable":
      return "/assets/ui/pictures/icons/status/vulnerable.png";
    case "disarm":
    case "counterattack":
      return "/assets/ui/pictures/icons/status/disarm.png";
    case "multicast":
    case "splash":
      return "/assets/ui/pictures/icons/status/buff_atk_hp.png";
    default:
      return "";
  }
}

function resolveTargetEffectType(match: MaskedBattleMatchState, event: BattleEvent) {
  for (const target of event.targets ?? []) {
    const unit = findUnit(match, target.instance_id);
    if (!unit?.effects?.length) {
      continue;
    }

    const sourceEffect = event.source_instance_id
      ? [...unit.effects].reverse().find((effect) => effect.source_instance_id === event.source_instance_id)
      : null;
    const effect = sourceEffect ?? unit.effects[unit.effects.length - 1];
    if (effect?.effect_type) {
      return effect.effect_type;
    }
  }

  return "";
}

function resolveHeroAbilityEffectType(match: MaskedBattleMatchState, heroes: Hero[], event: BattleEvent) {
  const heroCode =
    event.source_hero_code ||
    ((event.source_kind === "hero" || event.source_instance_id?.startsWith("hero:"))
      ? getHeroCodeByInstanceId(match, event.source_instance_id)
      : "");
  const ability = heroes.find((hero) => hero.hero_code === heroCode)?.ability;
  if (!ability) {
    return "";
  }

  if (ability.kind === "heal") {
    return "hp";
  }

  return ability.kind === "debuff" ? ability.debuff_effect : ability.buff_effect;
}

function resolveSkillEffectType(match: MaskedBattleMatchState, heroes: Hero[], event: BattleEvent) {
  const sourceUnit = findUnit(match, event.source_instance_id);
  if (sourceUnit?.skill) {
    if (sourceUnit.skill.kind === "heal") {
      return "hp";
    }
    if (sourceUnit.skill.kind === "debuff") {
      return sourceUnit.skill.debuff_effect;
    }
    if (sourceUnit.skill.kind === "buff" || sourceUnit.skill.kind === "hybrid") {
      return sourceUnit.skill.buff_effect || sourceUnit.skill.debuff_effect;
    }
  }

  if (event.type === "hero_spell") {
    return resolveHeroAbilityEffectType(match, heroes, event);
  }

  return resolveTargetEffectType(match, event);
}

function getActionMeta(match: MaskedBattleMatchState, heroes: Hero[], event: BattleEvent) {
  const sourceUnit = findUnit(match, event.source_instance_id);
  const effectKind = event.effect_kind?.toLowerCase() ?? "";
  const effectIconSrc = resolveEffectIconSrc(resolveSkillEffectType(match, heroes, event));

  switch (event.type) {
    case "attack":
    case "hero_attack":
      return { src: ACTION_ATTACK_ICON, label: "Атака" };
    case "passive":
      if (effectIconSrc && effectKind !== "heal") {
        return { src: effectIconSrc, label: "Passive" };
      }
      if (effectKind === "heal") {
        return { src: ACTION_HEAL_ICON, label: "Лечение" };
      }
      return {
        src: resolveImageSrc(sourceUnit?.skill_image_key, ACTION_SKILL_FALLBACK_ICON),
        label: "Пассивка",
      };
    case "card_skill":
    case "hero_spell":
      if (effectIconSrc) {
        return { src: effectIconSrc, label: "Skill" };
      }
      return {
        src: resolveImageSrc(sourceUnit?.skill_image_key, ACTION_SKILL_FALLBACK_ICON),
        label: "Навык",
      };
    case "heal":
      return { src: ACTION_HEAL_ICON, label: "Лечение" };
    case "buff":
      if (effectIconSrc) {
        return { src: effectIconSrc, label: "Buff" };
      }
      return { src: ACTION_BUFF_ICON, label: "Buff" };
    case "summon":
      return { src: ACTION_SUMMON_ICON, label: "Summon" };
    case "resurrect":
      return { src: ACTION_HEAL_ICON, label: "Resurrect" };
    case "legacy_buff":
      return { src: ACTION_BUFF_ICON, label: "Усиление" };
    case "death":
      return { src: ACTION_DEATH_ICON, label: "Падение" };
    default:
      return { src: ACTION_SKILL_FALLBACK_ICON, label: "Событие" };
  }
}

function createFeedItems(match: MaskedBattleMatchState, heroes: Hero[], event: BattleEvent, eventId: string): FeedItem[] {
  if (event.type === "turn") {
    return [];
  }

  const source = resolveSourceEntity(match, heroes, event);
  const targets = (event.targets ?? [])
    .map((target) => resolveTargetEntity(match, heroes, target))
    .filter((target): target is FeedEntity => Boolean(target));

  if (!source && targets.length === 0) {
    return [];
  }

  const action = getActionMeta(match, heroes, event);
  return [
    {
      id: eventId,
      type: event.type,
      source,
      targets,
      actionSrc: action.src,
      actionLabel: action.label,
    },
  ];
}

export function BattleEventFeed({ match, heroes }: Props) {
  const [items, setItems] = useState<FeedItem[]>([]);
  const processedEventIdsRef = useRef(new Set<string>());
  const timeoutIdsRef = useRef<number[]>([]);

  useEffect(() => {
    return () => {
      timeoutIdsRef.current.forEach((timeoutId) => window.clearTimeout(timeoutId));
      timeoutIdsRef.current = [];
    };
  }, []);

  useEffect(() => {
    if (!match?.events?.length) {
      return;
    }

    const nextItems: FeedItem[] = [];

    match.events.forEach((event, eventIndex) => {
      const targetFingerprint = (event.targets ?? [])
        .map((target) => `${target.instance_id ?? "none"}:${target.template_id ?? "none"}:${target.amount ?? 0}:${target.new_hp ?? "na"}:${target.died ? 1 : 0}`)
        .join("|");
      const eventId = getEventSequenceId(
        event,
        `${match.version}-${eventIndex}-${event.type}-${event.source_instance_id ?? "none"}-${event.source_template_id ?? "none"}-${targetFingerprint}`,
      );

      if (processedEventIdsRef.current.has(eventId)) {
        return;
      }
      processedEventIdsRef.current.add(eventId);
      nextItems.push(...createFeedItems(match, heroes, event, eventId));
    });

    if (nextItems.length === 0) {
      return;
    }

    setItems((current) => [...nextItems, ...current].slice(0, 8));

    nextItems.forEach((item) => {
      const timeoutId = window.setTimeout(() => {
        setItems((current) => current.filter((entry) => entry.id !== item.id));
      }, EVENT_LIFETIME_MS);
      timeoutIdsRef.current.push(timeoutId);
    });
  }, [heroes, match]);

  if (items.length === 0) {
    return null;
  }

  return (
    <div className="battle-event-feed" aria-live="polite" aria-label="Battle events">
      {items.map((item, index) => (
        <div
          key={item.id}
          className={`battle-event-feed__item battle-event-feed__item--${item.type}`}
          style={{ "--battle-event-index": index } as CSSProperties}
        >
          {item.source ? <EventPortrait entity={item.source} side="source" /> : <span className="battle-event-feed__portrait battle-event-feed__portrait--empty" />}
          <span className="battle-event-feed__action" title={item.actionLabel}>
            <img src={item.actionSrc} alt={item.actionLabel} />
          </span>
          <span className="battle-event-feed__targets">
            {item.targets.length > 0 ? (
              item.targets.map((target) => <EventPortrait key={`${item.id}-${target.id}`} entity={target} side="target" />)
            ) : (
              <span className="battle-event-feed__portrait battle-event-feed__portrait--empty" />
            )}
          </span>
        </div>
      ))}
    </div>
  );
}

function EventPortrait({ entity, side }: { entity: FeedEntity; side: "source" | "target" }) {
  return (
    <span className={`battle-event-feed__portrait battle-event-feed__portrait--${entity.kind} battle-event-feed__portrait--${side}`}>
      <img
        src={entity.src}
        alt={entity.name}
        onError={(event) => {
          const target = event.currentTarget;
          if (target.dataset.fallbackApplied === "1") {
            return;
          }
          target.dataset.fallbackApplied = "1";
          target.src = entity.kind === "hero" ? "/assets/placeholders/hero_image.svg" : "/assets/placeholders/card_image.svg";
        }}
      />
    </span>
  );
}

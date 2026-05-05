import { useEffect, useMemo, useRef, useState } from "react";
import { request, withDevSessionToken } from "../../lib/api";
import { AbilityBlock } from "./AbilityBlock";
import { AttackBlock } from "./AttackBlock";
import { getStunTurns, useCardSkill } from "./CARD_SKILLS";
import { CardAttackAnimation, type CardAttackAnimationState } from "./CardAttackAnimation";
import { BattleCardViewer, type BattleCardViewerOrigin } from "./BattleCardViewer";
import { BattleField } from "./BattleField";
import { BattleEventFeed } from "./BattleEventFeed";
import { BattleFloatingNumbers, type FloatingNumber } from "./BattleFloatingNumbers";
import { BattleGraveyardModal } from "./BattleGraveyardModal";
import { BattleInfoToast } from "./BattleInfoToast";
import { BattleLoadingScreen } from "./BattleLoadingScreen";
import { DeckCounter } from "./DeckCounter";
import { Defeat } from "./Defeat";
import { Draw } from "./Draw";
import { EnemyCharacter } from "./EnemyCharacter";
import { GamerCharacter } from "./GamerCharacter";
import { GraveyardBlock } from "./GraveyardBlock";
import { HandPanel } from "./HandPanel";
import { useHeroAbility } from "./hero_ability";
import { LeaveMatchButton } from "./LeaveMatchButton";
import { useBattlePreload } from "./battle_preload";
import { canUnitAttackNow, useCardAttack } from "./card_attack";
import { BattleDeathAnimations, collectDeathAnimations, type DeathAnimationState } from "./on_death_effects";
import { useOnHitEffects } from "./on_hit_effects";
import { BattlePlayAnimations, useOnPlayEffects } from "./on_play_effects";
import { ProjectileLayer, useProjectileRuntime } from "./projectiles";
import { useBattleEventSfx } from "./sfx";
import { Victory } from "./Victory";
import type { ApplyBattleActionRequest, BattleCardInMatch, BattleEvent, BattleUnitState, MaskedBattleMatchState, MaskedBattlePlayerState } from "./types";
import type { BattleCard, DeckEntry, Hero } from "../../types";
import "./battle.css";

type GraveyardResponse = {
  cards: BattleCardInMatch[];
  count: number;
};

type Props = {
  currentUserId: number;
  matchId: number;
  heroes: Hero[];
  battleCards: BattleCard[];
  deckEntries: DeckEntry[];
  onLeaveToMenu: () => void;
};

const TURN_AURA_EXIT_MS = 620;

type TurnAuraItem = {
  id: string;
  instanceId: string;
  centerX: number;
  centerY: number;
  width: number;
  height: number;
  phase: "active" | "leaving";
};

export function BattleScreen({ currentUserId, matchId, heroes, battleCards, deckEntries, onLeaveToMenu }: Props) {
  const [match, setMatch] = useState<MaskedBattleMatchState | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [outcome, setOutcome] = useState<"victory" | "defeat" | "draw" | null>(null);
  const [previewCard, setPreviewCard] = useState<BattleCardInMatch | null>(null);
  const [previewOrigin, setPreviewOrigin] = useState<BattleCardViewerOrigin | null>(null);
  const [previewClosing, setPreviewClosing] = useState(false);
  const [graveyardOpen, setGraveyardOpen] = useState(false);
  const [graveyardCards, setGraveyardCards] = useState<BattleCardInMatch[]>([]);
  const [graveyardLoading, setGraveyardLoading] = useState(false);
  const [graveyardError, setGraveyardError] = useState("");
  const [attackAnimation, setAttackAnimation] = useState<CardAttackAnimationState | null>(null);
  const [deathAnimations, setDeathAnimations] = useState<DeathAnimationState[]>([]);
  const [floatingNumbers, setFloatingNumbers] = useState<FloatingNumber[]>([]);
  const battleCardNameByTemplateId = useMemo(
    () => Object.fromEntries(battleCards.map((card) => [card.template_id, card.name] as const)),
    [battleCards],
  );
  const [heroAttackSelected, setHeroAttackSelected] = useState(false);
  const [toastMessage, setToastMessage] = useState("");
  const [toastNonce, setToastNonce] = useState(0);
  const [turnAuras, setTurnAuras] = useState<TurnAuraItem[]>([]);
  const shellRef = useRef<HTMLDivElement | null>(null);
  const turnAuraExitTimeoutsRef = useRef<number[]>([]);
  const lastProcessedVersionRef = useRef(0);
  const currentMatchRef = useRef<MaskedBattleMatchState | null>(null);
  const streamConnectedRef = useRef(false);
  const streamHadStateRef = useRef(false);
  const lastStreamStateAtRef = useRef(0);
  const attackAnimationQueueRef = useRef(Promise.resolve());
  const processedAttackEventIdsRef = useRef(new Set<string>());
  const readyRequestSentForMatchRef = useRef<number | null>(null);

  const playerIndex = useMemo(() => {
    if (!match) {
      return -1;
    }
    return match.players.findIndex((player) => player?.user_id === currentUserId);
  }, [currentUserId, match]);

  const player = playerIndex >= 0 ? match?.players[playerIndex] ?? null : null;
  const enemy = playerIndex >= 0 ? match?.players[playerIndex === 0 ? 1 : 0] ?? null : null;
  const playerGraveyardCount = player?.graveyard_count ?? player?.graveyard?.length ?? 0;
  const preload = useBattlePreload(match, deckEntries);
  const playerReady = playerIndex >= 0 ? Boolean(match?.loading_ready?.[playerIndex]) : false;
  const enemyReady = playerIndex >= 0 ? Boolean(match?.loading_ready?.[playerIndex === 0 ? 1 : 0]) : false;
  const playerHeroInstanceId = playerIndex >= 0 ? `hero:p${playerIndex}` : "";
  const enemyHeroInstanceId = playerIndex >= 0 ? `hero:p${playerIndex === 0 ? 1 : 0}` : "";
  const playerHero = player ? heroes.find((hero) => hero.hero_code === player.hero_code) ?? null : null;
  const enemyHero = enemy ? heroes.find((hero) => hero.hero_code === enemy.hero_code) ?? null : null;
  const isPlayerTurn = Boolean(match) && playerIndex >= 0 && match?.active_player === playerIndex;
  const activeTurnLabel = isPlayerTurn ? "ВАШ ХОД" : "ХОД ПРОТИВНИКА";
  const turnNumber = Math.max(player?.turns ?? 0, enemy?.turns ?? 0, 1);
  const canEndTurn =
    Boolean(match) && playerIndex === match?.active_player && match?.phase === "MAIN" && !match?.finished && !busy;
  const canPlaySelectedBattleCard =
    Boolean(match) &&
    playerIndex === match?.active_player &&
    match?.phase === "MAIN" &&
    !match?.finished &&
    !busy &&
    previewCard?.kind === "battle" &&
    (player?.mana ?? 0) >= (previewCard?.mana_cost ?? 0);
  const canHeroAttack =
    Boolean(match) &&
    Boolean(player) &&
    playerIndex === match?.active_player &&
    match?.phase === "MAIN" &&
    !match?.finished &&
    !busy;
  const shellWidth = shellRef.current?.clientWidth ?? 0;
  const shellHeight = shellRef.current?.clientHeight ?? 0;
  const heroAttackAnimation =
    attackAnimation?.attacker.card_type === "hero" && attackAnimation.attacker.instance_id === playerHeroInstanceId
      ? { dx: attackAnimation.dx, dy: attackAnimation.dy }
      : null;
  const enemyHeroAttackAnimation =
    attackAnimation?.attacker.card_type === "hero" && attackAnimation.attacker.instance_id === enemyHeroInstanceId
      ? { dx: attackAnimation.dx, dy: attackAnimation.dy }
      : null;

  function getUnitRect(instanceId: string) {
    const shell = shellRef.current;
    if (!shell || !instanceId) {
      return null;
    }

    const shellRect = shell.getBoundingClientRect();
    const node = shell.querySelector<HTMLElement>(`[data-unit-instance-id="${instanceId}"]`);
    if (!node) {
      return null;
    }

    const rect = node.getBoundingClientRect();
    return {
      left: rect.left - shellRect.left,
      top: rect.top - shellRect.top,
      width: rect.width,
      height: rect.height,
      centerX: rect.left - shellRect.left + rect.width / 2,
      centerY: rect.top - shellRect.top + rect.height / 2,
    };
  }

  function getHeroRect(heroInstanceId: string) {
    const shell = shellRef.current;
    if (!shell || !heroInstanceId) {
      return null;
    }

    const heroSide = heroInstanceId === playerHeroInstanceId ? "player" : "enemy";
    const node = shell.querySelector<HTMLElement>(`[data-hero-side="${heroSide}"]`);
    if (!node) {
      return null;
    }

    const shellRect = shell.getBoundingClientRect();
    const rect = node.getBoundingClientRect();
    return {
      left: rect.left - shellRect.left,
      top: rect.top - shellRect.top,
      width: rect.width,
      height: rect.height,
      centerX: rect.left - shellRect.left + rect.width / 2,
      centerY: rect.top - shellRect.top + rect.height / 2,
    };
  }

  function getCombatantRect(instanceId: string) {
    return instanceId.startsWith("hero:") ? getHeroRect(instanceId) : getUnitRect(instanceId);
  }

  function scheduleTurnAuraRemoval(id: string) {
    const timeoutId = window.setTimeout(() => {
      setTurnAuras((current) => current.filter((item) => item.id !== id));
    }, TURN_AURA_EXIT_MS);
    turnAuraExitTimeoutsRef.current.push(timeoutId);
  }

  function retireActiveTurnAuras() {
    setTurnAuras((current) =>
      current.map((item) => {
        if (item.phase === "leaving") {
          return item;
        }

        const leavingItem: TurnAuraItem = {
          ...item,
          id: `${item.id}-leaving-${Date.now()}`,
          phase: "leaving",
        };
        scheduleTurnAuraRemoval(leavingItem.id);
        return leavingItem;
      }),
    );
  }

  useEffect(() => {
    return () => {
      turnAuraExitTimeoutsRef.current.forEach((timeoutId) => window.clearTimeout(timeoutId));
      turnAuraExitTimeoutsRef.current = [];
    };
  }, []);

  useEffect(() => {
    if (!match || playerIndex < 0 || preload.visible) {
      retireActiveTurnAuras();
      return;
    }

    const shell = shellRef.current;
    if (!shell) {
      retireActiveTurnAuras();
      return;
    }

    const activeHeroInstanceId = match.active_player === playerIndex ? playerHeroInstanceId : enemyHeroInstanceId;

    function measureAura() {
      const rect = getUnitRect(activeHeroInstanceId);
      if (!rect) {
        retireActiveTurnAuras();
        return;
      }

      const nextActiveAura: TurnAuraItem = {
        id: `active-${activeHeroInstanceId}`,
        instanceId: activeHeroInstanceId,
        centerX: rect.centerX,
        centerY: rect.centerY,
        width: rect.width * 3.8,
        height: rect.height * 1.18,
        phase: "active",
      };

      setTurnAuras((current) => {
        let activeAuraUpdated = false;
        const nextItems = current.map((item) => {
          if (item.phase === "leaving") {
            return item;
          }

          if (item.instanceId === activeHeroInstanceId) {
            activeAuraUpdated = true;
            return {
              ...nextActiveAura,
              id: item.id,
            };
          }

          const leavingItem: TurnAuraItem = {
            ...item,
            id: `${item.id}-leaving-${Date.now()}`,
            phase: "leaving",
          };
          scheduleTurnAuraRemoval(leavingItem.id);
          return leavingItem;
        });

        return activeAuraUpdated ? nextItems : [...nextItems, nextActiveAura];
      });
    }

    const frameId = window.requestAnimationFrame(measureAura);
    const handleResize = () => measureAura();
    window.addEventListener("resize", handleResize);

    let resizeObserver: ResizeObserver | null = null;
    if ("ResizeObserver" in window) {
      resizeObserver = new ResizeObserver(() => measureAura());
      resizeObserver.observe(shell);
      const activeHeroNode = shell.querySelector<HTMLElement>(`[data-unit-instance-id="${activeHeroInstanceId}"]`);
      if (activeHeroNode) {
        resizeObserver.observe(activeHeroNode);
      }
    }

    return () => {
      window.cancelAnimationFrame(frameId);
      window.removeEventListener("resize", handleResize);
      resizeObserver?.disconnect();
    };
  }, [enemyHeroInstanceId, match, playerHeroInstanceId, playerIndex, preload.visible]);

  function spawnFloatingNumbers(nextMatch: MaskedBattleMatchState) {
    const shell = shellRef.current;
    if (!shell || !nextMatch.events?.length) {
      return;
    }

    const shellRect = shell.getBoundingClientRect();
    const entries: FloatingNumber[] = [];

    nextMatch.events.forEach((event, eventIndex) => {
      const numberKind = resolveFloatingNumberKind(nextMatch, event);

      event.targets?.forEach((target, targetIndex) => {
        if (!target.instance_id || !target.amount || target.amount <= 0) {
          return;
        }
        const node = shell.querySelector<HTMLElement>(`[data-unit-instance-id="${target.instance_id}"]`);
        if (!node) {
          return;
        }

        const rect = node.getBoundingClientRect();
        entries.push({
          id: `${nextMatch.version}-${eventIndex}-${targetIndex}-${target.instance_id}`,
          left: rect.left - shellRect.left + rect.width / 2,
          top: rect.top - shellRect.top + rect.height * 0.18,
          amount: target.amount,
          kind: numberKind,
        });
      });
    });

    if (entries.length === 0) {
      return;
    }

    setFloatingNumbers((current) => [...current, ...entries]);
    window.setTimeout(() => {
      setFloatingNumbers((current) => current.filter((entry) => !entries.some((added) => added.id === entry.id)));
    }, 2000);
  }

  function resolveFloatingNumberKind(nextMatch: MaskedBattleMatchState, event: BattleEvent): FloatingNumber["kind"] {
    const eventType = event.type.toLowerCase();
    const effectKind = event.effect_kind?.toLowerCase() ?? "";

    if (eventType.includes("heal") || effectKind === "heal") {
      return "heal";
    }
    if (eventType === "buff" || effectKind === "buff") {
      return "heal";
    }
    if (eventType === "card_skill" || eventType === "hero_spell") {
      const firstTargetEffect = event.targets?.find((target) => target.instance_id)?.instance_id
        ? findEventTargetEffectKind(nextMatch, event)
        : "";
      if (firstTargetEffect === "buff" || firstTargetEffect === "heal") {
        return "heal";
      }
    }

    return "damage";
  }

  function findEventTargetEffectKind(nextMatch: MaskedBattleMatchState, event: BattleEvent) {
    const targetIds = new Set((event.targets ?? []).map((target) => target.instance_id).filter(Boolean));
    if (targetIds.size === 0) {
      return "";
    }

    for (const battlePlayer of nextMatch.players ?? []) {
      for (const unit of battlePlayer?.table ?? []) {
        if (!unit || !targetIds.has(unit.instance_id)) {
          continue;
        }
        const matchingEffect = [...(unit.effects ?? [])]
          .reverse()
          .find((effect) => !event.source_instance_id || effect.source_instance_id === event.source_instance_id);
        if (matchingEffect?.polarity === "buff") {
          return "buff";
        }
        if (matchingEffect?.effect_type === "heal_per_turn" || matchingEffect?.effect_type === "hp") {
          return "heal";
        }
        if (matchingEffect?.polarity === "debuff") {
          return "debuff";
        }
      }
    }

    return "";
  }

  function getEventSequenceId(event: BattleEvent, fallback: string) {
    return event.id ?? event.event_id ?? fallback;
  }

  function getViewerIndex(matchState: MaskedBattleMatchState | null) {
    if (!matchState) {
      return -1;
    }
    return matchState.players.findIndex((battlePlayer) => battlePlayer?.user_id === currentUserId);
  }

  function findUnitInMatch(matchState: MaskedBattleMatchState | null, instanceId: string) {
    if (!matchState || !instanceId) {
      return null;
    }

    for (const battlePlayer of matchState.players) {
      const unit = battlePlayer?.table.find((entry) => entry?.instance_id === instanceId);
      if (unit) {
        return unit;
      }
    }

    return null;
  }

  function makeHeroAnimationUnit(matchState: MaskedBattleMatchState, sourcePlayerIndex: number): BattleUnitState | null {
    const sourcePlayer = matchState.players[sourcePlayerIndex];
    if (!sourcePlayer) {
      return null;
    }

    const heroTemplate = heroes.find((hero) => hero.hero_code === sourcePlayer.hero_code) ?? null;
    return {
      instance_id: `hero:p${sourcePlayerIndex}`,
      template_id: sourcePlayer.hero_code,
      gamer_card_id: 0,
      card_level: sourcePlayer.hero_level,
      hp: sourcePlayer.hero_hp,
      max_hp: heroTemplate?.health_points ?? sourcePlayer.hero_hp,
      attack: sourcePlayer.hero_attack_power,
      splash_radius: sourcePlayer.hero_splash_radius,
      is_tank: false,
      card_type: "hero",
      base_cooldown: sourcePlayer.hero_attack_base_cooldown,
      cooldown: sourcePlayer.hero_attack_cooldown,
      attacks_this_turn: 0,
      summoned_in_turn: 0,
      image_key: "",
      asset_base_key: "",
      has_skill: false,
      skill_image_key: "",
      skill: null,
      effects: [],
      resurrected_used: false,
    };
  }

  function enqueueAttackAnimations(nextMatch: MaskedBattleMatchState, prevMatchOverride?: MaskedBattleMatchState | null) {
    void nextMatch;
    void prevMatchOverride;
  }

  const cardAttack = useCardAttack({
    player,
    enemy,
    isPlayerTurn,
    busy,
    finished: Boolean(match?.finished),
    onAttack: async (attacker, target, attackHero) => {
      if (!match) {
        return;
      }
      await applyAction({
        type: "card_attack",
        expected_version: match.version,
        card_instance_id: attacker.instance_id,
        target_instance_id: target?.instance_id,
        attack_hero: attackHero,
      });
    },
  });
  const cardSkill = useCardSkill({
    player,
    enemy,
    isPlayerTurn,
    busy,
    finished: Boolean(match?.finished),
    onInfo: showToast,
    onCast: async (caster, target, attackHero) => {
      if (!match) {
        return;
      }

      const canAttackAfterSkill = canUnitAttackNow(caster, player?.turns ?? 0);

      await applyAction({
        type: "card_skill",
        expected_version: match.version,
        card_instance_id: caster.instance_id,
        target_instance_id: target?.instance_id,
        attack_hero: attackHero,
      });

      if (canAttackAfterSkill) {
        cardAttack.selectAttacker(caster);
      }
    },
  });
  const heroAbility = useHeroAbility({
    player,
    enemy,
    playerHero,
    isPlayerTurn,
    busy,
    finished: Boolean(match?.finished),
    onInfo: showToast,
    onCast: async (target, attackHero) => {
      if (!match) {
        return;
      }

      await applyAction({
        type: "hero_spell",
        expected_version: match.version,
        target_instance_id: target?.instance_id,
        attack_hero: attackHero,
      });
    },
  });
  const onHitEffects = useOnHitEffects(match);
  const onPlayEffects = useOnPlayEffects({
    version: match?.version ?? 0,
    playerTable: player?.table ?? [],
    enemyTable: enemy?.table ?? [],
    shellWidth,
    shellHeight,
    getUnitRect,
  });
  const projectileRuntime = useProjectileRuntime({ match, playerIndex, getUnitRect: getCombatantRect });
  useBattleEventSfx(match);

  function showToast(message: string) {
    setToastMessage(message);
    setToastNonce((current) => current + 1);
  }

  function mapOriginRect(rect?: DOMRect | null): BattleCardViewerOrigin | null {
    const shellRect = shellRef.current?.getBoundingClientRect();
    if (!rect || !shellRect) {
      return null;
    }

    return {
      left: rect.left - shellRect.left,
      top: rect.top - shellRect.top,
      width: rect.width,
      height: rect.height,
    };
  }

  function enqueueDeathAnimations(nextMatch: MaskedBattleMatchState) {
    const entries = collectDeathAnimations({
      prevMatch: currentMatchRef.current,
      nextMatch,
      getUnitRect,
    });

    if (entries.length === 0) {
      return;
    }

    setDeathAnimations((current) => [...current, ...entries]);
  }

  useEffect(() => {
    currentMatchRef.current = match;
  }, [match]);

  useEffect(() => {
    readyRequestSentForMatchRef.current = null;
    setGraveyardOpen(false);
    setGraveyardCards([]);
    setGraveyardError("");
    setGraveyardLoading(false);
  }, [matchId]);

  useEffect(() => {
    if (!graveyardOpen) {
      return;
    }
    if (playerGraveyardCount <= 0) {
      setGraveyardCards([]);
      setGraveyardError("");
      setGraveyardLoading(false);
      return;
    }

    let cancelled = false;
    setGraveyardLoading(true);
    setGraveyardError("");

    request<GraveyardResponse>(`/matches/${matchId}/graveyard`)
      .then((response) => {
        if (cancelled) {
          return;
        }
        setGraveyardCards(response.cards ?? []);
      })
      .catch((err) => {
        if (cancelled) {
          return;
        }
        setGraveyardError(err instanceof Error ? err.message : "Failed to load graveyard");
      })
      .finally(() => {
        if (!cancelled) {
          setGraveyardLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [graveyardOpen, matchId, playerGraveyardCount]);

  function handlePreview(card: BattleCardInMatch | null, originRect?: DOMRect) {
    if (!card) {
      setPreviewClosing(true);
      cardAttack.clearSelection();
      cardSkill.clearSelection();
      heroAbility.clearSelection();
      setHeroAttackSelected(false);
      return;
    }

    cardAttack.clearSelection();
    cardSkill.clearSelection();
    heroAbility.clearSelection();
    setHeroAttackSelected(false);
    setPreviewOrigin(mapOriginRect(originRect));
    setPreviewCard(card);
    setPreviewClosing(false);
  }

  const heroAttackTargetIds = useMemo(() => {
    if (!heroAttackSelected || !enemy || !canHeroAttack) {
      return [];
    }
    if ((player?.hero_attack_cooldown ?? 0) > 0) {
      return [];
    }

    const tanks = enemy.table.filter((unit): unit is NonNullable<typeof unit> => Boolean(unit?.is_tank));
    return (tanks.length > 0 ? tanks : enemy.table.filter((unit): unit is NonNullable<typeof unit> => Boolean(unit))).map(
      (unit) => unit.instance_id,
    );
  }, [canHeroAttack, enemy, heroAttackSelected, player?.hero_attack_cooldown]);

  const canHeroAttackEnemyHero = useMemo(() => {
    if (!heroAttackSelected || !enemy || !canHeroAttack) {
      return false;
    }
    if ((player?.hero_attack_cooldown ?? 0) > 0) {
      return false;
    }

    return enemy.table.every((unit) => !unit?.is_tank);
  }, [canHeroAttack, enemy, heroAttackSelected, player?.hero_attack_cooldown]);

  const heroAttackHint = useMemo(() => {
    if (!heroAttackSelected) {
      return "";
    }
    if ((player?.hero_attack_cooldown ?? 0) > 0) {
      return `КД АТАКИ ГЕРОЯ - ${player?.hero_attack_cooldown ?? 0}`;
    }
    return "";
  }, [heroAttackSelected, player?.hero_attack_cooldown]);

  const readyUnitIds = useMemo(() => {
    if (!isPlayerTurn || !player || busy || match?.finished) {
      return [];
    }
    return player.table
      .filter((unit): unit is NonNullable<typeof unit> => Boolean(unit && canUnitAttackNow(unit, player.turns)))
      .map((unit) => unit.instance_id);
  }, [busy, isPlayerTurn, match?.finished, player]);

  useEffect(() => {
    if (!match?.finished || playerIndex < 0 || outcome) {
      return;
    }

    let nextOutcome: "victory" | "defeat" | "draw" | null = null;
    if (match.result === "DRAW") {
      nextOutcome = "draw";
    } else if (
      (match.result === "P1_WIN" && playerIndex === 0) ||
      (match.result === "P2_WIN" && playerIndex === 1)
    ) {
      nextOutcome = "victory";
    } else if (match.result === "P1_WIN" || match.result === "P2_WIN") {
      nextOutcome = "defeat";
    }

    if (!nextOutcome) {
      return;
    }

    setOutcome(nextOutcome);
  }, [match, outcome, onLeaveToMenu, playerIndex]);

  useEffect(() => {
    if (!outcome) {
      return;
    }

    const id = window.setTimeout(() => {
      setOutcome(null);
      onLeaveToMenu();
    }, 2000);

    return () => window.clearTimeout(id);
  }, [onLeaveToMenu, outcome]);

  useEffect(() => {
    if (!match || match.version <= lastProcessedVersionRef.current) {
      return;
    }

    lastProcessedVersionRef.current = match.version;
    spawnFloatingNumbers(match);
  }, [match]);

  useEffect(() => {
    if (!isPlayerTurn) {
      setHeroAttackSelected(false);
      cardAttack.clearSelection();
      cardSkill.clearSelection();
      heroAbility.clearSelection();
    }
  }, [cardAttack, cardSkill, heroAbility, isPlayerTurn]);

  useEffect(() => {
    if (!match || match.phase !== "START" || playerIndex < 0 || !preload.completed || playerReady) {
      return;
    }

    if (readyRequestSentForMatchRef.current === match.match_id) {
      return;
    }

    readyRequestSentForMatchRef.current = match.match_id;

    void request<MaskedBattleMatchState>(`/matches/${matchId}/ready`, {
      method: "POST",
    })
      .then((nextMatch) => {
        setMatch(nextMatch);
        setError("");
      })
      .catch((err) => {
        readyRequestSentForMatchRef.current = null;
        setError(err instanceof Error ? err.message : "Failed to confirm battle readiness");
      });
  }, [match, matchId, playerIndex, playerReady, preload.completed]);

  useEffect(() => {
    async function loadMatch() {
      setLoading(true);
      try {
        const nextMatch = await request<MaskedBattleMatchState>(`/matches/${matchId}`);
        setMatch(nextMatch);
        setError("");
      } catch (err) {
        setError(err instanceof Error ? err.message : "Не удалось загрузить матч");
      } finally {
        setLoading(false);
      }
    }

    void loadMatch();
  }, [matchId]);

  useEffect(() => {
    if (!match || match.finished) {
      return;
    }

    const id = window.setInterval(() => {
      const current = currentMatchRef.current;
      const now = Date.now();
      const streamHasState = streamHadStateRef.current;
      const streamIsFresh = streamConnectedRef.current && streamHasState;
      const startupStreamStalled =
        current?.phase === "START" && (!streamHasState || now - lastStreamStateAtRef.current > 5000);

      if (current?.finished || (streamIsFresh && !startupStreamStalled)) {
        return;
      }

      void request<MaskedBattleMatchState>(`/matches/${matchId}`)
        .then((nextMatch) => {
          const currentVersion = currentMatchRef.current?.version ?? 0;
          if (nextMatch.version <= currentVersion) {
            return;
          }
          if (nextMatch.version > currentVersion) {
            enqueueDeathAnimations(nextMatch);
          }
          setMatch(nextMatch);
          setError("");
        })
        .catch(() => undefined);
    }, 5000);

    return () => window.clearInterval(id);
  }, [match?.finished, matchId]);

  useEffect(() => {
    streamConnectedRef.current = false;
    streamHadStateRef.current = false;
    lastStreamStateAtRef.current = 0;

    const stream = new EventSource(withDevSessionToken(`/matches/${matchId}/stream`), { withCredentials: true });

    stream.onopen = () => {
      streamConnectedRef.current = true;
    };

    stream.addEventListener("state", (event) => {
      const payload = JSON.parse((event as MessageEvent<string>).data) as MaskedBattleMatchState;
      const currentVersion = currentMatchRef.current?.version ?? 0;
      streamConnectedRef.current = true;
      streamHadStateRef.current = true;
      lastStreamStateAtRef.current = Date.now();
      if (payload.version <= currentVersion) {
        return;
      }
      enqueueAttackAnimations(payload);
      enqueueDeathAnimations(payload);
      setMatch(payload);
      setError("");
    });

    stream.onerror = () => {
      streamConnectedRef.current = false;
    };

    return () => stream.close();
  }, [matchId, playerIndex]);

  async function applyAction(payload: ApplyBattleActionRequest, leaveAfter = false) {
    if (!match) {
      return;
    }

    const prevMatch = match;
    setBusy(true);
    try {
      const nextMatch = await request<MaskedBattleMatchState>(`/matches/${matchId}/actions`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      enqueueAttackAnimations(nextMatch, prevMatch);
      enqueueDeathAnimations(nextMatch);
      setMatch(nextMatch);
      setError("");
      if (leaveAfter && !nextMatch.finished) {
        onLeaveToMenu();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Действие не выполнено");
    } finally {
      setBusy(false);
    }
  }

  if (loading) {
    return <section className="battle-screen battle-screen--state">ЗАГРУЖАЕМ МАТЧ...</section>;
  }

  if (!match || !player || !enemy) {
    return <section className="battle-screen battle-screen--state">{error || "МАТЧ НЕДОСТУПЕН"}</section>;
  }

  if (preload.visible || match.phase === "START") {
    const loadingLabel = preload.completed
      ? playerReady
        ? "Ready. Waiting for opponent"
        : "Confirming readiness"
      : preload.label;

    return (
      <BattleLoadingScreen
        progress={preload.completed ? 1 : preload.progress}
        label={loadingLabel}
        playerReady={playerReady}
        enemyReady={enemyReady}
      />
    );
  }

  return (
    <section className="battle-screen">
      <div ref={shellRef} className="battle-shell surface">
        {turnAuras.length > 0 ? (
          <div className="battle-turn-aura-layer" aria-hidden="true">
            {turnAuras.map((turnAura) => (
              <div
                key={turnAura.id}
                className={`battle-turn-aura battle-turn-aura--${turnAura.phase}`}
                style={{
                  left: `${turnAura.centerX}px`,
                  top: `${turnAura.centerY}px`,
                  width: `${turnAura.width}px`,
                  height: `${turnAura.height}px`,
                }}
              />
            ))}
          </div>
        ) : null}
        <div className="battle-top">
          <div className="battle-top__header">
            <LeaveMatchButton
              disabled={busy}
              onLeave={() =>
                void applyAction(
                  {
                    type: "leave_match",
                    expected_version: match.version,
                  },
                  true,
                )
              }
            />
            <div className="battle-turn-status" aria-live="polite">
              <span className={`battle-turn-status__state ${isPlayerTurn ? "battle-turn-status__state--active" : ""}`}>
                {activeTurnLabel}
              </span>
              <span className="battle-turn-status__turn">ХОД {turnNumber}</span>
            </div>
          </div>
          <BattleEventFeed match={match} heroes={heroes} />
          <EnemyCharacter
            player={enemy as MaskedBattlePlayerState}
            maxHp={enemyHero?.health_points ?? enemy.hero_hp}
            isActive={match.active_player !== playerIndex}
            attackTarget={cardAttack.canAttackHero || canHeroAttackEnemyHero || cardSkill.canTargetHero || (heroAbility.selected && heroAbility.canTargetHero)}
            heroInstanceId={enemyHeroInstanceId}
            hitToken={onHitEffects.hitTokens[enemyHeroInstanceId] ?? 0}
            attackAnimation={enemyHeroAttackAnimation}
            onClick={
              cardAttack.canAttackHero || canHeroAttackEnemyHero || cardSkill.canTargetHero || (heroAbility.selected && heroAbility.canTargetHero)
                ? () => {
                    setPreviewClosing(true);
                    if (cardSkill.canTargetHero) {
                      void cardSkill.tryCastOnHero();
                      return;
                    }
                    if (heroAbility.selected && heroAbility.canTargetHero) {
                      void heroAbility.tryCastOnHero();
                      return;
                    }
                    if (heroAttackSelected) {
                      void (async () => {
                        await applyAction({
                          type: "hero_attack",
                          expected_version: match.version,
                          attack_hero: true,
                        });
                        setHeroAttackSelected(false);
                      })();
                      return;
                    }
                    void cardAttack.tryAttackHero();
                  }
                : undefined
            }
          />
        </div>

          <BattleField
            match={match}
            enemy={enemy as MaskedBattlePlayerState}
            player={player as MaskedBattlePlayerState}
            cardNameByTemplateId={battleCardNameByTemplateId}
             canEndTurn={canEndTurn}
             canPlaySelectedBattleCard={Boolean(canPlaySelectedBattleCard)}
             selectedAttackerId={cardAttack.selectedAttackerId}
             selectedSkillCasterId={cardSkill.selectedCasterId}
             attackTargetIds={heroAttackSelected ? heroAttackTargetIds : cardAttack.attackTargets}
             readyUnitIds={readyUnitIds}
             skillTargetIds={
               heroAttackSelected
                 ? []
                 : cardSkill.selectedCasterId
                   ? cardSkill.skillTargetIds
                   : heroAbility.selected
                     ? heroAbility.targetIds
                     : []
             }
             skillTargetTone={
               heroAttackSelected
                 ? null
                 : cardSkill.selectedCasterId
                   ? cardSkill.targetTone
                   : heroAbility.selected
                     ? heroAbility.targetTone
                     : null
             }
             attackHint={
               heroAttackSelected
                 ? heroAttackHint
                 : cardSkill.selectedCasterId
                   ? ""
                   : heroAbility.selected
                     ? heroAbility.infoMessage
                     : cardAttack.infoMessage
             }
             animatingUnitId={attackAnimation?.attacker.instance_id ?? ""}
             hitTokens={onHitEffects.hitTokens}
             disabledPlayerUnitIds={
               cardSkill.selectedCasterId || heroAbility.selected
                 ? []
                 : player.table
                     .filter((unit): unit is NonNullable<typeof unit> => Boolean(unit && getStunTurns(unit) > 0))
                     .map((unit) => unit.instance_id)
             }
             skillDisabledPlayerUnitIds={player.table
               .filter((unit): unit is NonNullable<typeof unit> => Boolean(unit && getStunTurns(unit) > 0))
               .map((unit) => unit.instance_id)}
             onEndTurn={() =>
              void applyAction({
                type: "end_turn",
                expected_version: match.version,
              })
            }
            onPlayerUnitSelect={(unit) => {
              setPreviewClosing(true);
              if (cardSkill.selectedCasterId) {
                if (cardSkill.skillTargetIds.includes(unit.instance_id)) {
                  void cardSkill.tryCastOnUnit(unit);
                  return;
                }
                cardSkill.clearSelection();
                return;
              }
              if (heroAbility.selected) {
                if (heroAbility.targetIds.includes(unit.instance_id)) {
                  void heroAbility.tryCastOnUnit(unit);
                  return;
                }
                heroAbility.clearSelection();
                return;
              }
              if (getStunTurns(unit) > 0) {
                return;
              }
              setHeroAttackSelected(false);
              cardAttack.selectAttacker(unit);
            }}
            onEnemyUnitSelect={(unit) => {
              setPreviewClosing(true);
              if (cardSkill.selectedCasterId) {
                if (cardSkill.skillTargetIds.includes(unit.instance_id)) {
                  void cardSkill.tryCastOnUnit(unit);
                  return;
                }
                cardSkill.clearSelection();
                return;
              }
              if (heroAbility.selected) {
                if (heroAbility.targetIds.includes(unit.instance_id)) {
                  void heroAbility.tryCastOnUnit(unit);
                  return;
                }
                heroAbility.clearSelection();
                return;
              }
              if (heroAttackSelected) {
                void (async () => {
                  await applyAction({
                    type: "hero_attack",
                    expected_version: match.version,
                    target_instance_id: unit.instance_id,
                  });
                  setHeroAttackSelected(false);
                })();
                return;
              }
              void cardAttack.tryAttack(unit);
            }}
            onBoardClearSelection={() => {
              setPreviewClosing(true);
              cardSkill.clearSelection();
              heroAbility.clearSelection();
              setHeroAttackSelected(false);
              cardAttack.clearSelection();
            }}
            onPlayBattleCard={(slotIndex) => {
              cardSkill.clearSelection();
              heroAbility.clearSelection();
              if (!previewCard || previewCard.kind !== "battle") {
                cardAttack.clearSelection();
                return;
              }

              void applyAction({
                type: "play_battle_card",
                expected_version: match.version,
                card_instance_id: previewCard.instance_id,
                target_slot: slotIndex,
              }).then(() => {
                cardAttack.clearSelection();
                setPreviewClosing(false);
                setPreviewCard(null);
                setPreviewOrigin(null);
              });
            }}
            onPlayerUnitSkill={(unit) => {
              setPreviewClosing(true);
              setPreviewCard(null);
              setPreviewOrigin(null);
              cardAttack.clearSelection();
              heroAbility.clearSelection();
              setHeroAttackSelected(false);
              void cardSkill.selectCaster(unit);
            }}
          />

        <div className="battle-bottom">
          <div className="battle-bottom__hero-row">
            <div className="battle-bottom__cluster battle-bottom__cluster--left">
              <div className="battle-bottom__mini">
                <GraveyardBlock
                  count={playerGraveyardCount}
                  onOpen={() => {
                    setPreviewClosing(true);
                    setPreviewCard(null);
                    setPreviewOrigin(null);
                    setGraveyardOpen(true);
                  }}
                />
              </div>
              <div className="battle-bottom__main-block">
                <AttackBlock
                  attack={player.hero_attack_power}
                  cooldown={player.hero_attack_cooldown}
                  selected={heroAttackSelected}
                  disabled={!canHeroAttack}
                  onClick={() => {
                    if (!canHeroAttack) {
                      return;
                    }
                    setPreviewClosing(true);
                    setPreviewCard(null);
                    setPreviewOrigin(null);
                    cardAttack.clearSelection();
                    cardSkill.clearSelection();
                    heroAbility.clearSelection();
                    setHeroAttackSelected((current) => !current);
                  }}
                />
              </div>
            </div>

            <div className="battle-bottom__hero">
              <GamerCharacter
                player={player as MaskedBattlePlayerState}
                maxHp={playerHero?.health_points ?? player.hero_hp}
                isActive={match.active_player === playerIndex}
                heroInstanceId={playerHeroInstanceId}
                hitToken={onHitEffects.hitTokens[playerHeroInstanceId] ?? 0}
                attackAnimation={heroAttackAnimation}
              />
            </div>

            <div className="battle-bottom__cluster battle-bottom__cluster--right">
              <div className="battle-bottom__main-block">
                <AbilityBlock
                  cooldown={player.hero_ability_cooldown}
                  manaCost={player.hero_ability_mana_cost ?? 0}
                  selected={heroAbility.selected}
                  disabled={!canHeroAttack}
                  onClick={() => {
                    if (!canHeroAttack) {
                      return;
                    }
                    setPreviewClosing(true);
                    setPreviewCard(null);
                    setPreviewOrigin(null);
                    cardAttack.clearSelection();
                    cardSkill.clearSelection();
                    setHeroAttackSelected(false);
                    void heroAbility.toggleSelection();
                  }}
                />
              </div>
              <div className="battle-bottom__mini">
                <DeckCounter count={player.deck_count ?? player.deck?.length ?? 0} />
              </div>
            </div>
          </div>

          <div className="battle-bottom__hand">
            <HandPanel hand={player.hand ?? []} selectedCardId={previewCard?.instance_id ?? ""} onPreview={handlePreview} />
          </div>
        </div>

        {previewCard ? (
          <BattleCardViewer
            card={previewCard}
            origin={previewOrigin}
            shellHeight={shellHeight}
            closing={previewClosing}
            onClose={() => setPreviewClosing(true)}
            onExited={() => {
              setPreviewClosing(false);
              setPreviewCard(null);
              setPreviewOrigin(null);
            }}
          />
        ) : null}
        {graveyardOpen ? (
          <BattleGraveyardModal
            cards={graveyardCards}
            loading={graveyardLoading}
            error={graveyardError}
            onClose={() => setGraveyardOpen(false)}
            onOpenCard={(card, originRect) => {
              setGraveyardOpen(false);
              handlePreview(card, originRect);
            }}
          />
        ) : null}
        <BattlePlayAnimations animations={onPlayEffects.animations} onDone={onPlayEffects.removeAnimation} />
        {attackAnimation ? (
          <CardAttackAnimation state={attackAnimation} onDone={() => setAttackAnimation(null)} />
        ) : null}
      <ProjectileLayer
        projectiles={projectileRuntime.projectiles}
        impacts={projectileRuntime.impacts}
        spreads={projectileRuntime.spreads}
      />
        <BattleDeathAnimations
          animations={deathAnimations}
          onDone={(id) => {
            setDeathAnimations((current) => current.filter((entry) => entry.id !== id));
          }}
        />
        <BattleFloatingNumbers numbers={floatingNumbers} />
        {error || toastMessage ? <BattleInfoToast message={error || toastMessage} nonce={toastNonce} /> : null}
      </div>
      {outcome === "victory" ? <Victory /> : null}
      {outcome === "defeat" ? <Defeat /> : null}
      {outcome === "draw" ? <Draw /> : null}
    </section>
  );
}

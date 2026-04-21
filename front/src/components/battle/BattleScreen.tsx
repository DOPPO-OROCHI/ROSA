import { useEffect, useMemo, useRef, useState } from "react";
import { request } from "../../lib/api";
import { AbilityBlock } from "./AbilityBlock";
import { AttackBlock } from "./AttackBlock";
import { getStunTurns, useCardSkill } from "./CARD_SKILLS";
import { CardAttackAnimation, type CardAttackAnimationState } from "./CardAttackAnimation";
import { BattleCardViewer, type BattleCardViewerOrigin } from "./BattleCardViewer";
import { BattleField } from "./BattleField";
import { BattleFloatingNumbers, type FloatingNumber } from "./BattleFloatingNumbers";
import { BattleInfoToast } from "./BattleInfoToast";
import { BattleLoadingScreen } from "./BattleLoadingScreen";
import { DeckCounter } from "./DeckCounter";
import { Defeat } from "./Defeat";
import { Draw } from "./Draw";
import { EnemyCharacter } from "./EnemyCharacter";
import { GamerCharacter } from "./GamerCharacter";
import { GraveyardBlock } from "./GraveyardBlock";
import { HandPanel } from "./HandPanel";
import { LeaveMatchButton } from "./LeaveMatchButton";
import { useBattlePreload } from "./battle_preload";
import { useCardAttack } from "./card_attack";
import { BattleDeathAnimations, collectDeathAnimations, type DeathAnimationState } from "./on_death_effects";
import { useOnHitEffects } from "./on_hit_effects";
import { BattlePlayAnimations, useOnPlayEffects } from "./on_play_effects";
import { ProjectileLayer, useProjectileRuntime } from "./projectiles";
import { useBattleEventSfx } from "./sfx";
import { Victory } from "./Victory";
import type { ApplyBattleActionRequest, BattleCardInMatch, MaskedBattleMatchState, MaskedBattlePlayerState } from "./types";
import type { Hero } from "../../types";
import "./battle.css";

type Props = {
  currentUserId: number;
  matchId: number;
  heroes: Hero[];
  onLeaveToMenu: () => void;
};

export function BattleScreen({ currentUserId, matchId, heroes, onLeaveToMenu }: Props) {
  const [match, setMatch] = useState<MaskedBattleMatchState | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [outcome, setOutcome] = useState<"victory" | "defeat" | "draw" | null>(null);
  const [previewCard, setPreviewCard] = useState<BattleCardInMatch | null>(null);
  const [previewOrigin, setPreviewOrigin] = useState<BattleCardViewerOrigin | null>(null);
  const [previewClosing, setPreviewClosing] = useState(false);
  const [attackAnimation, setAttackAnimation] = useState<CardAttackAnimationState | null>(null);
  const [deathAnimations, setDeathAnimations] = useState<DeathAnimationState[]>([]);
  const [floatingNumbers, setFloatingNumbers] = useState<FloatingNumber[]>([]);
  const [heroAttackSelected, setHeroAttackSelected] = useState(false);
  const [toastMessage, setToastMessage] = useState("");
  const [toastNonce, setToastNonce] = useState(0);
  const [turnAura, setTurnAura] = useState<{
    centerX: number;
    centerY: number;
    width: number;
    height: number;
  } | null>(null);
  const shellRef = useRef<HTMLDivElement | null>(null);
  const lastProcessedVersionRef = useRef(0);
  const currentMatchRef = useRef<MaskedBattleMatchState | null>(null);

  const playerIndex = useMemo(() => {
    if (!match) {
      return -1;
    }
    return match.players.findIndex((player) => player?.user_id === currentUserId);
  }, [currentUserId, match]);

  const player = playerIndex >= 0 ? match?.players[playerIndex] ?? null : null;
  const enemy = playerIndex >= 0 ? match?.players[playerIndex === 0 ? 1 : 0] ?? null : null;
  const preload = useBattlePreload(match);
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

  useEffect(() => {
    if (!match || playerIndex < 0 || preload.visible) {
      setTurnAura(null);
      return;
    }

    const shell = shellRef.current;
    if (!shell) {
      setTurnAura(null);
      return;
    }

    const activeHeroInstanceId = match.active_player === playerIndex ? playerHeroInstanceId : enemyHeroInstanceId;

    function measureAura() {
      const rect = getUnitRect(activeHeroInstanceId);
      if (!rect) {
        setTurnAura(null);
        return;
      }

      setTurnAura({
        centerX: rect.centerX,
        centerY: rect.centerY,
        width: rect.width * 3.8,
        height: rect.height * 1.18,
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
      const isHealEvent = event.type.toLowerCase().includes("heal");

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
          kind: isHealEvent ? "heal" : "damage",
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

  async function playCardAttackAnimation(attacker: MaskedBattlePlayerState["table"][number], target: MaskedBattlePlayerState["table"][number]) {
    if (!attacker || !target) {
      return;
    }

    const from = getUnitRect(attacker.instance_id);
    const to = getUnitRect(target.instance_id);
    if (!from || !to) {
      return;
    }

    await new Promise<void>((resolve) => {
      setAttackAnimation({
        attacker,
        from: {
          left: from.left,
          top: from.top,
          width: from.width,
          height: from.height,
        },
        dx: to.centerX - from.centerX,
        dy: to.centerY - from.centerY,
      });

      window.setTimeout(resolve, 460);
    });
  }

  async function playCardAttackHeroAnimation(attacker: MaskedBattlePlayerState["table"][number]) {
    if (!attacker) {
      return;
    }

    const from = getUnitRect(attacker.instance_id);
    const shell = shellRef.current;
    const heroNode = shell?.querySelector<HTMLElement>('[data-hero-side="enemy"]');
    if (!from || !heroNode || !shell) {
      return;
    }

    const shellRect = shell.getBoundingClientRect();
    const rect = heroNode.getBoundingClientRect();
    const to = {
      centerX: rect.left - shellRect.left + rect.width / 2,
      centerY: rect.top - shellRect.top + rect.height / 2,
    };

    await new Promise<void>((resolve) => {
      setAttackAnimation({
        attacker,
        from: {
          left: from.left,
          top: from.top,
          width: from.width,
          height: from.height,
        },
        dx: to.centerX - from.centerX,
        dy: to.centerY - from.centerY,
      });

      window.setTimeout(resolve, 460);
    });
  }

  async function playHeroAttackAnimation(target: MaskedBattlePlayerState["table"][number] | null, attackHero: boolean) {
    const shell = shellRef.current;
    if (!shell || !player) {
      return;
    }

    const heroNode = shell.querySelector<HTMLElement>('[data-hero-side="player"]');
    if (!heroNode) {
      return;
    }

    const shellRect = shell.getBoundingClientRect();
    const heroRect = heroNode.getBoundingClientRect();
    const from = {
      left: heroRect.left - shellRect.left,
      top: heroRect.top - shellRect.top,
      width: heroRect.width,
      height: heroRect.height,
      centerX: heroRect.left - shellRect.left + heroRect.width / 2,
      centerY: heroRect.top - shellRect.top + heroRect.height / 2,
    };

    let targetCenter: { centerX: number; centerY: number } | null = null;
    if (attackHero) {
      const enemyHeroNode = shell.querySelector<HTMLElement>('[data-hero-side="enemy"]');
      if (enemyHeroNode) {
        const rect = enemyHeroNode.getBoundingClientRect();
        targetCenter = {
          centerX: rect.left - shellRect.left + rect.width / 2,
          centerY: rect.top - shellRect.top + rect.height / 2,
        };
      }
    } else if (target) {
      const rect = getUnitRect(target.instance_id);
      if (rect) {
        targetCenter = {
          centerX: rect.centerX,
          centerY: rect.centerY,
        };
      }
    }

    if (!targetCenter) {
      return;
    }

    await new Promise<void>((resolve) => {
      setAttackAnimation({
        attacker: {
          instance_id: playerHeroInstanceId,
          template_id: player.hero_code,
          gamer_card_id: 0,
          card_level: player.hero_level,
          hp: player.hero_hp,
          max_hp: playerHero?.health_points ?? player.hero_hp,
          attack: player.hero_attack_power,
          splash_radius: player.hero_splash_radius,
          is_tank: false,
          card_type: "hero",
          base_cooldown: player.hero_attack_base_cooldown,
          cooldown: player.hero_attack_cooldown,
          summoned_in_turn: 0,
          image_key: "",
          asset_base_key: "",
          has_skill: false,
          skill_image_key: "",
          skill: null,
          effects: [],
          resurrected_used: false,
        },
        from: {
          left: from.left,
          top: from.top,
          width: from.width,
          height: from.height,
        },
        dx: targetCenter.centerX - from.centerX,
        dy: targetCenter.centerY - from.centerY,
      });

      window.setTimeout(resolve, 460);
    });
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
      await Promise.all([
        attackHero ? playCardAttackHeroAnimation(attacker) : playCardAttackAnimation(attacker, target),
        (async () => {
          await new Promise((resolve) => window.setTimeout(resolve, 190));
          await applyAction({
            type: "card_attack",
            expected_version: match.version,
            card_instance_id: attacker.instance_id,
            target_instance_id: target?.instance_id,
            attack_hero: attackHero,
          });
        })(),
      ]);
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

      await applyAction({
        type: "card_skill",
        expected_version: match.version,
        card_instance_id: caster.instance_id,
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
  const projectileRuntime = useProjectileRuntime({ match, playerIndex, getUnitRect });
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

  function handlePreview(card: BattleCardInMatch | null, originRect?: DOMRect) {
    if (!card) {
      setPreviewClosing(true);
      cardAttack.clearSelection();
      cardSkill.clearSelection();
      setHeroAttackSelected(false);
      return;
    }

    cardAttack.clearSelection();
    cardSkill.clearSelection();
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
    }
  }, [cardAttack, cardSkill, isPlayerTurn]);

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
    const stream = new EventSource(`/matches/${matchId}/stream`, { withCredentials: true });

    stream.addEventListener("state", (event) => {
      const payload = JSON.parse((event as MessageEvent<string>).data) as MaskedBattleMatchState;
      enqueueDeathAnimations(payload);
      setMatch(payload);
      setError("");
    });

    stream.onerror = () => {
      stream.close();
    };

    return () => stream.close();
  }, [matchId]);

  async function applyAction(payload: ApplyBattleActionRequest, leaveAfter = false) {
    if (!match) {
      return;
    }

    setBusy(true);
    try {
      const nextMatch = await request<MaskedBattleMatchState>(`/matches/${matchId}/actions`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      enqueueDeathAnimations(nextMatch);
      setMatch(nextMatch);
      setError("");
      if (leaveAfter) {
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

  if (preload.visible) {
    return <BattleLoadingScreen progress={preload.progress} label={preload.label} />;
  }

  return (
    <section className="battle-screen">
      <div ref={shellRef} className="battle-shell surface">
        {turnAura ? (
          <div className="battle-turn-aura-layer" aria-hidden="true">
            <div
              className="battle-turn-aura"
              style={{
                left: `${turnAura.centerX}px`,
                top: `${turnAura.centerY}px`,
                width: `${turnAura.width}px`,
                height: `${turnAura.height}px`,
              }}
            />
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
          <EnemyCharacter
            player={enemy as MaskedBattlePlayerState}
            maxHp={enemyHero?.health_points ?? enemy.hero_hp}
            isActive={match.active_player !== playerIndex}
            attackTarget={cardAttack.canAttackHero || canHeroAttackEnemyHero || cardSkill.canTargetHero}
            heroInstanceId={enemyHeroInstanceId}
            hitToken={onHitEffects.hitTokens[enemyHeroInstanceId] ?? 0}
            onClick={
              cardAttack.canAttackHero || canHeroAttackEnemyHero || cardSkill.canTargetHero
                ? () => {
                    setPreviewClosing(true);
                    if (cardSkill.canTargetHero) {
                      void cardSkill.tryCastOnHero();
                      return;
                    }
                    if (heroAttackSelected) {
                      void (async () => {
                        await Promise.all([
                          playHeroAttackAnimation(null, true),
                          (async () => {
                            await new Promise((resolve) => window.setTimeout(resolve, 190));
                            await applyAction({
                              type: "hero_attack",
                              expected_version: match.version,
                              attack_hero: true,
                            });
                          })(),
                        ]);
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
             canEndTurn={canEndTurn}
             canPlaySelectedBattleCard={Boolean(canPlaySelectedBattleCard)}
             selectedAttackerId={cardAttack.selectedAttackerId}
             selectedSkillCasterId={cardSkill.selectedCasterId}
             attackTargetIds={heroAttackSelected ? heroAttackTargetIds : cardAttack.attackTargets}
             skillTargetIds={heroAttackSelected ? [] : cardSkill.skillTargetIds}
             skillTargetTone={heroAttackSelected ? null : cardSkill.targetTone}
             attackHint={heroAttackSelected ? heroAttackHint : cardAttack.infoMessage}
             animatingUnitId={attackAnimation?.attacker.instance_id ?? ""}
             hitTokens={onHitEffects.hitTokens}
             disabledPlayerUnitIds={
               cardSkill.selectedCasterId
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
              }
              if (heroAttackSelected) {
                void (async () => {
                  await Promise.all([
                    playHeroAttackAnimation(unit, false),
                    (async () => {
                      await new Promise((resolve) => window.setTimeout(resolve, 190));
                      await applyAction({
                        type: "hero_attack",
                        expected_version: match.version,
                        target_instance_id: unit.instance_id,
                      });
                    })(),
                  ]);
                  setHeroAttackSelected(false);
                })();
                return;
              }
              void cardAttack.tryAttack(unit);
            }}
            onBoardClearSelection={() => {
              setPreviewClosing(true);
              cardSkill.clearSelection();
              setHeroAttackSelected(false);
              cardAttack.clearSelection();
            }}
            onPlayBattleCard={(slotIndex) => {
              cardSkill.clearSelection();
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
              setHeroAttackSelected(false);
              void cardSkill.selectCaster(unit);
            }}
          />

        <div className="battle-bottom">
          <div className="battle-bottom__hero-row">
            <div className="battle-bottom__cluster battle-bottom__cluster--left">
              <div className="battle-bottom__mini">
                <GraveyardBlock count={player.discard_count ?? player.discard?.length ?? 0} />
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
                <AbilityBlock cooldown={player.hero_ability_cooldown} manaCost={player.hero_ability_mana_cost ?? 0} />
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

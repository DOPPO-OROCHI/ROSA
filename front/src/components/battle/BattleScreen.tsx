import { useEffect, useMemo, useRef, useState } from "react";
import { request } from "../../lib/api";
import { AbilityBlock } from "./AbilityBlock";
import { AttackBlock } from "./AttackBlock";
import { CardAttackAnimation, type CardAttackAnimationState } from "./CardAttackAnimation";
import { BattleCardViewer, type BattleCardViewerOrigin } from "./BattleCardViewer";
import { BattleField } from "./BattleField";
import { BattleFloatingNumbers, type FloatingNumber } from "./BattleFloatingNumbers";
import { BattleInfoToast } from "./BattleInfoToast";
import { DeckCounter } from "./DeckCounter";
import { Defeat } from "./Defeat";
import { Draw } from "./Draw";
import { EnemyCharacter } from "./EnemyCharacter";
import { GamerCharacter } from "./GamerCharacter";
import { GraveyardBlock } from "./GraveyardBlock";
import { HandPanel } from "./HandPanel";
import { LeaveMatchButton } from "./LeaveMatchButton";
import { useCardAttack } from "./card_attack";
import { useOnHitEffects } from "./on_hit_effects";
import { BattlePlayAnimations, useOnPlayEffects } from "./on_play_effects";
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
  const [floatingNumbers, setFloatingNumbers] = useState<FloatingNumber[]>([]);
  const shellRef = useRef<HTMLDivElement | null>(null);
  const lastProcessedVersionRef = useRef(0);

  const playerIndex = useMemo(() => {
    if (!match) {
      return -1;
    }
    return match.players.findIndex((player) => player?.user_id === currentUserId);
  }, [currentUserId, match]);

  const player = playerIndex >= 0 ? match?.players[playerIndex] ?? null : null;
  const enemy = playerIndex >= 0 ? match?.players[playerIndex === 0 ? 1 : 0] ?? null : null;
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
  const shellWidth = shellRef.current?.clientWidth ?? 0;
  const shellHeight = shellRef.current?.clientHeight ?? 0;

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
        if (target.instance_id.startsWith("hero:")) {
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

  const cardAttack = useCardAttack({
    player,
    enemy,
    isPlayerTurn,
    busy,
    finished: Boolean(match?.finished),
    onAttack: async (attacker, target) => {
      if (!match) {
        return;
      }
      await Promise.all([
        playCardAttackAnimation(attacker, target),
        (async () => {
          await new Promise((resolve) => window.setTimeout(resolve, 190));
          await applyAction({
            type: "card_attack",
            expected_version: match.version,
            card_instance_id: attacker.instance_id,
            target_instance_id: target.instance_id,
          });
        })(),
      ]);
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

  function handlePreview(card: BattleCardInMatch | null, originRect?: DOMRect) {
    if (!card) {
      setPreviewClosing(true);
      cardAttack.clearSelection();
      return;
    }

    cardAttack.clearSelection();
    setPreviewOrigin(mapOriginRect(originRect));
    setPreviewCard(card);
    setPreviewClosing(false);
  }

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

  return (
    <section className="battle-screen">
      <div ref={shellRef} className="battle-shell surface">
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
          />
        </div>

          <BattleField
            match={match}
            enemy={enemy as MaskedBattlePlayerState}
            player={player as MaskedBattlePlayerState}
            canEndTurn={canEndTurn}
            canPlaySelectedBattleCard={Boolean(canPlaySelectedBattleCard)}
            selectedAttackerId={cardAttack.selectedAttackerId}
            attackTargetIds={cardAttack.attackTargets}
            attackHint={cardAttack.infoMessage}
            animatingUnitId={attackAnimation?.attacker.instance_id ?? ""}
            hitUnitIds={onHitEffects.hitUnitIds}
            onEndTurn={() =>
              void applyAction({
                type: "end_turn",
                expected_version: match.version,
              })
            }
            onPlayerUnitSelect={(unit) => {
              setPreviewClosing(true);
              cardAttack.selectAttacker(unit);
            }}
            onEnemyUnitSelect={(unit) => {
              setPreviewClosing(true);
              void cardAttack.tryAttack(unit);
            }}
            onBoardClearSelection={() => {
              setPreviewClosing(true);
              cardAttack.clearSelection();
            }}
            onPlayBattleCard={(slotIndex) => {
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
          />

        <div className="battle-bottom">
          <div className="battle-bottom__hero-row">
            <div className="battle-bottom__cluster battle-bottom__cluster--left">
              <div className="battle-bottom__mini">
                <GraveyardBlock count={player.discard_count ?? player.discard?.length ?? 0} />
              </div>
              <div className="battle-bottom__main-block">
                <AttackBlock attack={player.hero_attack_power} cooldown={player.hero_attack_cooldown} />
              </div>
            </div>

            <div className="battle-bottom__hero">
              <GamerCharacter
                player={player as MaskedBattlePlayerState}
                maxHp={playerHero?.health_points ?? player.hero_hp}
                isActive={match.active_player === playerIndex}
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
        {attackAnimation ? <CardAttackAnimation state={attackAnimation} onDone={() => setAttackAnimation(null)} /> : null}
        <BattleFloatingNumbers numbers={floatingNumbers} />
        {error ? <BattleInfoToast message={error} /> : null}
      </div>
      {outcome === "victory" ? <Victory /> : null}
      {outcome === "defeat" ? <Defeat /> : null}
      {outcome === "draw" ? <Draw /> : null}
    </section>
  );
}

import { useEffect, useMemo, useState } from "react";
import { request } from "../../lib/api";
import { AbilityBlock } from "./AbilityBlock";
import { AttackBlock } from "./AttackBlock";
import { BattleField } from "./BattleField";
import { DeckCounter } from "./DeckCounter";
import { Defeat } from "./Defeat";
import { Draw } from "./Draw";
import { EnemyCharacter } from "./EnemyCharacter";
import { GamerCharacter } from "./GamerCharacter";
import { GraveyardBlock } from "./GraveyardBlock";
import { HandPanel } from "./HandPanel";
import { LeaveMatchButton } from "./LeaveMatchButton";
import { Victory } from "./Victory";
import type { ApplyBattleActionRequest, MaskedBattleMatchState, MaskedBattlePlayerState } from "./types";
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
      <div className="battle-shell surface">
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
          onEndTurn={() =>
            void applyAction({
              type: "end_turn",
              expected_version: match.version,
            })
          }
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
            <HandPanel hand={player.hand ?? []} />
          </div>
        </div>

        {error ? <p className="battle-error">{error}</p> : null}
      </div>
      {outcome === "victory" ? <Victory /> : null}
      {outcome === "defeat" ? <Defeat /> : null}
      {outcome === "draw" ? <Draw /> : null}
    </section>
  );
}

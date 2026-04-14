import { useEffect, useMemo, useState } from "react";
import { request } from "../../lib/api";
import { AbilityBlock } from "./AbilityBlock";
import { AttackBlock } from "./AttackBlock";
import { BattleField } from "./BattleField";
import { DeckCounter } from "./DeckCounter";
import { EnemyCharacter } from "./EnemyCharacter";
import { GamerCharacter } from "./GamerCharacter";
import { GraveyardBlock } from "./GraveyardBlock";
import { HandPanel } from "./HandPanel";
import { LeaveMatchButton } from "./LeaveMatchButton";
import type { ApplyBattleActionRequest, MaskedBattleMatchState, MaskedBattlePlayerState } from "./types";
import "./battle.css";

type Props = {
  currentUserId: number;
  matchId: number;
  onLeaveToMenu: () => void;
};

export function BattleScreen({ currentUserId, matchId, onLeaveToMenu }: Props) {
  const [match, setMatch] = useState<MaskedBattleMatchState | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const playerIndex = useMemo(() => {
    if (!match) {
      return -1;
    }
    return match.players.findIndex((player) => player?.user_id === currentUserId);
  }, [currentUserId, match]);

  const player = playerIndex >= 0 ? match?.players[playerIndex] ?? null : null;
  const enemy = playerIndex >= 0 ? match?.players[playerIndex === 0 ? 1 : 0] ?? null : null;
  const canEndTurn =
    Boolean(match) && playerIndex === match?.active_player && match?.phase === "MAIN" && !match?.finished && !busy;

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
          <EnemyCharacter player={enemy as MaskedBattlePlayerState} />
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
          <div className="battle-bottom__side battle-bottom__side--left">
            <GraveyardBlock count={player.discard_count ?? player.discard?.length ?? 0} />
            <AttackBlock attack={player.hero_attack_power} cooldown={player.hero_attack_cooldown} />
          </div>

          <div className="battle-bottom__center">
            <GamerCharacter player={player as MaskedBattlePlayerState} />
            <HandPanel hand={player.hand ?? []} />
          </div>

          <div className="battle-bottom__side battle-bottom__side--right">
            <AbilityBlock cooldown={player.hero_ability_cooldown} manaCost={player.hero_ability_mana_cost ?? 0} />
            <DeckCounter count={player.deck_count ?? player.deck?.length ?? 0} />
          </div>
        </div>

        {error ? <p className="battle-error">{error}</p> : null}
      </div>
    </section>
  );
}

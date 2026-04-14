type Verdict = "idle" | "accepted" | "declined_self" | "declined_opponent" | "countdown";

type Props = {
  open: boolean;
  busy: boolean;
  error: string;
  acceptedByMe: boolean;
  acceptedByOpponent: boolean;
  deadlineSec: number;
  verdict: Verdict;
  countdownSec: number;
  playerRating?: number;
  opponentRating?: number;
  onAccept: () => void;
  onDecline: () => void;
};

function blockState(side: "me" | "opponent", verdict: Verdict, accepted: boolean): string {
  if (verdict === "declined_self" && side === "me") {
    return "match-found-player--declined";
  }
  if (verdict === "declined_opponent" && side === "opponent") {
    return "match-found-player--declined";
  }
  if ((verdict === "accepted" || verdict === "countdown") && accepted) {
    return "match-found-player--accepted";
  }
  return "";
}

export function MatchFoundPanel({
  open,
  busy,
  error,
  acceptedByMe,
  acceptedByOpponent,
  deadlineSec,
  verdict,
  countdownSec,
  playerRating,
  opponentRating,
  onAccept,
  onDecline,
}: Props) {
  if (!open) {
    return null;
  }

  return (
    <div className="overlay match-found-overlay">
      <section className="match-found-panel surface">
        <p className="eyebrow">МАТЧ НАЙДЕН</p>

        <div className="match-found-panel__players">
          <div className={`match-found-player ${blockState("me", verdict, acceptedByMe)}`}>
            <span className="match-found-player__tag">ВЫ</span>
            <div className="match-found-player__avatar" />
            <span className="match-found-player__rating">MMR {playerRating ?? 0}</span>
          </div>

          <div className={`match-found-player ${blockState("opponent", verdict, acceptedByOpponent)}`}>
            <span className="match-found-player__tag">ПРОТИВНИК</span>
            <div className="match-found-player__avatar" />
            <span className="match-found-player__rating">MMR {opponentRating ?? 0}</span>
          </div>
        </div>

        {verdict === "countdown" ? (
          <div className="match-found-panel__countdown">МАТЧ НАЧНЕТСЯ ЧЕРЕЗ {countdownSec}</div>
        ) : (
          <div className="match-found-panel__deadline">ПОДТВЕРЖДЕНИЕ {deadlineSec}С</div>
        )}

        <div className="match-found-panel__actions">
          <button
            type="button"
            className="match-found-panel__accept"
            onClick={onAccept}
            disabled={busy || acceptedByMe || verdict === "declined_self" || verdict === "countdown"}
          >
            В БОЙ
          </button>
          <button
            type="button"
            className="match-found-panel__decline"
            onClick={onDecline}
            disabled={busy || verdict === "declined_self" || verdict === "countdown"}
          >
            ОТКАЗ
          </button>
        </div>

        {error ? <p className="match-found-panel__error">{error}</p> : null}
      </section>
    </div>
  );
}

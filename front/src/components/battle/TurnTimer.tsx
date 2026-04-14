import { useEffect, useMemo, useState } from "react";

type Props = {
  startedAt: number;
  deadlineAt: number;
  totalSec: number;
};

export function TurnTimer({ startedAt, deadlineAt, totalSec }: Props) {
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    const id = window.setInterval(() => setNow(Date.now()), 100);
    return () => window.clearInterval(id);
  }, []);

  const { percent, urgent } = useMemo(() => {
    const started = startedAt * 1000;
    const deadline = deadlineAt * 1000;
    const total = Math.max(1, totalSec * 1000);
    const remaining = Math.max(0, deadline - now);
    const elapsed = Math.min(total, Math.max(0, now - started));
    const basePercent = Math.max(0, 100 - (elapsed / total) * 100);
    return {
      percent: basePercent,
      urgent: remaining <= 5000,
    };
  }, [deadlineAt, now, startedAt, totalSec]);

  return (
    <div className={`battle-turn-timer ${urgent ? "battle-turn-timer--urgent" : ""}`}>
      <div
        className="battle-turn-timer__fill"
        style={{
          clipPath: `inset(0 ${Math.max(0, (100 - percent) / 2)}% 0 ${Math.max(0, (100 - percent) / 2)}%)`,
        }}
      />
    </div>
  );
}

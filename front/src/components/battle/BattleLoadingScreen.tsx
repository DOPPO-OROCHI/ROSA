type Props = {
  progress: number;
  label: string;
  playerReady?: boolean;
  enemyReady?: boolean;
};

export function BattleLoadingScreen({ progress, label, playerReady = false, enemyReady = false }: Props) {
  const percent = Math.round(progress * 100);

  return (
    <section className="battle-loading-screen surface">
      <div className="video-stage battle-loading-screen__video" aria-hidden="true">
        <div className="video-stage__glow" />
        <div className="video-stage__label">battle sync zone</div>
      </div>

      <div className="battle-loading-screen__overlay">
        <p className="battle-loading-screen__eyebrow">PRE BATTLE</p>
        <h1 className="battle-loading-screen__title">ЗАГРУЗКА БОЯ</h1>
        <p className="battle-loading-screen__label">{label}</p>

        <div className="battle-loading-screen__bar">
          <div className="battle-loading-screen__bar-fill" style={{ width: `${percent}%` }} />
        </div>

        <div className="battle-loading-screen__meta">
          <span>Подготавливаем арену</span>
          <span>{percent}%</span>
        </div>

        <div className="battle-loading-screen__meta">
          <span>YOU: {playerReady ? "READY" : "LOADING"}</span>
          <span>ENEMY: {enemyReady ? "READY" : "WAITING"}</span>
        </div>
      </div>
    </section>
  );
}

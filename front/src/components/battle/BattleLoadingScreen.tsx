type Props = {
  progress: number;
  label: string;
};

export function BattleLoadingScreen({ progress, label }: Props) {
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
      </div>
    </section>
  );
}

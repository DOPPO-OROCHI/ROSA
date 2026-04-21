import { useEffect, useMemo, useState } from "react";
import type { ProjectileImpact, ProjectileSnapshot, ProjectileSpread } from "./types";

type Props = {
  projectiles: ProjectileSnapshot[];
  impacts: ProjectileImpact[];
  spreads: ProjectileSpread[];
};

function ProjectileBolt({ projectile }: { projectile: ProjectileSnapshot }) {
  const [progress, setProgress] = useState(0);
  const angle = Math.atan2(projectile.dy, projectile.dx);
  const x = projectile.fromX + projectile.dx * progress;
  const y = projectile.fromY + projectile.dy * progress;

  useEffect(() => {
    let rafId = 0;
    const startedAt = performance.now();

    function frame(now: number) {
      const elapsed = now - startedAt;
      const nextProgress = Math.min(1, elapsed / 180);
      setProgress(nextProgress);
      if (nextProgress < 1) {
        rafId = window.requestAnimationFrame(frame);
      }
    }

    rafId = window.requestAnimationFrame(frame);
    return () => window.cancelAnimationFrame(rafId);
  }, []);

  return (
    <div
      className={`battle-projectile battle-projectile--${projectile.tone}`}
      style={{
        left: `${x}px`,
        top: `${y}px`,
        width: `${projectile.width}px`,
        height: `${projectile.height}px`,
        transform: `translate(-50%, -50%) rotate(${angle}rad)`,
        opacity: progress < 0.08 ? progress / 0.08 : progress > 0.78 ? (1 - progress) / 0.22 : 1,
      }}
    >
      <span className="battle-projectile__tail" />
      <span className="battle-projectile__head" />
    </div>
  );
}

function ProjectileSpreadNode({ spread }: { spread: ProjectileSpread }) {
  const [progress, setProgress] = useState(0);
  const angle = Math.atan2(spread.dy, spread.dx);
  const length = Math.max(24, Math.hypot(spread.dx, spread.dy));

  useEffect(() => {
    let rafId = 0;
    const startedAt = performance.now();

    function frame(now: number) {
      const elapsed = now - startedAt;
      const nextProgress = Math.min(1, elapsed / 40);
      setProgress(nextProgress);
      if (nextProgress < 1) {
        rafId = window.requestAnimationFrame(frame);
      }
    }

    rafId = window.requestAnimationFrame(frame);
    return () => window.cancelAnimationFrame(rafId);
  }, []);

  return (
    <div
      className={`battle-projectile-spread battle-projectile-spread--${spread.tone}`}
      style={{
        left: `${spread.fromX}px`,
        top: `${spread.fromY}px`,
        width: `${length}px`,
        height: `${spread.height}px`,
        transform: `translateY(-50%) rotate(${angle}rad) scaleX(${0.18 + progress * 0.82})`,
        opacity: progress < 0.2 ? progress / 0.2 : 1 - progress * 0.5,
      }}
    >
      <span className="battle-projectile-spread__core" />
      <span className="battle-projectile-spread__tip" />
    </div>
  );
}

function ProjectileImpactNode({ impact }: { impact: ProjectileImpact }) {
  return (
    <>
      <div
        className={`battle-projectile-impact battle-projectile-impact--${impact.tone}`}
        style={{
          left: `${impact.centerX}px`,
          top: `${impact.centerY}px`,
          width: `${impact.width}px`,
          height: `${impact.height}px`,
        }}
      />
      <div
        className={`battle-projectile-card-flash battle-projectile-card-flash--${impact.tone}`}
        style={{
          left: `${impact.flashLeft}px`,
          top: `${impact.flashTop}px`,
          width: `${impact.flashWidth}px`,
          height: `${impact.flashHeight}px`,
        }}
      />
    </>
  );
}

export function ProjectileLayer({ projectiles, impacts, spreads }: Props) {
  const visibleProjectiles = useMemo(() => projectiles, [projectiles]);
  const visibleImpacts = useMemo(() => impacts, [impacts]);
  const visibleSpreads = useMemo(() => spreads, [spreads]);

  if (visibleProjectiles.length === 0 && visibleImpacts.length === 0 && visibleSpreads.length === 0) {
    return null;
  }

  return (
    <div className="battle-projectile-layer" aria-hidden="true">
      {visibleProjectiles.map((projectile) => (
        <ProjectileBolt key={projectile.id} projectile={projectile} />
      ))}
      {visibleSpreads.map((spread) => (
        <ProjectileSpreadNode key={spread.id} spread={spread} />
      ))}
      {visibleImpacts.map((impact) => (
        <ProjectileImpactNode key={impact.id} impact={impact} />
      ))}
    </div>
  );
}

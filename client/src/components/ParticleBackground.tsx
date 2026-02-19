import React, { useEffect, useRef } from 'react';
import './ParticleBackground.css';

const ParticleBackground: React.FC = () => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animationFrameId: number;
    let time = 0;

    const DPR = window.devicePixelRatio || 1;

    const resize = () => {
      const { innerWidth, innerHeight } = window;
      canvas.width = innerWidth * DPR;
      canvas.height = innerHeight * DPR;
      canvas.style.width = `${innerWidth}px`;
      canvas.style.height = `${innerHeight}px`;
      ctx.setTransform(DPR, 0, 0, DPR, 0, 0);
    };

    resize();
    window.addEventListener('resize', resize);

    const draw = () => {
      const width = canvas.width / DPR;
      const height = canvas.height / DPR;

      // Transparent clear – page background shows through; do not draw a fill
      ctx.clearRect(0, 0, width, height);

      const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
      const baseR = isDark ? 255 : 30;
      const baseG = isDark ? 255 : 30;
      const baseB = isDark ? 255 : 30;

      // Grid resolution – denser mesh, larger area
      const columns = 110;
      const rows = 65;

      // Simple camera / perspective parameters
      const cameraZ = 4;
      const waveWidth = 4.5; // horizontal spread in world units
      const depthSpan = 3.8; // how far \"back\" the grid goes
      const baseYWorld = -0.25; // vertical offset of the mesh

      const t = time * 0.001;

      for (let ix = 0; ix <= columns; ix++) {
        for (let iz = 0; iz <= rows; iz++) {
          const u = ix / columns; // -0.5 .. 0.5 across
          const v = iz / rows; // 0 (front) .. 1 (back)

          // World-space grid (X,Z) – regular grid
          const xWorld = (u - 0.5) * waveWidth;
          const zWorld = v * depthSpan;

          // Wave height at this grid point (Y in world space)
          const mainWave =
            Math.sin((u - 0.5) * Math.PI * 4 + t * 1.0) * 0.3 +
            Math.cos(v * Math.PI * 3 - t * 0.7) * 0.2;

          const crossWave =
            Math.sin((u + v) * Math.PI * 3 + t * 1.3) * 0.15 +
            Math.cos((u - v) * Math.PI * 2 - t * 0.6) * 0.1;

          const yWorld = baseYWorld + mainWave + crossWave;

          // Tilt the whole mesh slightly toward the viewer (rotation around X axis)
          const tilt = 0.75; // flipped direction of tilt
          const cosTilt = Math.cos(tilt);
          const sinTilt = Math.sin(tilt);

          const yTilt = yWorld * cosTilt - zWorld * sinTilt;
          const zTilt = yWorld * sinTilt + zWorld * cosTilt;

          // Perspective projection to screen space
          const zCamera = cameraZ + zTilt;
          const perspective = cameraZ / zCamera;

          const x =
            width * 0.5 +
            xWorld * width * 0.28 * perspective;
          const y =
            height * 0.95 +
            yTilt * height * 0.25 * perspective -
            v * height * 0.18;

          // Depth: dots further back are smaller and dimmer
          const depth = 1 - v;
          const radius = 0.5 + depth * 1.8 * perspective;
          const alpha = (0.12 + depth * 0.7 * perspective).toFixed(3);

          ctx.beginPath();
          ctx.fillStyle = `rgba(${baseR}, ${baseG}, ${baseB}, ${alpha})`;
          ctx.arc(x, y, radius, 0, Math.PI * 2);
          ctx.fill();
        }
      }

      time += 16;
      animationFrameId = window.requestAnimationFrame(draw);
    };

    draw();

    return () => {
      window.cancelAnimationFrame(animationFrameId);
      window.removeEventListener('resize', resize);
    };
  }, []);

  return (
    <div className="particle-background" aria-hidden>
      <canvas ref={canvasRef} className="particle-background-canvas" />
    </div>
  );
};

export default ParticleBackground;

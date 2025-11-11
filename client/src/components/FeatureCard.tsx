export function FeatureCard({ title, desc }: { title: string; desc: string }) {
  return (
    <div className="rounded-2xl border border-black/10 dark:border-white/10 p-4 bg-white/60 dark:bg-white/5">
      <h3 className="font-medium mb-1">{title}</h3>
      <p className="text-sm opacity-80">{desc}</p>
    </div>
  );
}

// Главная в маркетинговом стиле; оставлена только одна плитка "Фокус-комнаты".
import { FeatureCard } from "../components/FeatureCard";

export function Home() {
  return (
    <section className="space-y-8">
      <div className="text-center space-y-4">
        <h1 className="text-4xl font-semibold tracking-tight">
          CWRK Planet — стриминговый сервис для совместного обучения
        </h1>
        <p className="text-neutral-600 dark:text-neutral-300 max-w-2xl mx-auto">
          Создавай или подключайся к «фокус-комнатам» с видео/аудио и таймерами, как на стриме — только для учёбы.
          Учись рядом с единомышленниками, сохраняй мотивацию и получай помощь от AI-ассистента.
        </p>
        <div className="flex items-center justify-center gap-3">
          <a href="#/auth" className="rounded-xl border border-black/10 dark:border-white/15 px-5 py-2 text-sm hover:bg-black/5 dark:hover:bg-white/10">Начать</a>
          <a href="#/profile" className="rounded-xl border border-black/10 dark:border-white/15 px-5 py-2 text-sm hover:bg-black/5 dark:hover:bg-white/10">Мой профиль</a>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <FeatureCard
          title="Фокус-комнаты"
          desc="Тихие виртуальные лобби на 5–10 человек — эффект библиотеки и совместной работы."
        />
      </div>
    </section>
  );
}

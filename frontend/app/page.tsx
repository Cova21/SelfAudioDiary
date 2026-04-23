'use client'

import { useState, useEffect } from 'react'
import styles from './page.module.scss'
import Header from './components/Header/Header'
import EntryList from './components/EntryList/EntryList'
import AudioRecorder from './components/AudioRecorder/AudioRecorder'
import Calendar from './components/Calendar/Calendar'

const FEATURES = [
  {
    icon: '🎙️',
    title: 'Голосовые записи',
    desc: 'Записывайте мысли голосом прямо в браузере. Никаких приложений — нажмите кнопку и говорите.',
  },
  {
    icon: '✍️',
    title: 'Автоматическая транскрипция',
    desc: 'AI на базе Whisper мгновенно превращает аудио в текст. Ваши записи всегда можно перечитать.',
  },
  {
    icon: '🧠',
    title: 'Анализ эмоций',
    desc: 'Нейросеть определяет настроение и эмоции каждой записи. Отслеживайте своё состояние.',
  },
  {
    icon: '🔍',
    title: 'Умный поиск',
    desc: 'Полнотекстовый поиск по всем записям. Найдите нужную мысль за секунду.',
  },
  {
    icon: '🔒',
    title: 'Приватность',
    desc: 'Ваши данные принадлежат только вам. Авторизация, шифрование и безопасное хранение.',
  },
  {
    icon: '⚡',
    title: 'Мгновенные уведомления',
    desc: 'Получайте обновления в реальном времени, когда запись транскрибирована и проанализирована.',
  },
]

const STEPS = [
  {
    title: 'Создайте аккаунт',
    desc: 'Быстрая регистрация — и вы готовы начать вести свой дневник.',
  },
  {
    title: 'Запишите мысль голосом',
    desc: 'Нажмите кнопку записи и расскажите, что у вас на уме. Можно сколько угодно.',
  },
  {
    title: 'Получите транскрипцию и анализ',
    desc: 'AI автоматически переведёт речь в текст и определит эмоциональный тон записи.',
  },
  {
    title: 'Перечитывайте и ищите',
    desc: 'Возвращайтесь к своим записям, ищите по ключевым словам и отслеживайте своё настроение.',
  },
]

function WaveBars() {
  const bars = Array.from({ length: 40 }, (_, i) => ({
    height: Math.random() * 22 + 6,
    delay: i * 0.05,
  }))

  return (
    <>
      {bars.map((bar, i) => (
        <span
          key={i}
          className={styles.waveBar}
          style={{
            height: `${bar.height}px`,
            animationDelay: `${bar.delay}s`,
          }}
        />
      ))}
    </>
  )
}

function LandingPage() {
  return (
    <div className={styles.landing}>
      {/* Nav */}
      <nav className={styles.landingNav}>
        <div className={styles.navLogo}>
          <span className={styles.navLogoIcon}>🎙️</span>
          Voice Diary
        </div>
        <div className={styles.navButtons}>
          <a href="/auth/login" className={`${styles.navBtn} ${styles.navBtnGhost}`}>
            Войти
          </a>
          <a href="/auth/register" className={`${styles.navBtn} ${styles.navBtnPrimary}`}>
            Начать бесплатно
          </a>
        </div>
      </nav>

      {/* Hero */}
      <section className={styles.hero}>
        <div className={styles.heroGlow} />

        <div className={styles.heroBadge}>
          ✨ AI-powered голосовой дневник
        </div>

        <h1 className={styles.heroTitle}>
          Ваши мысли — в текст, эмоции — в данные
        </h1>

        <p className={styles.heroSubtitle}>
          Говорите — и нейросеть запишет, расшифрует и проанализирует ваши эмоции.
          Ведите дневник голосом, не набирая ни строчки.
        </p>

        <div className={styles.heroActions}>
          <a href="/auth/register" className={`${styles.btnLarge} ${styles.btnPrimary}`}>
            Попробовать бесплатно
          </a>
          <a href="#features" className={`${styles.btnLarge} ${styles.btnOutline}`}>
            Узнать больше
          </a>
        </div>

        {/* Decorative audio visualization */}
        <div className={styles.heroVisual}>
          <div className={styles.visualBar}>
            <div className={styles.visualDot} />
            <div className={styles.visualWave}>
              <WaveBars />
            </div>
            <span className={styles.visualText}>00:42</span>
          </div>
          <div className={styles.visualBar}>
            <span className={styles.visualText}>
              «Сегодня был отличный день. Наконец-то закончил проект, который откладывал целый месяц...»
            </span>
          </div>
          <div className={styles.visualMeta}>
            <span className={styles.visualTag}>😊 Позитивно</span>
            <span className={styles.visualTag}>🎯 Достижение</span>
            <span className={styles.visualTag}>💪 Мотивация</span>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className={styles.features} id="features">
        <span className={styles.sectionLabel}>Возможности</span>
        <h2 className={styles.sectionTitle}>Всё, что нужно для вашего дневника</h2>
        <p className={styles.sectionSubtitle}>
          Сочетание голосового ввода, искусственного интеллекта и удобного интерфейса.
        </p>

        <div className={styles.featuresGrid}>
          {FEATURES.map((f) => (
            <div key={f.title} className={styles.featureCard}>
              <div className={styles.featureIcon}>{f.icon}</div>
              <h3 className={styles.featureTitle}>{f.title}</h3>
              <p className={styles.featureDesc}>{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* How it works */}
      <section className={styles.howItWorks}>
        <span className={styles.sectionLabel}>Как это работает</span>
        <h2 className={styles.sectionTitle}>Четыре простых шага</h2>
        <p className={styles.sectionSubtitle}>
          От идеи до готовой записи с анализом — за минуту.
        </p>

        <div className={styles.steps}>
          {STEPS.map((s, i) => (
            <div key={s.title} className={styles.step}>
              <div className={styles.stepLine}>
                <div className={styles.stepNumber}>{i + 1}</div>
                {i < STEPS.length - 1 && <div className={styles.stepConnector} />}
              </div>
              <div className={styles.stepContent}>
                <h3 className={styles.stepTitle}>{s.title}</h3>
                <p className={styles.stepDesc}>{s.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <section className={styles.cta}>
        <div className={styles.ctaCard}>
          <div className={styles.ctaGlow} />
          <h2 className={styles.ctaTitle}>Готовы начать вести дневник?</h2>
          <p className={styles.ctaText}>
            Присоединяйтесь и начните записывать свои мысли голосом уже сегодня.
            Это бесплатно и занимает меньше минуты.
          </p>
          <a
            href="/auth/register"
            className={`${styles.btnLarge} ${styles.btnPrimary} ${styles.ctaButton}`}
          >
            Создать аккаунт
          </a>
        </div>
      </section>

      {/* Footer */}
      <footer className={styles.footer}>
        Voice Diary &copy; {new Date().getFullYear()} — AI-powered голосовой дневник
      </footer>
    </div>
  )
}

export default function Home() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [isRecording, setIsRecording] = useState(false)
  const [refreshEntries, setRefreshEntries] = useState(0)
  const [searchQuery, setSearchQuery] = useState('')
  const [activeEmotions, setActiveEmotions] = useState<string[]>([])
  const [activeSentiment, setActiveSentiment] = useState<string | null>(null)
  const [selectedDate, setSelectedDate] = useState<string | null>(null)
  const [entryDates, setEntryDates] = useState<string[]>([])

  useEffect(() => {
    const token = localStorage.getItem('accessToken')
    if (token) {
      setIsAuthenticated(true)
    }
  }, [])

  const handleRecordingComplete = () => {
    setIsRecording(false)
    setRefreshEntries((prev) => prev + 1)
  }

  const handleEmotionToggle = (emotion: string) => {
    setActiveEmotions((prev) =>
      prev.includes(emotion)
        ? prev.filter((e) => e !== emotion)
        : [...prev, emotion]
    )
  }

  const handleSentimentToggle = (sentiment: string) => {
    setActiveSentiment((prev) => (prev === sentiment ? null : sentiment))
  }

  if (!isAuthenticated) {
    return <LandingPage />
  }

  return (
    <div className={styles.container}>
      <Header
        onSearchChange={setSearchQuery}
        onEmotionToggle={handleEmotionToggle}
        onSentimentToggle={handleSentimentToggle}
        activeEmotions={activeEmotions}
        activeSentiment={activeSentiment}
      />

      <main className={styles.main}>
        {!isRecording ? (
          <div className={styles.contentLayout}>
            <div className={styles.sidebar}>
              <button
                className={styles.recordButton}
                onClick={() => setIsRecording(true)}
              >
                <span className={styles.recordIcon}>🎙️</span>
                <span>Новая запись</span>
              </button>

              <Calendar
                entryDates={entryDates}
                selectedDate={selectedDate}
                onSelectDate={setSelectedDate}
              />
            </div>

            <div className={styles.entriesColumn}>
              <EntryList
                key={refreshEntries}
                searchQuery={searchQuery}
                activeEmotions={activeEmotions}
                activeSentiment={activeSentiment}
                selectedDate={selectedDate}
                onEntryDatesChange={setEntryDates}
              />
            </div>
          </div>
        ) : (
          <AudioRecorder
            onComplete={handleRecordingComplete}
            onCancel={() => setIsRecording(false)}
          />
        )}
      </main>
    </div>
  )
}

'use client'

import { useState } from 'react'
import styles from './Header.module.scss'

const EMOTION_FILTERS = [
  { label: '😊 Радость', value: 'радость' },
  { label: '😢 Грусть', value: 'грусть' },
  { label: '😌 Спокойствие', value: 'спокойствие' },
  { label: '😰 Тревога', value: 'тревога' },
  { label: '😠 Гнев', value: 'гнев' },
  { label: '😲 Удивление', value: 'удивление' },
  { label: '✨ Вдохновение', value: 'вдохновение' },
  { label: '😴 Усталость', value: 'усталость' },
  { label: '🙏 Благодарность', value: 'благодарность' },
  { label: '💭 Ностальгия', value: 'ностальгия' },
]

const SENTIMENT_FILTERS = [
  { label: '😊 Позитивное', value: 'positive' },
  { label: '😐 Нейтральное', value: 'neutral' },
  { label: '😔 Негативное', value: 'negative' },
]

interface HeaderProps {
  onSearchChange: (query: string) => void
  onEmotionToggle: (emotion: string) => void
  onSentimentToggle: (sentiment: string) => void
  activeEmotions: string[]
  activeSentiment: string | null
}

export default function Header({
  onSearchChange,
  onEmotionToggle,
  onSentimentToggle,
  activeEmotions,
  activeSentiment,
}: HeaderProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [filtersOpen, setFiltersOpen] = useState(false)

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setSearchQuery(value)
    onSearchChange(value)
  }

  const handleLogout = () => {
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
    window.location.reload()
  }

  const hasActiveFilters = activeEmotions.length > 0 || activeSentiment !== null

  return (
    <header className={styles.header}>
      <div className={styles.container}>
        <div className={styles.logo}>
          <span className={styles.icon}>🎙️</span>
          <span className={styles.title}>Voice Diary</span>
        </div>

        <div className={styles.search}>
          <input
            type="text"
            placeholder="Поиск по тексту записей..."
            value={searchQuery}
            onChange={handleSearchChange}
            className={styles.searchInput}
          />
          <button
            className={`${styles.filterToggle} ${hasActiveFilters ? styles.filterActive : ''}`}
            onClick={() => setFiltersOpen(!filtersOpen)}
          >
            🎛️ {hasActiveFilters && <span className={styles.filterDot} />}
          </button>
        </div>

        <button onClick={handleLogout} className={styles.logoutButton}>
          Выйти
        </button>
      </div>

      {filtersOpen && (
        <div className={styles.filtersPanel}>
          <div className={styles.filtersInner}>
            <div className={styles.filterGroup}>
              <span className={styles.filterLabel}>Настроение</span>
              <div className={styles.filterChips}>
                {SENTIMENT_FILTERS.map((s) => (
                  <button
                    key={s.value}
                    className={`${styles.chip} ${activeSentiment === s.value ? styles.chipActive : ''}`}
                    onClick={() => onSentimentToggle(s.value)}
                  >
                    {s.label}
                  </button>
                ))}
              </div>
            </div>

            <div className={styles.filterGroup}>
              <span className={styles.filterLabel}>Эмоции</span>
              <div className={styles.filterChips}>
                {EMOTION_FILTERS.map((e) => (
                  <button
                    key={e.value}
                    className={`${styles.chip} ${activeEmotions.includes(e.value) ? styles.chipActive : ''}`}
                    onClick={() => onEmotionToggle(e.value)}
                  >
                    {e.label}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </header>
  )
}

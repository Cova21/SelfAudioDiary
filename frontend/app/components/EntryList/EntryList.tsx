'use client'

import { useState, useEffect, useMemo, useRef } from 'react'
import styles from './EntryList.module.scss'

interface Entry {
  id: string
  title: string
  transcription: string
  sentiment: string
  emotions: string[]
  audioFileId: string
  createdAt: string
  status: string | number
}

interface EntryListProps {
  searchQuery?: string
  activeEmotions?: string[]
  activeSentiment?: string | null
  selectedDate?: string | null
  onEntryDatesChange?: (dates: string[]) => void
}

export default function EntryList({
  searchQuery = '',
  activeEmotions = [],
  activeSentiment = null,
  selectedDate = null,
  onEntryDatesChange,
}: EntryListProps) {
  const [entries, setEntries] = useState<Entry[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedEntry, setExpandedEntry] = useState<string | null>(null)
  const [audioUrls, setAudioUrls] = useState<{ [key: string]: string }>({})
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editTitle, setEditTitle] = useState('')
  const [saving, setSaving] = useState(false)
  const editInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    fetchEntries()
  }, [])

  useEffect(() => {
    if (editingId && editInputRef.current) {
      editInputRef.current.focus()
      editInputRef.current.select()
    }
  }, [editingId])

  const fetchEntries = async () => {
    try {
      const token = localStorage.getItem('accessToken')
      const response = await fetch('/api/entries', {
        headers: { 'Authorization': `Bearer ${token}` },
      })
      if (response.ok) {
        const data = await response.json()
        setEntries(data.entries || [])
      }
    } catch (error) {
      console.error('Failed to fetch entries:', error)
    } finally {
      setLoading(false)
    }
  }

  const startRename = (entry: Entry) => {
    setEditingId(entry.id)
    setEditTitle(entry.title)
  }

  const cancelRename = () => {
    setEditingId(null)
    setEditTitle('')
  }

  const saveRename = async (entryId: string) => {
    const trimmed = editTitle.trim()
    if (!trimmed) {
      cancelRename()
      return
    }

    const entry = entries.find((e) => e.id === entryId)
    if (entry && entry.title === trimmed) {
      cancelRename()
      return
    }

    setSaving(true)
    try {
      const token = localStorage.getItem('accessToken')
      const response = await fetch(`/api/entries/${entryId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ title: trimmed }),
      })

      if (response.ok) {
        setEntries((prev) =>
          prev.map((e) => (e.id === entryId ? { ...e, title: trimmed } : e))
        )
      }
    } catch (error) {
      console.error('Failed to rename entry:', error)
    } finally {
      setSaving(false)
      cancelRename()
    }
  }

  const handleEditKeyDown = (e: React.KeyboardEvent, entryId: string) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      saveRename(entryId)
    } else if (e.key === 'Escape') {
      cancelRename()
    }
  }

  useEffect(() => {
    if (onEntryDatesChange && entries.length > 0) {
      const dates = entries.map((e) => {
        const d = new Date(e.createdAt)
        const y = d.getFullYear()
        const m = String(d.getMonth() + 1).padStart(2, '0')
        const day = String(d.getDate()).padStart(2, '0')
        return `${y}-${m}-${day}`
      })
      onEntryDatesChange(dates)
    }
  }, [entries, onEntryDatesChange])

  const filteredEntries = useMemo(() => {
    let result = entries

    if (selectedDate) {
      result = result.filter((entry) => {
        const d = new Date(entry.createdAt)
        const y = d.getFullYear()
        const m = String(d.getMonth() + 1).padStart(2, '0')
        const day = String(d.getDate()).padStart(2, '0')
        return `${y}-${m}-${day}` === selectedDate
      })
    }

    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase()
      result = result.filter((entry) => {
        const inTitle = entry.title?.toLowerCase().includes(q)
        const inTranscription = entry.transcription?.toLowerCase().includes(q)
        return inTitle || inTranscription
      })
    }

    if (activeSentiment) {
      result = result.filter((entry) => entry.sentiment === activeSentiment)
    }

    if (activeEmotions.length > 0) {
      result = result.filter((entry) => {
        if (!entry.emotions || entry.emotions.length === 0) return false
        const entryEmotionsLower = entry.emotions.map((em) => em.toLowerCase())
        return activeEmotions.some((filterEmotion) =>
          entryEmotionsLower.some((em) => em.includes(filterEmotion))
        )
      })
    }

    return result
  }, [entries, searchQuery, activeSentiment, activeEmotions, selectedDate])

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const getStatusText = (status: string | number) => {
    if (typeof status === 'string') {
      const statusMap: { [key: string]: string } = {
        UPLOADED: 'Загружено',
        TRANSCRIBING: 'Транскрибируется',
        TRANSCRIBED: 'Транскрибировано',
        ANALYZING: 'Анализируется',
        COMPLETED: 'Готово',
        FAILED: 'Ошибка',
      }
      return statusMap[status] || 'Обработка...'
    }
    const statuses = [
      'Неизвестно', 'Загружено', 'Транскрибируется',
      'Транскрибировано', 'Анализируется', 'Готово', 'Ошибка',
    ]
    return statuses[status] || 'Неизвестно'
  }

  const toggleExpanded = async (entryId: string, audioFileId: string) => {
    if (expandedEntry === entryId) {
      setExpandedEntry(null)
    } else {
      setExpandedEntry(entryId)
      if (!audioUrls[audioFileId]) {
        const token = localStorage.getItem('accessToken')
        const audioUrl = `/api/storage/${audioFileId}/download?token=${token}`
        setAudioUrls((prev) => ({ ...prev, [audioFileId]: audioUrl }))
      }
    }
  }

  if (loading) {
    return (
      <div className={styles.loading}>
        <div className={styles.spinner}></div>
        <p>Загрузка записей...</p>
      </div>
    )
  }

  if (entries.length === 0) {
    return (
      <div className={styles.empty}>
        <span className={styles.emptyIcon}>📝</span>
        <h3>Пока нет записей</h3>
        <p>Создайте свою первую голосовую запись</p>
      </div>
    )
  }

  const hasFilters = searchQuery.trim() || activeSentiment || activeEmotions.length > 0 || selectedDate

  return (
    <div className={styles.container}>
      <div className={styles.headingRow}>
        <h2 className={styles.heading}>Мои записи</h2>
        {hasFilters && (
          <span className={styles.resultCount}>
            {filteredEntries.length} из {entries.length}
          </span>
        )}
      </div>

      {hasFilters && filteredEntries.length === 0 ? (
        <div className={styles.empty}>
          <span className={styles.emptyIcon}>🔍</span>
          <h3>Ничего не найдено</h3>
          <p>Попробуйте изменить запрос или сбросить фильтры</p>
        </div>
      ) : (
        <div className={styles.list}>
          {filteredEntries.map((entry) => {
            const isExpanded = expandedEntry === entry.id
            const isEditing = editingId === entry.id

            return (
              <div key={entry.id} className={styles.entry}>
                <div className={styles.entryHeader}>
                  {isEditing ? (
                    <div className={styles.editTitleRow}>
                      <input
                        ref={editInputRef}
                        type="text"
                        className={styles.editTitleInput}
                        value={editTitle}
                        onChange={(e) => setEditTitle(e.target.value)}
                        onKeyDown={(e) => handleEditKeyDown(e, entry.id)}
                        onBlur={() => saveRename(entry.id)}
                        disabled={saving}
                        maxLength={200}
                      />
                    </div>
                  ) : (
                    <div className={styles.titleRow}>
                      <h3 className={styles.entryTitle}>{entry.title}</h3>
                      <button
                        className={styles.renameBtn}
                        onClick={() => startRename(entry)}
                        title="Переименовать"
                      >
                        ✏️
                      </button>
                    </div>
                  )}
                  <span className={styles.entryDate}>{formatDate(entry.createdAt)}</span>
                </div>

                {entry.audioFileId && (
                  <div className={styles.audioSection}>
                    <button
                      onClick={() => toggleExpanded(entry.id, entry.audioFileId)}
                      className={styles.expandButton}
                    >
                      {isExpanded ? '🔽 Свернуть' : '▶️ Раскрыть запись'}
                    </button>

                    {isExpanded && (
                      <div className={styles.expandedContent}>
                        {audioUrls[entry.audioFileId] && (
                          <div className={styles.audioPlayer}>
                            <audio
                              controls
                              className={styles.audioElement}
                              src={audioUrls[entry.audioFileId]}
                            >
                              Ваш браузер не поддерживает аудио элемент.
                            </audio>
                          </div>
                        )}

                        {entry.transcription ? (
                          <div className={styles.transcriptionFull}>
                            <h4 className={styles.transcriptionTitle}>📝 Транскрипция:</h4>
                            <p className={styles.transcriptionText}>{entry.transcription}</p>
                          </div>
                        ) : (
                          <p className={styles.processing}>⏳ Идет обработка записи...</p>
                        )}
                      </div>
                    )}
                  </div>
                )}

                {!isExpanded && entry.transcription && (
                  <p className={styles.transcription}>
                    {entry.transcription.length > 150
                      ? entry.transcription.substring(0, 150) + '...'
                      : entry.transcription}
                  </p>
                )}

                <div className={styles.entryFooter}>
                  <div className={styles.metadata}>
                    {entry.sentiment && (
                      <span className={`${styles.badge} ${styles[entry.sentiment]}`}>
                        {entry.sentiment === 'positive'
                          ? '😊 Позитивное'
                          : entry.sentiment === 'negative'
                          ? '😔 Негативное'
                          : '😐 Нейтральное'}
                      </span>
                    )}
                    {entry.emotions &&
                      entry.emotions.length > 0 &&
                      entry.emotions.map((emotion) => (
                        <span key={emotion} className={styles.badge}>
                          {emotion}
                        </span>
                      ))}
                  </div>
                  <span className={styles.status}>{getStatusText(entry.status)}</span>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

'use client'

import { useState, useEffect, useRef } from 'react'
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

export default function EntryList() {
  const [entries, setEntries] = useState<Entry[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedEntry, setExpandedEntry] = useState<string | null>(null)
  const [audioUrls, setAudioUrls] = useState<{ [key: string]: string }>({})

  useEffect(() => {
    fetchEntries()
  }, [])

  const fetchEntries = async () => {
    try {
      const token = localStorage.getItem('accessToken')
      const response = await fetch('/api/entries', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
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

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const getStatusText = (status: string | number) => {
    if (typeof status === 'string') {
      const statusMap: { [key: string]: string } = {
        'UPLOADED': 'Загружено',
        'TRANSCRIBING': 'Транскрибируется',
        'TRANSCRIBED': 'Транскрибировано',
        'ANALYZING': 'Анализируется',
        'COMPLETED': 'Готово',
        'FAILED': 'Ошибка'
      }
      return statusMap[status] || 'Обработка...'
    }
    const statuses = ['Неизвестно', 'Загружено', 'Транскрибируется', 'Транскрибировано', 'Анализируется', 'Готово', 'Ошибка']
    return statuses[status] || 'Неизвестно'
  }
  
  const toggleExpanded = async (entryId: string, audioFileId: string) => {
    if (expandedEntry === entryId) {
      setExpandedEntry(null)
    } else {
      setExpandedEntry(entryId)
      
      // Create authenticated audio URL
      if (!audioUrls[audioFileId]) {
        const token = localStorage.getItem('accessToken')
        // Use direct download endpoint with token in URL for audio element compatibility
        const audioUrl = `/api/storage/${audioFileId}/download?token=${token}`
        setAudioUrls(prev => ({ ...prev, [audioFileId]: audioUrl }))
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

  return (
    <div className={styles.container}>
      <h2 className={styles.heading}>Мои записи</h2>
      
      <div className={styles.list}>
        {entries.map((entry) => {
          const isExpanded = expandedEntry === entry.id
          
          return (
            <div key={entry.id} className={styles.entry}>
              <div className={styles.entryHeader}>
                <h3 className={styles.entryTitle}>{entry.title}</h3>
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
                        <p className={styles.processing}>
                          ⏳ Идет обработка записи...
                        </p>
                      )}
                    </div>
                  )}
                </div>
              )}

              {!isExpanded && entry.transcription && (
                <p className={styles.transcription}>
                  {entry.transcription.length > 150 
                    ? entry.transcription.substring(0, 150) + '...'
                    : entry.transcription
                  }
                </p>
              )}

              <div className={styles.entryFooter}>
                <div className={styles.metadata}>
                  {entry.sentiment && (
                    <span className={`${styles.badge} ${styles[entry.sentiment]}`}>
                      {entry.sentiment === 'positive' ? '😊 Позитивно' : '😔 Негативно'}
                    </span>
                  )}
                  {entry.emotions && entry.emotions.length > 0 && (
                    entry.emotions.map((emotion) => (
                      <span key={emotion} className={styles.badge}>
                        {emotion}
                      </span>
                    ))
                  )}
                </div>
                <span className={styles.status}>{getStatusText(entry.status)}</span>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

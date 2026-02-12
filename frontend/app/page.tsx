'use client'

import { useState, useEffect } from 'react'
import styles from './page.module.scss'
import Header from './components/Header/Header'
import EntryList from './components/EntryList/EntryList'
import AudioRecorder from './components/AudioRecorder/AudioRecorder'

export default function Home() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [isRecording, setIsRecording] = useState(false)
  const [refreshEntries, setRefreshEntries] = useState(0)

  useEffect(() => {
    // Check if user is authenticated
    const token = localStorage.getItem('accessToken')
    if (token) {
      setIsAuthenticated(true)
    }
  }, [])

  const handleRecordingComplete = () => {
    setIsRecording(false)
    setRefreshEntries(prev => prev + 1)
  }

  if (!isAuthenticated) {
    return (
      <div className={styles.container}>
        <div className={styles.authContainer}>
          <h1 className={styles.logo}>🎙️ Voice Diary</h1>
          <p className={styles.subtitle}>Голосовой дневник с AI-анализом</p>
          <div className={styles.authButtons}>
            <a href="/auth/login" className={styles.button}>
              Войти
            </a>
            <a href="/auth/register" className={styles.buttonOutline}>
              Регистрация
            </a>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <Header />
      
      <main className={styles.main}>
        {!isRecording ? (
          <>
            <div className={styles.recordSection}>
              <button
                className={styles.recordButton}
                onClick={() => setIsRecording(true)}
              >
                <span className={styles.recordIcon}>🎙️</span>
                <span>Новая запись</span>
              </button>
            </div>

            <EntryList key={refreshEntries} />
          </>
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

'use client'

import { useState, useRef } from 'react'
import styles from './AudioRecorder.module.scss'

interface AudioRecorderProps {
  onComplete: () => void
  onCancel: () => void
}

export default function AudioRecorder({ onComplete, onCancel }: AudioRecorderProps) {
  const [isRecording, setIsRecording] = useState(false)
  const [recordingTime, setRecordingTime] = useState(0)
  const [isUploading, setIsUploading] = useState(false)
  const mediaRecorderRef = useRef<MediaRecorder | null>(null)
  const chunksRef = useRef<Blob[]>([])
  const timerRef = useRef<NodeJS.Timeout | null>(null)

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const mediaRecorder = new MediaRecorder(stream)
      
      mediaRecorderRef.current = mediaRecorder
      chunksRef.current = []

      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) {
          chunksRef.current.push(e.data)
        }
      }

      mediaRecorder.start()
      setIsRecording(true)

      // Start timer
      timerRef.current = setInterval(() => {
        setRecordingTime((prev) => prev + 1)
      }, 1000)

    } catch (error) {
      console.error('Failed to start recording:', error)
      alert('Не удалось получить доступ к микрофону')
    }
  }

  const stopRecording = () => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop()
      setIsRecording(false)

      if (timerRef.current) {
        clearInterval(timerRef.current)
      }

      mediaRecorderRef.current.onstop = () => {
        uploadRecording()
      }
    }
  }

  const uploadRecording = async () => {
    setIsUploading(true)

    try {
      const audioBlob = new Blob(chunksRef.current, { type: 'audio/webm' })
      const formData = new FormData()
      formData.append('audio', audioBlob, 'recording.webm')
      formData.append('title', `Запись от ${new Date().toLocaleString('ru-RU')}`)
      formData.append('duration_seconds', recordingTime.toString())

      const token = localStorage.getItem('accessToken')
      const response = await fetch('/api/entries/upload', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData
      })

      if (response.ok) {
        onComplete()
      } else {
        alert('Ошибка при загрузке записи')
      }
    } catch (error) {
      console.error('Upload failed:', error)
      alert('Ошибка при загрузке записи')
    } finally {
      setIsUploading(false)
    }
  }

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }

  return (
    <div className={styles.container}>
      <div className={styles.recorder}>
        <h2 className={styles.title}>
          {isRecording ? 'Идет запись...' : isUploading ? 'Загрузка...' : 'Готовы к записи'}
        </h2>

        <div className={styles.visualizer}>
          {isRecording && (
            <div className={styles.pulse}></div>
          )}
          <span className={styles.icon}>🎙️</span>
        </div>

        <div className={styles.time}>{formatTime(recordingTime)}</div>

        <div className={styles.controls}>
          {!isRecording && !isUploading && (
            <>
              <button onClick={startRecording} className={styles.recordButton}>
                Начать запись
              </button>
              <button onClick={onCancel} className={styles.cancelButton}>
                Отмена
              </button>
            </>
          )}

          {isRecording && (
            <button onClick={stopRecording} className={styles.stopButton}>
              Остановить запись
            </button>
          )}

          {isUploading && (
            <div className={styles.uploading}>
              <div className={styles.spinner}></div>
              <span>Загрузка записи...</span>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

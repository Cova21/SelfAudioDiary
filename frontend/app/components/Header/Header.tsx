'use client'

import { useState } from 'react'
import styles from './Header.module.scss'

export default function Header() {
  const [searchQuery, setSearchQuery] = useState('')

  const handleLogout = () => {
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
    window.location.reload()
  }

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
            placeholder="Поиск по записям..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className={styles.searchInput}
          />
        </div>

        <button onClick={handleLogout} className={styles.logoutButton}>
          Выйти
        </button>
      </div>
    </header>
  )
}

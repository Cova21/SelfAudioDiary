'use client'

import { useState, useMemo } from 'react'
import styles from './Calendar.module.scss'

const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']
const MONTHS = [
  'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
  'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь',
]

interface CalendarProps {
  entryDates: string[]
  selectedDate: string | null
  onSelectDate: (date: string | null) => void
}

function toDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export default function Calendar({ entryDates, selectedDate, onSelectDate }: CalendarProps) {
  const today = new Date()
  const [viewYear, setViewYear] = useState(today.getFullYear())
  const [viewMonth, setViewMonth] = useState(today.getMonth())

  const dateSet = useMemo(() => new Set(entryDates), [entryDates])

  const daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate()
  const firstDayOfWeek = (new Date(viewYear, viewMonth, 1).getDay() + 6) % 7

  const prevMonth = () => {
    if (viewMonth === 0) {
      setViewMonth(11)
      setViewYear(viewYear - 1)
    } else {
      setViewMonth(viewMonth - 1)
    }
  }

  const nextMonth = () => {
    if (viewMonth === 11) {
      setViewMonth(0)
      setViewYear(viewYear + 1)
    } else {
      setViewMonth(viewMonth + 1)
    }
  }

  const goToday = () => {
    setViewYear(today.getFullYear())
    setViewMonth(today.getMonth())
  }

  const handleDateClick = (dateKey: string) => {
    onSelectDate(selectedDate === dateKey ? null : dateKey)
  }

  const todayKey = toDateKey(today)
  const cells: (number | null)[] = []
  for (let i = 0; i < firstDayOfWeek; i++) cells.push(null)
  for (let d = 1; d <= daysInMonth; d++) cells.push(d)

  const entryCount = (dateKey: string) => entryDates.filter((d) => d === dateKey).length

  return (
    <div className={styles.calendar}>
      <div className={styles.header}>
        <button className={styles.navBtn} onClick={prevMonth}>‹</button>
        <button className={styles.monthLabel} onClick={goToday}>
          {MONTHS[viewMonth]} {viewYear}
        </button>
        <button className={styles.navBtn} onClick={nextMonth}>›</button>
      </div>

      <div className={styles.weekdays}>
        {WEEKDAYS.map((wd) => (
          <span key={wd} className={styles.weekday}>{wd}</span>
        ))}
      </div>

      <div className={styles.grid}>
        {cells.map((day, i) => {
          if (day === null) {
            return <span key={`e-${i}`} className={styles.emptyCell} />
          }

          const dateKey = `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
          const hasEntries = dateSet.has(dateKey)
          const isToday = dateKey === todayKey
          const isSelected = dateKey === selectedDate
          const count = hasEntries ? entryCount(dateKey) : 0

          const cellClass = [
            styles.dayCell,
            isToday ? styles.today : '',
            hasEntries ? styles.hasEntries : '',
            isSelected ? styles.selected : '',
          ].filter(Boolean).join(' ')

          return (
            <button
              key={dateKey}
              className={cellClass}
              onClick={() => hasEntries ? handleDateClick(dateKey) : undefined}
              disabled={!hasEntries}
            >
              <span className={styles.dayNumber}>{day}</span>
              {hasEntries && (
                <span className={styles.dotRow}>
                  {count >= 1 && <span className={styles.dot} />}
                  {count >= 2 && <span className={styles.dot} />}
                  {count >= 3 && <span className={styles.dot} />}
                </span>
              )}
            </button>
          )
        })}
      </div>

      {selectedDate && (
        <button className={styles.clearBtn} onClick={() => onSelectDate(null)}>
          Сбросить дату
        </button>
      )}
    </div>
  )
}

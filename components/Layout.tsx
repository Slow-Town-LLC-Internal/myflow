// components/Layout.tsx
import React from 'react'
import { useTheme } from 'next-themes'
import Link from 'next/link'

export default function Layout({ children }) {
  const { theme, setTheme } = useTheme()

  return (
    <div className="min-h-screen bg-white dark:bg-gray-900">
      <header className="border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-4xl mx-auto px-4 py-4 flex justify-between items-center">
          <Link href="/" className="text-xl font-semibold text-gray-800 dark:text-gray-200">
            MyFlow
          </Link>
          
          <nav className="flex items-center gap-6">
            <Link href="/docs/about" className="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white">
              About
            </Link>
            <button
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              className="p-2 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200"
            >
              {theme === 'dark' ? '🌞' : '🌙'}
            </button>
          </nav>
        </div>
      </header>
      <main className="max-w-4xl mx-auto px-4">
        {children}
      </main>
    </div>
  )
}

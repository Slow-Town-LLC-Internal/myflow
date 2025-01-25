// layouts/default.tsx
import React from 'react'

export default function DefaultLayout({ children }) {
  return (
    <div className="prose max-w-3xl mx-auto p-8">
      {children}
    </div>
  )
}
import Link from 'next/link'

export default function Home() {
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold">MyFlow Documentation</h1>
      <div className="mt-4">
        <Link href="/docs/about" className="text-blue-500 hover:underline">
          About
        </Link>
      </div>
    </div>
  )
}
// pages/docs/[...slug].tsx
import React from 'react'
import fs from 'fs'
import path from 'path'
import * as Markdoc from '@markdoc/markdoc'

export default function DocsPage({ html }) {
  return (
    <article className="prose max-w-4xl mx-auto p-8" 
             dangerouslySetInnerHTML={{ __html: html }} />
  )
}

export async function getStaticProps({ params }) {
  const filePath = path.join(process.cwd(), 'docs', `${params.slug.join('/')}.md`)
  const source = fs.readFileSync(filePath, 'utf8')
  const ast = Markdoc.parse(source)
  const content = Markdoc.transform(ast)
  const html = Markdoc.renderers.html(content)

  return { props: { html } }
}

export async function getStaticPaths() {
  return { paths: [], fallback: 'blocking' }
}

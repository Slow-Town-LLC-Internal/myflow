
// pages/[...slug].tsx
import { getMarkdocPageContent } from '@markdoc/next.js/runtime'

export default function MarkdocPage(props) {
  return getMarkdocPageContent(props)
}

export const getStaticProps = getMarkdocPageContent.getStaticProps
export const getStaticPaths = getMarkdocPageContent.getStaticPaths

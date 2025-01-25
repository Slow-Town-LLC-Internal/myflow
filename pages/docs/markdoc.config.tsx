// pages/docs/markdoc.config.ts
import React from 'react';

export const nodes = {
  heading: {
    render: 'h2',
    attributes: {
      level: { type: Number, required: true }
    }
  },
  paragraph: { render: 'p' },
  link: { 
    render: 'a',
    attributes: {
      href: { type: String, required: true }
    }
  }
}

export const tags = {
  callout: {
    render: 'Callout',
    attributes: {
      type: { type: String, default: 'note' }
    }
  }
}

interface CalloutProps {
  type: string;
  children: React.ReactNode;
}

const Callout: React.FC<CalloutProps> = ({ type, children }) => (
  <div className={`callout callout-${type}`}>
    {children}
  </div>
)

export const components = {
  Callout
}

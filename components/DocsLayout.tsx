import React from 'react';

interface DocsLayoutProps {
  children: React.ReactNode;
}

const DocsLayout: React.FC<DocsLayoutProps> = ({ children }) => {
  return (
    <div className="prose max-w-3xl mx-auto p-8">
      {children}
    </div>
  );
};

export default DocsLayout;

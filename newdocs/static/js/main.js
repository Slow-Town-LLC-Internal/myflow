document.addEventListener('DOMContentLoaded', () => {
  // Mobile sidebar toggle
  const sidebarToggle = document.querySelector('.sidebar-toggle');
  const sidebar = document.querySelector('.sidebar');
  
  if (sidebarToggle && sidebar) {
    sidebarToggle.addEventListener('click', () => {
      sidebar.classList.toggle('open');
    });
    
    // Close sidebar when clicking outside
    document.addEventListener('click', (e) => {
      if (sidebar.classList.contains('open') && 
          !sidebar.contains(e.target) && 
          e.target !== sidebarToggle) {
        sidebar.classList.remove('open');
      }
    });
  }
  
  // Section expand/collapse
  const sectionTitles = document.querySelectorAll('.section-title');
  
  sectionTitles.forEach(title => {
    title.addEventListener('click', () => {
      const section = title.parentElement;
      const items = section.querySelector('.section-items');
      
      // Toggle active class on section
      section.classList.toggle('active');
      
      // Toggle expanded class on items
      if (items) {
        items.classList.toggle('expanded');
      }
    });
  });
  
  // Add copy button to code blocks
  const codeBlocks = document.querySelectorAll('pre code');
  
  codeBlocks.forEach(block => {
    const copyButton = document.createElement('button');
    copyButton.className = 'copy-button';
    copyButton.textContent = 'Copy';
    
    // Add button to pre element (code block container)
    const pre = block.parentElement;
    pre.style.position = 'relative';
    pre.appendChild(copyButton);
    
    copyButton.addEventListener('click', () => {
      navigator.clipboard.writeText(block.textContent)
        .then(() => {
          copyButton.textContent = 'Copied!';
          setTimeout(() => {
            copyButton.textContent = 'Copy';
          }, 2000);
        })
        .catch(err => {
          console.error('Failed to copy code: ', err);
          copyButton.textContent = 'Failed';
          setTimeout(() => {
            copyButton.textContent = 'Copy';
          }, 2000);
        });
    });
  });
  
  // Highlight active link based on current page
  const currentPath = window.location.pathname;
  const navLinks = document.querySelectorAll('.nav-link');
  
  navLinks.forEach(link => {
    if (link.getAttribute('href') === currentPath) {
      link.parentElement.classList.add('active');
      
      // Expand the parent section
      const section = link.closest('.nav-section');
      if (section) {
        section.classList.add('active');
        const items = section.querySelector('.section-items');
        if (items) {
          items.classList.add('expanded');
        }
      }
    }
  });
  
  // Search functionality removed for now
  // We'll implement proper search in a future version
});

// Add CSS styles for copy button
document.head.insertAdjacentHTML('beforeend', `
  <style>
    .copy-button {
      position: absolute;
      top: 0.5rem;
      right: 0.5rem;
      padding: 0.25rem 0.5rem;
      font-size: 0.75rem;
      background-color: var(--accent);
      color: white;
      border: none;
      border-radius: 4px;
      opacity: 0.8;
      cursor: pointer;
      transition: opacity var(--transition-fast);
    }
    
    .copy-button:hover {
      opacity: 1;
    }
    
    pre:hover .copy-button {
      display: block;
    }
    
    pre .copy-button {
      display: none;
    }
  </style>
`);

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
  
  // Directory navigation
  const directoryLinks = document.querySelectorAll('.directory-link');
  directoryLinks.forEach(link => {
    link.addEventListener('click', (e) => {
      e.preventDefault();
      const dirPath = link.getAttribute('data-dir-path');
      
      // Get all page links
      const navLinks = document.querySelectorAll('.nav-link:not(.directory-link):not(.tag-link)');
      
      // Hide all links first
      navLinks.forEach(pageLink => {
        pageLink.parentElement.style.display = 'none';
      });
      
      // Show only links from the selected directory
      if (dirPath) {
        // Show links from this specific directory
        navLinks.forEach(pageLink => {
          const pagePath = pageLink.getAttribute('href');
          if (pagePath && pagePath.includes('/' + dirPath + '/')) {
            pageLink.parentElement.style.display = 'block';
          }
        });
      } else {
        // Show links from the root directory only (not in subdirectories)
        navLinks.forEach(pageLink => {
          const pagePath = pageLink.getAttribute('href');
          // Show if not in a subdirectory (no slash after the first slash)
          if (pagePath && pagePath.lastIndexOf('/') === 0) {
            pageLink.parentElement.style.display = 'block';
          }
        });
      }
      
      // Store current directory in sessionStorage
      sessionStorage.setItem('currentDirectory', dirPath || '');
      
      // Update UI to show current directory
      document.querySelectorAll('.directory-link').forEach(dirLink => {
        dirLink.classList.remove('current-dir');
      });
      link.classList.add('current-dir');
    });
  });
  
  // Tag filtering
  const tagLinks = document.querySelectorAll('.tag-link');
  
  // Create a container for tagged pages
  const taggedPagesContainer = document.createElement('div');
  taggedPagesContainer.id = 'tagged-pages-container';
  taggedPagesContainer.className = 'tagged-pages-container';
  
  // Add it after the sidebar
  const sidebar = document.querySelector('.sidebar');
  if (sidebar && !document.getElementById('tagged-pages-container')) {
    sidebar.parentNode.insertBefore(taggedPagesContainer, sidebar.nextSibling);
  }
  
  tagLinks.forEach(link => {
    link.addEventListener('click', (e) => {
      e.preventDefault();
      const tag = link.getAttribute('data-tag');
      
      // Show the tag filter container
      taggedPagesContainer.style.display = 'block';
      
      // Create the filter message and header
      taggedPagesContainer.innerHTML = `
        <div class="tag-filter-header">
          <h3>Pages tagged with: <span class="tag-name">${tag}</span></h3>
          <button id="close-tag-pages" class="close-button">✕</button>
        </div>
        <div class="tagged-pages-list"></div>
      `;
      
      // Now we need to find all pages with this tag
      // We'll do this by checking all page elements for data-tags attribute
      const allPages = document.querySelectorAll('[data-tags]');
      const taggedPagesList = taggedPagesContainer.querySelector('.tagged-pages-list');
      let foundPages = 0;
      
      allPages.forEach(page => {
        const tagsAttr = page.getAttribute('data-tags');
        if (tagsAttr) {
          const pageTags = tagsAttr.split(',').filter(t => t.trim() !== ''); // Filter out empty strings
          if (pageTags.includes(tag)) {
            // Clone the link element to display in our filtered list
            const pageTitle = page.textContent;
            const pageUrl = page.getAttribute('href');
            
            const pageLink = document.createElement('a');
            pageLink.href = pageUrl;
            pageLink.className = 'tagged-page-link';
            pageLink.textContent = pageTitle;
            
            const pageItem = document.createElement('div');
            pageItem.className = 'tagged-page-item';
            pageItem.appendChild(pageLink);
            
            taggedPagesList.appendChild(pageItem);
            foundPages++;
          }
        }
      });
      
      // If no pages found, display a message
      if (foundPages === 0) {
        taggedPagesList.innerHTML = '<p class="no-pages-found">No pages found with this tag.</p>';
      }
      
      // Add close button functionality
      document.getElementById('close-tag-pages').addEventListener('click', () => {
        taggedPagesContainer.style.display = 'none';
      });
    });
  });
  
  // Restore directory filter from session storage
  const savedDir = sessionStorage.getItem('currentDirectory');
  if (savedDir !== null) {
    const dirSelector = savedDir ? 
      `[data-dir-path="${savedDir}"]` : 
      '[data-dir-path=""]';
    
    const savedDirLink = document.querySelector(dirSelector);
    if (savedDirLink) {
      // Trigger click to apply the filter
      savedDirLink.click();
    }
  }
  
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
});

// Add CSS styles for copy button and navigation elements
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
    
    /* Directory and tag navigation */
    .directory-link, .tag-link {
      display: flex !important;
      align-items: center;
      gap: 0.5rem;
    }
    
    .directory-link svg, .tag-link svg {
      flex-shrink: 0;
    }
    
    .directory-link.current-dir {
      font-weight: bold;
      color: var(--accent);
    }
    
    #tag-filter-msg {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    
    #clear-tag-filter {
      background: transparent;
      border: none;
      cursor: pointer;
      font-size: 1rem;
      color: var(--text);
      opacity: 0.7;
    }
    
    #clear-tag-filter:hover {
      opacity: 1;
    }
    
    /* Folder and tag icons */
    .folder-icon, .tag-icon {
      color: var(--accent);
    }
    
    /* Tagged pages container */
    .tagged-pages-container {
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      background-color: var(--bg);
      border: 1px solid var(--border);
      border-radius: 6px;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      padding: 1rem;
      max-width: 90%;
      width: 500px;
      max-height: 80vh;
      overflow-y: auto;
      z-index: 1000;
      display: none;
    }
    
    .tag-filter-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1rem;
      padding-bottom: 0.5rem;
      border-bottom: 1px solid var(--border);
    }
    
    .tag-filter-header h3 {
      margin: 0;
      font-size: 1.2rem;
    }
    
    .tag-name {
      color: var(--accent);
      font-weight: bold;
    }
    
    .close-button {
      background: transparent;
      border: none;
      cursor: pointer;
      font-size: 1.2rem;
      color: var(--text);
      padding: 5px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      width: 30px;
      height: 30px;
    }
    
    .close-button:hover {
      background-color: var(--bg-light);
    }
    
    .tagged-pages-list {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }
    
    .tagged-page-item {
      padding: 0.5rem;
      border-radius: 4px;
    }
    
    .tagged-page-item:hover {
      background-color: var(--bg-light);
    }
    
    .tagged-page-link {
      text-decoration: none;
      color: var(--text);
      display: block;
    }
    
    .tagged-page-link:hover {
      color: var(--accent);
    }
    
    .no-pages-found {
      color: var(--text-muted);
      font-style: italic;
      text-align: center;
      padding: 1rem;
    }
  </style>
`);

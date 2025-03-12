/**
 * Simple Photo Gallery with Lightbox
 * This should be added as a separate file: gallery.js
 */
(function() {
  // Wait for DOM to load
  document.addEventListener('DOMContentLoaded', initGallery);

  function initGallery() {
    const galleryContainer = document.getElementById('gallery-container');
    if (!galleryContainer) return;

    // Create lightbox elements
    const lightbox = createLightbox();
    document.body.appendChild(lightbox);

    // Configuration
    const imagesPath = '/images/'; // Path to your images directory
    
    // Load images
    loadGalleryImages(galleryContainer, lightbox);
  }

  function createLightbox() {
    const lightbox = document.createElement('div');
    lightbox.className = 'lightbox';
    lightbox.innerHTML = `
      <div class="lightbox-header">
        <h3 class="lightbox-title"></h3>
        <button class="lightbox-close">&times;</button>
      </div>
      <div class="lightbox-content">
        <img class="lightbox-img" src="" alt="">
        <div class="lightbox-nav lightbox-prev">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M15 18L9 12L15 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="lightbox-nav lightbox-next">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M9 18L15 12L9 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
      </div>
      <div class="lightbox-footer">
        <span class="lightbox-description"></span>
        <span class="lightbox-counter"></span>
      </div>
    `;

    // Close lightbox when clicking the close button
    const closeBtn = lightbox.querySelector('.lightbox-close');
    closeBtn.addEventListener('click', () => {
      lightbox.classList.remove('active');
      document.body.style.overflow = '';
    });

    // Close lightbox when clicking outside the image
    lightbox.addEventListener('click', (e) => {
      if (e.target === lightbox || e.target.classList.contains('lightbox-content')) {
        lightbox.classList.remove('active');
        document.body.style.overflow = '';
      }
    });

    // Handle keyboard navigation
    document.addEventListener('keydown', (e) => {
      if (!lightbox.classList.contains('active')) return;
      
      if (e.key === 'Escape') {
        lightbox.classList.remove('active');
        document.body.style.overflow = '';
      } else if (e.key === 'ArrowLeft') {
        lightbox.querySelector('.lightbox-prev').click();
      } else if (e.key === 'ArrowRight') {
        lightbox.querySelector('.lightbox-next').click();
      }
    });

    return lightbox;
  }

  function loadGalleryImages(container, lightbox) {
    // Clear loading message
    container.innerHTML = '';
    
    // Define our images array
    const images = [];
    
    // Request the images directory
    fetch('/images/')
      .then(response => {
        if (!response.ok) {
          // If directory listing is not supported, check for images directly
          scanImageFiles();
          return null;
        }
        return response.text();
      })
      .then(html => {
        if (!html) return;
        
        // Parse directory listing
        const parser = new DOMParser();
        const doc = parser.parseFromString(html, 'text/html');
        const links = doc.querySelectorAll('a');
        
        links.forEach(link => {
          const href = link.getAttribute('href');
          if (href && /\.(jpg|jpeg|png|gif|webp)$/i.test(href)) {
            images.push({
              src: '/images/' + href,
              title: href.replace(/\.[^/.]+$/, '').replace(/_/g, ' '),
              date: ''
            });
          }
        });
        
        if (images.length === 0) {
          scanImageFiles();
        } else {
          renderGallery(images);
        }
      })
      .catch(error => {
        console.error('Error loading images:', error);
        scanImageFiles();
      });
    
    // Try to scan for image files with predictable names
    function scanImageFiles() {
      console.log('Directory listing not available, looking for images directly...');
      
      // Try common image patterns
      const patterns = [
        { prefix: '', start: 1, end: 50, format: 'vsco_%d.jpg' },
        { prefix: '', start: 1, end: 50, format: '%03d_vsco.jpg' },
        { prefix: '', start: 1, end: 50, format: 'photo%d.jpg' }
      ];
      
      let imagesFound = false;
      let checksCompleted = 0;
      
      patterns.forEach(pattern => {
        for (let i = pattern.start; i <= pattern.end; i++) {
          const filename = pattern.format.replace('%d', i).replace('%03d', i.toString().padStart(3, '0'));
          const imgPath = `/images/${filename}`;
          
          // Create a test image to see if the file exists
          const img = new Image();
          img.onload = function() {
            imagesFound = true;
            images.push({
              src: imgPath,
              title: `Photo ${i}`,
              date: ''
            });
            
            // Render once we've found some images
            if (images.length % 5 === 0) {
              renderGallery(images);
            }
          };
          
          img.onerror = function() {
            checksCompleted++;
            
            // If we've tried all patterns and found nothing, use placeholders
            const totalChecks = patterns.reduce((sum, p) => sum + (p.end - p.start + 1), 0);
            if (checksCompleted >= totalChecks && !imagesFound) {
              usePlaceholders();
            }
          };
          
          img.src = imgPath;
        }
      });
      
      // Give it a moment to find real images
      setTimeout(() => {
        if (images.length === 0) {
          usePlaceholders();
        }
      }, 2000);
    }
    
    // Use placeholder images as last resort
    function usePlaceholders() {
      console.log('No images found, using placeholders');
      if (images.length > 0) return; // Don't add placeholders if we found real images
      
      container.innerHTML = `
        <div style="grid-column: 1/-1; text-align: center; padding: 2rem;">
          <p>No photos found in the images directory. Add your photos to the "docs/images" folder.</p>
        </div>
      `;
    }
    
    function renderGallery(images) {
      // Clear container first
      container.innerHTML = '';
      
      images.forEach((image, index) => {
        const item = document.createElement('div');
        item.className = 'gallery-item';
        item.innerHTML = `
          <img src="${image.src}" alt="${image.title}" loading="lazy">
          <div class="overlay">
            <div class="image-title">${image.title}</div>
            ${image.date ? `<div class="image-date">${image.date}</div>` : ''}
          </div>
        `;
        
        // Open lightbox when clicking on an image
        item.addEventListener('click', () => openLightbox(index));
        
        container.appendChild(item);
      });
      
      // Configure lightbox navigation
      function openLightbox(index) {
        const image = images[index];
        const imgElement = lightbox.querySelector('.lightbox-img');
        const title = lightbox.querySelector('.lightbox-title');
        const description = lightbox.querySelector('.lightbox-description');
        const counter = lightbox.querySelector('.lightbox-counter');
        
        imgElement.src = image.src;
        imgElement.alt = image.title;
        title.textContent = image.title;
        description.textContent = image.date || '';
        counter.textContent = `${index + 1} / ${images.length}`;
        
        lightbox.classList.add('active');
        document.body.style.overflow = 'hidden';
        
        // Configure navigation buttons
        const prevBtn = lightbox.querySelector('.lightbox-prev');
        const nextBtn = lightbox.querySelector('.lightbox-next');
        
        prevBtn.onclick = () => {
          const newIndex = (index - 1 + images.length) % images.length;
          openLightbox(newIndex);
        };
        
        nextBtn.onclick = () => {
          const newIndex = (index + 1) % images.length;
          openLightbox(newIndex);
        };
      }
    }
  }
})();

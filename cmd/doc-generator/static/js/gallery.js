/**
 * Simple Photo Gallery with Lightbox
 * Fixed version with deduplication
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

    // Use a Set to track unique image URLs
    const imageSources = new Set();
    const images = [];

    // Define base path
    const basePath = '/images/'; // Default path

    // First try pattern with oldvsco prefix
    scanSpecificPattern();

    function scanSpecificPattern() {
      console.log('Looking for images with oldvsco pattern...');
      let found = false;
      const date = "20250312"; // Your date pattern

      // Try for a reasonable range of images (1-100)
      for (let i = 1; i <= 100; i++) {
        const filename = `oldvsco-${date}-${String(i).padStart(3, '0')}.jpg`;
        const imgPath = `${basePath}${filename}`;

        // We'll use an image check approach
        const img = new Image();

        img.onload = function() {
          // Found an image with this pattern
          if (!imageSources.has(imgPath)) {
            imageSources.add(imgPath);

            images.push({
              src: imgPath,
              title: `Photo ${i}`,
              date: formatDate(date)
            });

            found = true;

            // Render gallery after finding each batch of images
            if (images.length % 5 === 0) {
              renderGallery(images);
            }
          }
        };

        img.onerror = function() {
          // If we've checked a significant number and found nothing, try other methods
          if (i >= 20 && !found) {
            fallbackToGenericScan();
            return;
          }
        };

        img.src = imgPath;
      }

      // If no images found after some time, try next method
      setTimeout(() => {
        if (!found) {
          fallbackToGenericScan();
        }
      }, 1500);
    }

    function fallbackToGenericScan() {
      console.log('Trying generic image patterns...');

      // Common patterns to try
      const patterns = [
        { prefix: '', start: 1, end: 50, format: 'vsco_%d.jpg' },
        { prefix: '', start: 1, end: 50, format: '%03d_vsco.jpg' },
        { prefix: '', start: 1, end: 50, format: 'photo%d.jpg' },
        { prefix: '', start: 1, end: 50, format: 'vsco-%d.jpg' }
      ];

      let foundAny = false;

      patterns.forEach(pattern => {
        for (let i = pattern.start; i <= pattern.end; i++) {
          const filename = pattern.format.replace('%d', i).replace('%03d', String(i).padStart(3, '0'));
          const imgPath = `${basePath}${filename}`;

          // Skip if we've already found this image
          if (imageSources.has(imgPath)) continue;

          // Check if image exists
          const img = new Image();

          img.onload = function() {
            // Found an image
            imageSources.add(imgPath);

            images.push({
              src: imgPath,
              title: `Photo ${i}`,
              date: ''
            });

            foundAny = true;

            // Render gallery after finding each batch of images
            if (images.length % 5 === 0) {
              renderGallery(images);
            }
          };

          img.src = imgPath;
        }
      });

      // If still no images, try directory listing
      setTimeout(() => {
        if (!foundAny && images.length === 0) {
          tryDirectoryListing();
        } else {
          renderGallery(images);
        }
      }, 2000);
    }

    function tryDirectoryListing() {
      console.log('Trying directory listing...');

      fetch(basePath)
        .then(response => {
          if (!response.ok) {
            displayNoImagesMessage();
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
              const imgPath = basePath + href;

              // Skip if we've already found this image
              if (imageSources.has(imgPath)) return;

              imageSources.add(imgPath);
              images.push({
                src: imgPath,
                title: href.replace(/\.[^/.]+$/, '').replace(/_/g, ' '),
                date: ''
              });
            }
          });

          if (images.length === 0) {
            displayNoImagesMessage();
          } else {
            renderGallery(images);
          }
        })
        .catch(error => {
          console.error('Error with directory listing:', error);
          displayNoImagesMessage();
        });
    }

    function displayNoImagesMessage() {
      console.log('No images found, displaying message');

      container.innerHTML = `
        <div style="grid-column: 1/-1; text-align: center; padding: 2rem;">
          <p>No photos found in the images directory.</p>
          <p>Add your photos to the "docs/images" folder.</p>
        </div>
      `;
    }

    function formatDate(dateStr) {
      // Convert YYYYMMDD to a readable format
      const year = dateStr.substring(0, 4);
      const month = dateStr.substring(4, 6);
      const day = dateStr.substring(6, 8);
      return `${year}-${month}-${day}`;
    }

    function renderGallery(imagesToRender) {
      if (imagesToRender.length === 0) return;

      // Sort images by filename (which should have date info)
      imagesToRender.sort((a, b) => a.src.localeCompare(b.src));

      // Clear container first (to prevent duplicates)
      container.innerHTML = '';

      // Add header with count
      container.innerHTML = `<div class="gallery-header">Found ${imagesToRender.length} photos</div>`;

      // Render each image
      imagesToRender.forEach((image, index) => {
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
        const image = imagesToRender[index];
        const imgElement = lightbox.querySelector('.lightbox-img');
        const title = lightbox.querySelector('.lightbox-title');
        const description = lightbox.querySelector('.lightbox-description');
        const counter = lightbox.querySelector('.lightbox-counter');

        imgElement.src = image.src;
        imgElement.alt = image.title;
        title.textContent = image.title;
        description.textContent = image.date || '';
        counter.textContent = `${index + 1} / ${imagesToRender.length}`;

        lightbox.classList.add('active');
        document.body.style.overflow = 'hidden';

        // Configure navigation buttons
        const prevBtn = lightbox.querySelector('.lightbox-prev');
        const nextBtn = lightbox.querySelector('.lightbox-next');

        prevBtn.onclick = () => {
          const newIndex = (index - 1 + imagesToRender.length) % imagesToRender.length;
          openLightbox(newIndex);
        };

        nextBtn.onclick = () => {
          const newIndex = (index + 1) % imagesToRender.length;
          openLightbox(newIndex);
        };
      }
    }
  }
})();

const withMarkdoc = require('@markdoc/next.js');

module.exports = withMarkdoc({
  markdoc: {
    mode: 'static',
    contentDirs: ['docs'],  // Points to root-level docs
    pageExtensions: ['md', 'mdoc']
  }
});
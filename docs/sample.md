# docs/about.md
---
title: About MyFlow
description: A DevOps workflow and documentation platform
---

{% callout type="note" %}
MyFlow extends Markdoc to create a complete DevOps workflow system.
{% /callout %}

## Overview

MyFlow combines documentation, scripting, and containerization to streamline DevOps workflows. Built on Markdoc, it provides:

- **Documentation**: Integrated work logging and technical documentation
- **Scripting**: Centralized management of DevOps utilities
- **Containers**: Ready-to-use development environments
- **Cross-platform**: Support for Ubuntu and MacOS

## Implementation Timeline

### Phase 1: Core Platform
- Markdoc integration
- Documentation structure
- Basic scripting framework
- Container templates

### Phase 2: DevOps Tools
- AWS/GCP scripts
- Database containers
- Testing environments
- Ansible playbooks

### Phase 3: Documentation
```bash
# Example directory structure
myflow/
├── docs/
├── scripts/
└── configs/
```

## Security Considerations

{% callout type="warning" %}
Never commit sensitive information to the repository.
{% /callout %}

- Credentials managed externally
- Container isolation
- Secure defaults

## Quick Start

1. Clone repository
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start development server:
   ```bash
   npm run dev
   ```

{% callout type="check" %}
Access documentation at `http://localhost:3000`
{% /callout %}

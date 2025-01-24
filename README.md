# README.md
# DailyFlow - Personal Workflow Manager

A personal workflow management system for engineers, focusing on daily task logging, script management, and documentation.

## Features
- Daily work logging in markdown format
- Centralized script management for AWS, GCP, and system tasks
- Documentation site using Markdoc
- Configuration management for development tools
- Cross-platform support (Ubuntu/MacOS)

## Prerequisites
- Node.js >= 18
- Python >= 3.10
- Go >= 1.20
- AWS CLI v2.22+
- Terraform >= 1.5
- Vim/VSCode
- Kitty terminal

## Quick Start
1. Clone repository and setup markdoc
```
git clone git@github.com:Slow-Town-LLC-Internal/myflow.git
cd myflow
npm install
npm install @markdoc/markdoc @markdoc/next.js next react react-dom
npm run dev
```
3. Access docs at http://localhost:3000

## Security
- No sensitive information stored in repository
- AWS credentials managed via AWS SSO
- Environment variables stored in ~/.env (not tracked)


# future directory structure


```

myflow/
├── README.md
├── ROADMAP.md
├── docs/
│   ├── index.md
│   ├── worklogs/
│   │   └── 2024/
│   │       └── january.md
│   └── guides/
│       ├── setup.md
│       └── scripts.md
├── scripts/
│   ├── aws/
│   │   ├── eks-utils.sh
│   │   └── rds-utils.sh
│   ├── gcp/
│   ├── system/
│   │   ├── backup.sh
│   │   └── setup.sh
│   └── terraform/
│       └── modules/
├── configs/
│   ├── vim/
│   │   └── .vimrc
│   ├── kitty/
│   │   └── kitty.conf
│   └── git/
│       └── .gitignore
├── ansible/
│   ├── inventory/
│   └── playbooks/
├── package.json
└── .gitignore

```

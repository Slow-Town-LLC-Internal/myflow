#!/bin/bash

# Install Homebrew if not installed
if ! command -v brew &> /dev/null; then
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi

# Get and run shared setup script
curl -O https://raw.githubusercontent.com/Slow-Town-LLC-Internal/myflow/main/scripts/install.sh
chmod +x install.sh
./install.sh

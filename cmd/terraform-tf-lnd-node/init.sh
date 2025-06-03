export TF_VAR_linode_token=$LND_API_KEY
export TF_VAR_tailscale_authkey=$TS_API_KEY
export TF_VAR_ssh_public_key="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAINAKxL0lvEFEKfB1G9Ktn1gZNU5TJYuVBDhqQzHg8Gk3 andyxiang@pezware.com"
export TF_VAR_ssh_public_key=$(curl --silent https://github.com/arbeitandy.keys |head -1)

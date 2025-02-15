#!/usr/bin/env python3
import subprocess
import sys
import os
import yaml
from pathlib import Path

class EKSClusterSwitcher:
    def __init__(self, config_path):
        self.home = str(Path.home())
        self.aws_config_path = f"{self.home}/.aws/config"
        self.kubeconfig_path = f"{self.home}/.kube/config"
        self.config_path = f"{self.home}/.config/eks-config.yaml"
        self.config = self._load_config(config_path)

    def _load_config(self, config_path):
        """Load cluster configuration from YAML"""
        if not os.path.exists(config_path):
            raise FileNotFoundError(f"Configuration file not found: {config_path}")

        with open(config_path, 'r') as f:
            try:
                return yaml.safe_load(f)
            except yaml.YAMLError as e:
                raise Exception(f"Error parsing configuration file: {e}")

    def _validate_config(self, env):
        """Validate configuration for given environment"""
        if 'environments' not in self.config:
            raise ValueError("Invalid config: 'environments' section missing")

        if env not in self.config['environments']:
            raise ValueError(f"Environment '{env}' not found in config")

        env_config = self.config['environments'][env]
        required_keys = ['aws_profile', 'eks_cluster']

        for key in required_keys:
            if key not in env_config:
                raise ValueError(f"Missing required config '{key}' for environment '{env}'")

    def switch_cluster(self, env):
        """Switch to the specified environment cluster"""
        if env not in ["prd", "stg"]:
            raise ValueError("Environment must be 'prd' or 'stg'")

        self._validate_config(env)
        env_config = self.config['environments'][env]

        profile = env_config['aws_profile']
        cluster_name = env_config['eks_cluster']
        region = env_config['eks_region']

        try:
            # Verify AWS SSO session
            try:
                subprocess.run(
                    ["aws", "sts", "get-caller-identity", "--profile", profile],
                    check=True, capture_output=True
                )
            except subprocess.CalledProcessError:
                print(f"AWS SSO session expired. Initiating login...")
                subprocess.run(["aws", "sso", "login", "--profile", profile], check=True)

            # Update kubeconfig
            subprocess.run([
                "aws", "eks", "update-kubeconfig",
                "--name", cluster_name,
                "--profile", profile,
                "--region", region
            ], check=True)

            print(f"Successfully switched to {env} cluster: {cluster_name}")
            os.environ["AWS_PROFILE"] = profile
            print(f"AWS_PROFILE environment variable set to: {profile}")

            # Show current context
            subprocess.run(["kubectl", "config", "current-context"], check=True)

        except subprocess.CalledProcessError as e:
            raise Exception(f"Failed to switch cluster: {e}")

def main():
    if len(sys.argv) != 2 or sys.argv[1] not in ["prd", "stg"]:
        print("Usage: switch-cluster.py <prd|stg>")
        sys.exit(1)

    try:
        config_path = os.environ.get("EKS_CONFIG", f"{self.home}/.config/eks-config.yaml")
        switcher = EKSClusterSwitcher(config_path)
        switcher.switch_cluster(sys.argv[1])
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()

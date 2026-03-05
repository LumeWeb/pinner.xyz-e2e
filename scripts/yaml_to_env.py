#!/usr/bin/env python3
"""
Convert YAML configuration files to environment variables.
Converts nested YAML paths like 'core.db.type' to 'CORE__DB__TYPE=value'.
"""

import sys
import yaml


def flatten_dict(d, parent_key='', sep='__'):
    """Flatten nested dictionary into single-level dict with joined keys."""
    items = []
    for k, v in d.items():
        # Split key by dots to handle nested keys like "core.storage.sia"
        key_parts = k.split('.')
        # Convert each part to uppercase
        key_parts = [part.upper() for part in key_parts]
        # Join with double underscores
        clean_key = sep.join(key_parts)
        
        new_key = f"{parent_key}{sep}{clean_key}" if parent_key else clean_key
        if isinstance(v, dict):
            items.extend(flatten_dict(v, new_key, sep=sep).items())
        else:
            items.append((new_key, v))
    return dict(items)


def yaml_to_env(yaml_file, output_file=None):
    """Read YAML file and convert to env vars."""
    with open(yaml_file, 'r') as f:
        config = yaml.safe_load(f)
    
    if not config:
        return
    
    # Flatten nested structure
    flat_config = flatten_dict(config)
    
    # Convert to env var format
    env_vars = []
    for key, value in flat_config.items():
        # Portal expects PORTAL__ prefix for all env vars
        # Convert to uppercase and format
        env_key = key.upper()
        # Convert value to string and quote it
        env_value = str(value)
        # Always quote values to ensure proper shell handling
        env_value = f'"{env_value}"'
        env_vars.append(f"export PORTAL__{env_key}={env_value}")
    
    # Output
    if output_file:
        with open(output_file, 'a') as f:
            f.write('\n'.join(env_vars) + '\n')
    else:
        print('\n'.join(env_vars))


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python3 yaml_to_env.py <yaml_file> [output_file]")
        sys.exit(1)
    
    yaml_file = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) > 2 else None
    
    yaml_to_env(yaml_file, output_file)

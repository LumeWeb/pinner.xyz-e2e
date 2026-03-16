#!/usr/bin/env python3
"""
Convert YAML configuration files to environment variables.
Converts nested YAML paths like 'core.db.type' to 'CORE__DB__TYPE=value'.
"""

import sys
import yaml


def flatten_dict(d, parent_key='', sep='__', top_level=True):
    """Flatten nested dictionary into single-level dict with joined keys."""
    items = []
    for k, v in d.items():
        # Split key by dots to handle flattened YAML keys like "core.db"
        # Only split at the top level, not recursively in nested structures
        if top_level:
            key_parts = k.split('.')
        else:
            key_parts = [k]
        # Convert each part to uppercase
        key_parts = [part.upper() for part in key_parts]
        # Join with double underscores
        clean_key = sep.join(key_parts)
        
        new_key = f"{parent_key}{sep}{clean_key}" if parent_key else clean_key
        if isinstance(v, dict):
            items.extend(flatten_dict(v, new_key, sep=sep, top_level=False).items())
        else:
            items.append((new_key, v))
    return dict(items)


def format_value(value):
    """Format value for environment variable output."""
    if isinstance(value, list):
        # Convert list to JSON array format for proper env parsing
        return json.dumps(value)
    return str(value)


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
        
        # Convert value to appropriate format
        if isinstance(value, list):
            # Convert list to comma-separated string for array parsing
            env_value = ','.join(str(item) for item in value)
        elif isinstance(value, str):
            # Quote string values
            env_value = f'"{value}"'
        elif isinstance(value, bool):
            # Convert boolean to string
            env_value = f'"{str(value).lower()}"'
        elif value is None:
            # Convert None to empty string
            env_value = '""'
        else:
            # Convert other values to string
            env_value = f'"{str(value)}"'
        
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

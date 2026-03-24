#!/usr/bin/env python3
# Sample script to test DNS API endpoints

import requests
import json

BASE_URL = "http://127.0.0.1:8000"

def test_health():
    """Test health endpoint."""
    resp = requests.get(f"{BASE_URL}/health")
    print(f"Health: {resp.json()}")

def test_list_zones():
    """Test list zones endpoint."""
    resp = requests.get(f"{BASE_URL}/api/zones")
    print(f"Zones: {resp.json()}")

def test_add_zone():
    """Test add zone endpoint."""
    zone = {
        "domain": "test-website-nodns.example.com",
        "records": [
            {
                "type": "TXT",
                "value": "dnslink=/ipfs/QmX7Z3Xn5F1mX1Y1Z1",
                "ttl": 300
            },
            {
                "type": "TXT",
                "value": "lume-verification-token=abc123xyz",
                "ttl": 300
            },
            {
                "type": "CNAME",
                "value": "account.localhost",
                "ttl": 300
            }
        ]
    }
    resp = requests.post(
        f"{BASE_URL}/api/zones",
        json=zone
    )
    print(f"Add zone: {resp.json()}")

if __name__ == "__main__":
    print("Testing DNS Development Server API")
    print("=" * 50)
    
    try:
        test_health()
        test_add_zone()
        test_list_zones()
        print("✓ All tests passed")
    except Exception as e:
        print(f"✗ Test failed: {e}")

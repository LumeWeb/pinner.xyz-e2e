#!/usr/bin/env python3
"""
Dynamic DNS development server for E2E testing.
Supports whitelist pass-through to upstream DNS and local mappings for test domains.
Uses dnslib for full programmatic control without configuration files.
"""

import os
import socket
import sys
import urllib.parse

try:
    from dnslib import DNSRecord, DNSHeader, RR, A, MX, CNAME, TXT, NS
    from dnslib import QTYPE
except ImportError:
    print("Error: dnslib not installed. Install with: pip install dnslib", file=sys.stderr)
    sys.exit(1)

class DynamicDNSServer:
    """Dynamic DNS server with whitelist pass-through and local mappings."""
    
    def __init__(self, host='127.0.0.1', port=5353, upstream_dns='8.8.8.8'):
        self.host = host
        self.port = port
        self.upstream_dns = upstream_dns
        self.socket = None
        self.running = False
        
        # Build local mappings and whitelist from environment
        self.local_map, self.whitelist = self._build_dns_config()
        
        # DNS response handlers for different query types
        self.response_handlers = {
            QTYPE.A: self._handle_a_record,
            QTYPE.MX: self._handle_mx_record,
            QTYPE.CNAME: self._handle_cname_record,
            QTYPE.TXT: self._handle_txt_record,
            QTYPE.NS: self._handle_ns_record,
        }
    
    def _extract_hostname_from_url(self, url):
        """Extract hostname from URL string."""
        try:
            parsed = urllib.parse.urlparse(url)
            return parsed.hostname
        except Exception:
            return None
    
    def _build_dns_config(self):
        """Build local DNS mappings and whitelist from environment variables."""
        local_map = {}
        whitelist = []
        
        # --- Local mappings (explicit local services) ---
        # Format: {'hostname': 'ip_address'}
        local_map = {
            # Portal vhost from PORTAL_HOST env var
            os.getenv('PORTAL_HOST', 'account.pinner.xyz'): '127.0.0.1',
            # MySQL host from PORTAL__CORE__DB__HOST env var
            os.getenv('PORTAL__CORE__DB__HOST', 'mysql'): '127.0.0.1',
            # Other common test service hostnames
            'maildev': '127.0.0.1',
            'gofakes3': '127.0.0.1',
        }
        
        # --- Whitelist configuration ---
        # Format: [(env_var_name, extraction_type)]
        # extraction_type: 'url' = extract hostname from URL, 'domain' = use value directly
        whitelist_config = [
            ('RENTERD_URL', 'url'),
            ('PORTAL__CORE__STORAGE__SIA__URL', 'url'),
        ]
        
        # Build whitelist from config
        for env_var, extract_type in whitelist_config:
            value = os.getenv(env_var, '')
            if not value:
                continue
            
            if extract_type == 'url':
                domain = self._extract_hostname_from_url(value)
            elif extract_type == 'domain':
                domain = value
            else:
                continue
            
            if domain:
                whitelist.append(domain)
        
        return local_map, whitelist
    
    def should_pass_through(self, domain):
        """Check if domain should pass through to upstream DNS based on whitelist."""
        domain_lower = domain.lower()
        
        # Check exact matches first
        for pattern in self.whitelist:
            pattern_lower = pattern.lower()
            
            if pattern.startswith('*'):
                # Wildcard pattern: check if domain ends with the suffix
                suffix = pattern[1:]  # Remove '*'
                if domain_lower.endswith(suffix) or domain_lower == suffix[1:]:
                    return True
            else:
                # Exact match
                if domain_lower == pattern_lower:
                    return True
        
        return False
    
    def _send_dns_response(self, request, addr, rtype, rdata):
        """Send a DNS response with the given record type and data."""
        reply = request.reply()
        reply.add_answer(RR(request.q.qname, rtype=rtype, ttl=300, rdata=rdata))
        self.socket.sendto(reply.pack(), addr)
    
    def _handle_a_record(self, request, qname, addr):
        """Handle A record queries."""
        ip = self.local_map.get(qname, '127.0.0.1')
        self._send_dns_response(request, addr, QTYPE.A, A(ip))
    
    def _handle_mx_record(self, request, qname, addr):
        """Handle MX record queries."""
        self._send_dns_response(request, addr, QTYPE.MX, MX(label="maildev", preference=10))
    
    def _handle_cname_record(self, request, qname, addr):
        """Handle CNAME record queries."""
        self._send_dns_response(request, addr, QTYPE.CNAME, CNAME(label=qname))
    
    def _handle_txt_record(self, request, qname, addr):
        """Handle TXT record queries."""
        self._send_dns_response(request, addr, QTYPE.TXT, TXT(b"v=spf1 mx -all"))
    
    def _handle_ns_record(self, request, qname, addr):
        """Handle NS record queries."""
        self._send_dns_response(request, addr, QTYPE.NS, NS(label=qname))
    
    def forward_to_upstream(self, request_data):
        """Forward query to upstream DNS server."""
        upstream_socket = None
        try:
            # Create socket for upstream DNS
            upstream_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            upstream_socket.settimeout(5)
            
            # Forward request to upstream DNS
            upstream_socket.sendto(request_data, (self.upstream_dns, 53))
            response_data, _ = upstream_socket.recvfrom(1024)
            
            return response_data
        except Exception as e:
            print(f"Error forwarding to upstream DNS: {e}", file=sys.stderr)
            return None
        finally:
            if upstream_socket:
                upstream_socket.close()
    
    def handle_query(self, data, addr):
        """Handle incoming DNS query."""
        try:
            # Parse DNS query
            request = DNSRecord.parse(data)
            
            qname = str(request.q.qname).rstrip('.')
            qtype = QTYPE[request.q.qtype]
            qtype_int = request.q.qtype
            
            # Check if domain is in local map
            if qname in self.local_map:
                ip = self.local_map[qname]
                print(f"DNS query: {qname} ({qtype}) from {addr} -> LOCAL {ip}")
                
                # Use response handler if available
                handler = self.response_handlers.get(qtype_int)
                if handler:
                    handler(request, qname, addr)
                    return
                
                # For unsupported types, try to pass through
                if self.should_pass_through(qname):
                    print(f"DNS query: {qname} ({qtype}) from {addr} -> UPSTREAM {self.upstream_dns}")
                    response_data = self.forward_to_upstream(data)
                    if response_data:
                        self.socket.sendto(response_data, addr)
                        return
                
                # Return NXDOMAIN for unsupported types
                reply = request.reply()
                reply.header.rcode = 3  # NXDOMAIN
                self.socket.sendto(reply.pack(), addr)
                return
            
            # Check whitelist for pass-through FIRST (before universal handler)
            if self.should_pass_through(qname):
                print(f"DNS query: {qname} ({qtype}) from {addr} -> UPSTREAM (whitelist)")
                
                # Forward to upstream DNS
                response_data = self.forward_to_upstream(data)
                if response_data:
                    self.socket.sendto(response_data, addr)
                else:
                    # If upstream fails, return NXDOMAIN
                    reply = request.reply()
                    reply.header.rcode = 3  # NXDOMAIN
                    self.socket.sendto(reply.pack(), addr)
                return
            
            # For all other domains, handle MX and common email-related queries
            # This ensures email verification works for any test domain
            if qtype_int in (QTYPE.MX, QTYPE.A, QTYPE.TXT, QTYPE.CNAME, QTYPE.NS):
                print(f"DNS query: {qname} ({qtype}) from {addr} -> UNIVERSAL")
                handler = self.response_handlers.get(qtype_int)
                if handler:
                    handler(request, qname, addr)
                    return
            
            # Check whitelist for pass-through
            if self.should_pass_through(qname):
                print(f"DNS query: {qname} ({qtype}) from {addr} -> UPSTREAM {self.upstream_dns}")
                
                # Forward to upstream DNS
                response_data = self.forward_to_upstream(data)
                if response_data:
                    self.socket.sendto(response_data, addr)
                else:
                    # If upstream fails, return NXDOMAIN
                    reply = request.reply()
                    reply.header.rcode = 3  # NXDOMAIN
                    self.socket.sendto(reply.pack(), addr)
                return
            
            # Default: pass through to upstream for unknown domains
            print(f"DNS query: {qname} ({qtype}) from {addr} -> UPSTREAM (default)")
            
            response_data = self.forward_to_upstream(data)
            if response_data:
                self.socket.sendto(response_data, addr)
            else:
                # If upstream fails, return NXDOMAIN
                reply = request.reply()
                reply.header.rcode = 3
                self.socket.sendto(reply.pack(), addr)
            
        except Exception as e:
            print(f"Error handling query: {e}", file=sys.stderr)
    
    def start(self):
        """Start the DNS server."""
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        
        try:
            self.socket.bind((self.host, self.port))
        except PermissionError:
            print(f"Error: Cannot bind to port {self.port}. Try with sudo or use a different port.", file=sys.stderr)
            sys.exit(1)
        
        self.running = True
        print(f"Dynamic DNS server listening on {self.host}:{self.port}")
        print(f"Upstream DNS: {self.upstream_dns}")
        print("\nLocal mappings (explicit local services):")
        for hostname, ip in self.local_map.items():
            print(f"  {hostname} -> {ip}")
        print("\nWhitelist (pass-through to upstream DNS):")
        if self.whitelist:
            for pattern in self.whitelist:
                print(f"  {pattern}")
        else:
            print("  (empty - all domains pass through to upstream)")
        print("\nPress Ctrl+C to stop")
        
        try:
            while self.running:
                data, addr = self.socket.recvfrom(1024)
                self.handle_query(data, addr)
        except KeyboardInterrupt:
            print("\nShutting down...")
        finally:
            self.stop()
    
    def stop(self):
        """Stop the DNS server."""
        self.running = False
        if self.socket:
            self.socket.close()

def main():
    """Main entry point."""
    import argparse
    
    parser = argparse.ArgumentParser(description='Dynamic DNS dev server for E2E testing')
    parser.add_argument('--host', default='127.0.0.1', help='Host to bind to (default: 127.0.0.1)')
    parser.add_argument('--port', type=int, default=5353, help='Port to bind to (default: 5353)')
    parser.add_argument('--upstream', default='8.8.8.8', help='Upstream DNS server (default: 8.8.8.8)')
    
    args = parser.parse_args()
    
    server = DynamicDNSServer(host=args.host, port=args.port, upstream_dns=args.upstream)
    server.start()

if __name__ == '__main__':
    main()

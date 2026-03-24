#!/usr/bin/env python3
"""
Dynamic DNS development server for E2E testing with FastAPI REST API.
Supports in-memory DNS zone management and whitelist pass-through to upstream DNS.
"""

import os
import socket
import sys
import urllib.parse
from typing import Optional, Dict, List
from pydantic import BaseModel
from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
import uvicorn
from threading import Thread
import time

try:
    from dnslib import DNSRecord, DNSHeader, RR, A, MX, CNAME, TXT, NS, SOA
    from dnslib import QTYPE
except ImportError:
    print("Error: dnslib not installed. Install with: pip install dnslib", file=sys.stderr)
    sys.exit(1)

# ============================================================================
# FastAPI Models
# ============================================================================

class DNSRecordModel(BaseModel):
    """DNS record model for API requests."""
    name: str = "@"  # @ for zone apex, or subdomain like _dnslink
    type: str  # 'A', 'TXT', 'CNAME', 'MX', 'NS', 'SOA'
    value: str
    ttl: int = 300
    priority: Optional[int] = None  # For MX records

class DNSZoneModel(BaseModel):
    """DNS zone model for API requests."""
    domain: str
    records: List[DNSRecordModel]

class MessageResponse(BaseModel):
    """Standard message response."""
    message: str

# ============================================================================
# DNS Server with FastAPI
# ============================================================================

class DynamicDNSServer:
    """Dynamic DNS server with in-memory zone management and REST API."""
    
    def __init__(self, host='127.0.0.1', port=5353, api_port=8000, upstream_dns='8.8.8.8'):
        self.host = host
        self.port = port
        self.api_port = api_port
        self.upstream_dns = upstream_dns
        
        # UDP server
        self.socket = None
        self.running = False
        
        # FastAPI app
        self.app = FastAPI(title="DNS Development Server", version="2.0.0")
        
        # In-memory storage for DNS zones
        self.zones: Dict[str, Dict[str, List[Dict]]] = {}
        # zones = {domain: {record_type: [{'value': '...', 'ttl': 300}, ...]}}
        
        # Build local mappings and whitelist from environment
        self.local_map, self.whitelist = self._build_dns_config()
        
        # Build reverse mapping from QTYPE integer to name
        self.qtype_int_to_name = QTYPE.forward
        
        # DNS response handlers for different query types
        self.response_handlers = {
            QTYPE.A: self._handle_a_record,
            QTYPE.MX: self._handle_mx_record,
            QTYPE.CNAME: self._handle_cname_record,
            QTYPE.TXT: self._handle_txt_record,
            QTYPE.NS: self._handle_ns_record,
            QTYPE.SOA: self._handle_soa_record,
        }
        
        # Setup FastAPI routes
        self._setup_routes()
    
    def _setup_routes(self):
        """Setup FastAPI routes for DNS zone management."""
        
        @self.app.get("/")
        async def root():
            """Root endpoint."""
            return {
                "message": "DNS Development Server API",
                "version": "2.0.0",
                "endpoints": {
                    "zones": "/api/zones",
                    "add_zone": "/api/zones (POST)",
                    "delete_zone": "/api/zones/{domain} (DELETE)",
                    "health": "/health",
                }
            }
        
        @self.app.get("/health")
        async def health():
            """Health check endpoint."""
            return {"status": "ok", "zones_count": len(self.zones)}
        
        @self.app.get("/api/zones")
        async def list_zones():
            """List all in-memory DNS zones."""
            zones_list = []
            for domain, records in self.zones.items():
                zone_data = {"domain": domain, "records": []}
                # Structure: domain -> name -> type -> records
                for record_name, type_dict in records.items():
                    for record_type, record_list in type_dict.items():
                        for rec in record_list:
                            zone_data["records"].append({
                                "name": record_name if record_name else "@",
                                "type": record_type,
                                "value": rec['value'],
                                "ttl": rec.get('ttl', 300),
                                "priority": rec.get('priority')
                            })
                zones_list.append(zone_data)
            return {"zones": zones_list, "count": len(zones_list)}
        
        @self.app.post("/api/zones")
        async def add_zone(zone: DNSZoneModel):
            """Add a DNS zone with records to memory."""
            try:
                if not zone.domain:
                    raise HTTPException(status_code=400, detail="Domain is required")
                
                # Initialize zone records if not exists
                self.zones[zone.domain] = {}
                
                # Add each record
                for i, rec in enumerate(zone.records):
                    record_type = rec.type.upper()
                    record_name = rec.name if rec.name and rec.name != "@" else ""  # Empty string for zone apex
                    
                    # Initialize name dict if not exists
                    if record_name not in self.zones[zone.domain]:
                        self.zones[zone.domain][record_name] = {}
                    
                    # Initialize type dict if not exists
                    if record_type not in self.zones[zone.domain][record_name]:
                        self.zones[zone.domain][record_name][record_type] = []
                    
                    self.zones[zone.domain][record_name][record_type].append({
                        'value': rec.value,
                        'ttl': rec.ttl,
                        'priority': rec.priority
                    })
                
                print(f"Added zone: {zone.domain} with {len(zone.records)} records")
                return {"message": f"Zone {zone.domain} added successfully", "domain": zone.domain}
            
            except Exception as e:
                raise HTTPException(status_code=500, detail=str(e))
        
        @self.app.delete("/api/zones/{domain:path}")
        async def delete_zone(domain: str):
            """Delete a DNS zone from memory."""
            if domain not in self.zones:
                raise HTTPException(status_code=404, detail=f"Zone {domain} not found")
            
            del self.zones[domain]
            print(f"Deleted zone: {domain}")
            return {"message": f"Zone {domain} deleted successfully"}
        
        @self.app.get("/api/zones/{domain:path}")
        async def get_zone(domain: str):
            """Get a specific DNS zone."""
            if domain not in self.zones:
                raise HTTPException(status_code=404, detail=f"Zone {domain} not found")
            
            records = []
            for record_type, record_list in self.zones[domain].items():
                for rec in record_list:
                    records.append({
                        "type": record_type,
                        "value": rec['value'],
                        "ttl": rec.get('ttl', 300),
                        "priority": rec.get('priority')
                    })
            
            return {"domain": domain, "records": records, "count": len(records)}
    
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
        
        # Local mappings (explicit local services)
        local_map = {
            os.getenv('PORTAL_HOST', 'account.pinner.xyz'): '127.0.0.1',
            os.getenv('PORTAL__CORE__DB__HOST', 'mysql'): '127.0.0.1',
            'maildev': '127.0.0.1',
            'gofakes3': '127.0.0.1',
        }
        
        # Whitelist configuration
        whitelist_config = [
            ('RENTERD_URL', 'url'),
            ('PORTAL__CORE__STORAGE__SIA__URL', 'url'),
        ]
        
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
        
        for pattern in self.whitelist:
            pattern_lower = pattern.lower()
            
            if pattern.startswith('*'):
                suffix = pattern[1:]
                if domain_lower.endswith(suffix) or domain_lower == suffix[1:]:
                    return True
            else:
                if domain_lower == pattern_lower:
                    return True
        
        return False
    
    def _get_zone_record(self, qname: str, qtype_int: int):
        """Get DNS record from in-memory zones.
        
        Handles subdomain records by finding the appropriate zone and record name.
        For example:
        - _dnslink.example.com TXT looks for name=_dnslink in zone example.com
        - example.com TXT looks for name='' (zone apex) in zone example.com
        """
        qtype_str = None
        
        # Convert QTYPE integer to string using reverse mapping
        qtype_str = self.qtype_int_to_name.get(qtype_int)
        
        if not qtype_str:
            return None
        
        # Try to find matching zone
        # Check for exact zone match first, then check if qname ends with a known zone
        zone_domain = None
        record_name = None
        
        # Exact match (zone apex query)
        if qname in self.zones:
            zone_domain = qname
            record_name = ""
        else:
            # Find the zone by checking if qname ends with any known zone
            for zdom in self.zones:
                if qname == zdom or qname.endswith("." + zdom):
                    zone_domain = zdom
                    # Extract record name: remove zone domain from qname
                    if qname == zdom:
                        record_name = ""
                    else:
                        record_name = qname[:-len(zdom) - 1]  # Remove ".zone"
                    break
        
        if not zone_domain:
            return None
        
        # Check if name and record type exist in the zone
        zone = self.zones[zone_domain]
        if record_name in zone and qtype_str in zone[record_name]:
            records = zone[record_name][qtype_str]
            if records:
                return records  # Return all records
        return None
    
    def _handle_zone_record(self, request, qname, addr, qtype_int):
        """Handle DNS records from in-memory zones."""
        records = self._get_zone_record(qname, qtype_int)
        if not records:
            return False
        
        # Build response based on record type using qtype_int_to_name mapping
        qtype_name = self.qtype_int_to_name.get(qtype_int)
        if not qtype_name:
            return False
        
        # Create DNS reply once, but don't send it until all records are added
        reply = request.reply()
        
        # For each record, add it to the response
        for record in records:
            value = record['value']
            ttl = record.get('ttl', 300)
            
            if qtype_name == 'A':
                reply.add_answer(RR(request.q.qname, rtype=QTYPE.A, ttl=ttl, rdata=A(value)))
            elif qtype_name == 'TXT':
                # Handle TXT records properly - need bytes
                txt_value = value.encode() if isinstance(value, str) else value
                reply.add_answer(RR(request.q.qname, rtype=QTYPE.TXT, ttl=ttl, rdata=TXT(txt_value)))
            elif qtype_name == 'CNAME':
                reply.add_answer(RR(request.q.qname, rtype=QTYPE.CNAME, ttl=ttl, rdata=CNAME(label=value)))
            elif qtype_name == 'MX':
                priority = record.get('priority', 10)
                reply.add_answer(RR(request.q.qname, rtype=QTYPE.MX, ttl=ttl, rdata=MX(label=value, preference=priority)))
            elif qtype_name == 'NS':
                reply.add_answer(RR(request.q.qname, rtype=QTYPE.NS, ttl=ttl, rdata=NS(label=value)))
            elif qtype_name == 'SOA':
                reply.add_answer(RR(request.q.qname, rtype=QTYPE.SOA, ttl=ttl, rdata=SOA(
                    mname=value,
                    rname="hostmaster." + qname,
                    times=(2024010101, 3600, 1800, 604800, 86400)
                )))
        
        # Send the response with all records
        self.socket.sendto(reply.pack(), addr)
        return True
    
    def _send_dns_response(self, request, addr, rtype, rdata, ttl=300):
        """Send a DNS response with the given record type and data."""
        reply = request.reply()
        reply.add_answer(RR(request.q.qname, rtype=rtype, ttl=ttl, rdata=rdata))
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
        # First check if there's a zone record configured for this domain
        if self._handle_zone_record(request, qname, addr, QTYPE.TXT):
            return
        # Return NXDOMAIN for TXT queries when no zone record exists
        reply = request.reply()
        reply.header.rcode = 3
        self.socket.sendto(reply.pack(), addr)
    
    def _handle_ns_record(self, request, qname, addr):
        """Handle NS record queries."""
        # First check if there's a zone record configured for this domain
        if self._handle_zone_record(request, qname, addr, QTYPE.NS):
            return
        # Return NXDOMAIN for NS queries when no zone record exists
        self._send_dns_response(request, addr, QTYPE.NS, NS(label=qname))
    
    def _handle_soa_record(self, request, qname, addr):
        """Handle SOA record queries."""
        self._send_dns_response(request, addr, QTYPE.SOA, SOA(
            mname=f"ns.{qname}",
            rname=f"hostmaster.{qname}",
            times=(2024010101, 3600, 1800, 604800, 86400)
        ))
    
    def forward_to_upstream(self, request_data):
        """Forward query to upstream DNS server."""
        upstream_socket = None
        try:
            upstream_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            upstream_socket.settimeout(5)
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
            request = DNSRecord.parse(data)
            qname = str(request.q.qname).rstrip('.')
            qtype_int = request.q.qtype
            qtype = QTYPE[qtype_int]
            
            # First check if domain is in memory zones
            if qname in self.zones:
                print(f"DNS query: {qname} ({qtype}) from {addr} -> IN-MEMORY ZONE")
                if self._handle_zone_record(request, qname, addr, qtype_int):
                    return
                # If record type not found in zone, check if pass-through enabled
                if self.should_pass_through(qname):
                    print(f"DNS query: {qname} ({qtype}) from {addr} -> UPSTREAM")
                    response_data = self.forward_to_upstream(data)
                    if response_data:
                        self.socket.sendto(response_data, addr)
                        return
                # Return NXDOMAIN for unsupported record types in zone
                reply = request.reply()
                reply.header.rcode = 3
                self.socket.sendto(reply.pack(), addr)
                return
            
            # Check if domain is in local map
            if qname in self.local_map:
                ip = self.local_map[qname]
                print(f"DNS query: {qname} ({qtype}) from {addr} -> LOCAL {ip}")
                handler = self.response_handlers.get(qtype_int)
                if handler:
                    handler(request, qname, addr)
                    return
                if self.should_pass_through(qname):
                    response_data = self.forward_to_upstream(data)
                    if response_data:
                        self.socket.sendto(response_data, addr)
                        return
                reply = request.reply()
                reply.header.rcode = 3
                self.socket.sendto(reply.pack(), addr)
                return
            
            # Check whitelist for pass-through
            if self.should_pass_through(qname):
                print(f"DNS query: {qname} ({qtype}) from {addr} -> UPSTREAM (whitelist)")
                response_data = self.forward_to_upstream(data)
                if response_data:
                    self.socket.sendto(response_data, addr)
                else:
                    reply = request.reply()
                    reply.header.rcode = 3
                    self.socket.sendto(reply.pack(), addr)
                return
            
            # For all other domains, handle common record types
            if qtype_int in (QTYPE.MX, QTYPE.A, QTYPE.TXT, QTYPE.CNAME, QTYPE.NS, QTYPE.SOA):
                print(f"DNS query: {qname} ({qtype}) from {addr} -> UNIVERSAL")
                handler = self.response_handlers.get(qtype_int)
                if handler:
                    handler(request, qname, addr)
                    return
            
            # Default: pass through to upstream
            print(f"DNS query: {qname} ({qtype}) from {addr} -> UPSTREAM (default)")
            response_data = self.forward_to_upstream(data)
            if response_data:
                self.socket.sendto(response_data, addr)
            else:
                reply = request.reply()
                reply.header.rcode = 3
                self.socket.sendto(reply.pack(), addr)
        
        except Exception as e:
            print(f"Error handling query: {e}", file=sys.stderr)
    
    def start_dns_server(self):
        """Start the DNS server on UDP."""
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        
        try:
            self.socket.bind((self.host, self.port))
        except PermissionError:
            print(f"Error: Cannot bind to port {self.port}. Try with sudo or use a different port.", file=sys.stderr)
            sys.exit(1)
        
        self.running = True
        print(f"DNS server listening on {self.host}:{self.port}")
        print(f"API server on http://{self.host}:{self.api_port}")
        print(f"Upstream DNS: {self.upstream_dns}")
        print("\nLocal mappings:")
        for hostname, ip in self.local_map.items():
            print(f"  {hostname} -> {ip}")
        print("\nWhitelist (pass-through):")
        if self.whitelist:
            for pattern in self.whitelist:
                print(f"  {pattern}")
        else:
            print("  (empty)")
        print("\nAPI endpoints:")
        print(f"  GET  http://{self.host}:{self.api_port}/api/zones")
        print(f"  POST http://{self.host}:{self.api_port}/api/zones")
        print(f"  GET  http://{self.host}:{self.api_port}/api/zones/{{domain}}")
        print(f"  DEL  http://{self.host}:{self.api_port}/api/zones/{{domain}}")
        print(f"  GET  http://{self.host}:{self.api_port}/health")
        print("\nPress Ctrl+C to stop")
        
        try:
            while self.running:
                data, addr = self.socket.recvfrom(1024)
                self.handle_query(data, addr)
        except KeyboardInterrupt:
            print("\nShutting down DNS server...")
        finally:
            self.socket.close()
    
    def start_api_server(self):
        """Start the FastAPI server in a separate thread."""
        uvicorn.run(self.app, host=self.host, port=self.api_port, log_level="info")
    
    def start(self):
        """Start both DNS and API servers."""
        # Start API server in a thread
        api_thread = Thread(target=self.start_api_server, daemon=True)
        api_thread.start()
        
        # Start DNS server in main thread
        self.start_dns_server()
    
    def stop(self):
        """Stop both servers."""
        self.running = False
        if self.socket:
            self.socket.close()

def main():
    """Main entry point."""
    import argparse
    
    parser = argparse.ArgumentParser(description='Dynamic DNS dev server with FastAPI')
    parser.add_argument('--host', default='127.0.0.1', help='Host to bind to (default: 127.0.0.1)')
    parser.add_argument('--port', type=int, default=5353, help='DNS port (default: 5353)')
    parser.add_argument('--api-port', type=int, default=8000, help='API port (default: 8000)')
    parser.add_argument('--upstream', default='8.8.8.8', help='Upstream DNS server (default: 8.8.8.8)')
    
    args = parser.parse_args()
    
    # Check for required dependencies
    try:
        import fastapi
        import uvicorn
    except ImportError:
        print("Error: FastAPI or uvicorn not installed.", file=sys.stderr)
        print("Install with: pip install fastapi uvicorn", file=sys.stderr)
        sys.exit(1)
    
    server = DynamicDNSServer(host=args.host, port=args.port, api_port=args.api_port, upstream_dns=args.upstream)
    server.start()

if __name__ == '__main__':
    main()

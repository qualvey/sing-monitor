#!/usr/bin/env python3
"""观察 140 秒，验证无流量用户是否变离线"""
import base64, json, os, re, socket, subprocess, time

def snapshot():
    tok = json.loads(subprocess.check_output([
        'curl', '-s', '-X', 'POST', 'http://127.0.0.1:8091/api/v1/auth/login',
        '-H', 'Content-Type: application/json', '-d', '{"password":"testpass"}'
    ]).decode()).get('token', '')
    s = socket.create_connection(('127.0.0.1', 8091), timeout=5)
    key = base64.b64encode(os.urandom(16)).decode()
    req = (f'GET /api/v1/ws/rt?token={tok} HTTP/1.1\r\n'
           f'Host: 127.0.0.1:8091\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n'
           f'Sec-WebSocket-Key: {key}\r\nSec-WebSocket-Version: 13\r\n\r\n')
    s.sendall(req.encode())
    s.recv(4096)
    time.sleep(1.2)
    data = s.recv(65536)
    s.close()
    m = re.search(rb'\{.*\}', data)
    if not m:
        return []
    j = json.loads(m.group(0))
    return [(u.get('name'), u.get('online'), u.get('uplink', 0), u.get('downlink', 0)) for u in j.get('users', [])]

print('T0:', snapshot())
time.sleep(70)
print('T+70s:', snapshot())
time.sleep(70)
print('T+140s:', snapshot())

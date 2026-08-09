#!/usr/bin/env python3
"""连接 8091 测试实例的 WebSocket，检查在线状态"""
import base64, json, os, re, socket, subprocess, time

tok = json.loads(subprocess.check_output([
    'curl', '-s', '-X', 'POST', 'http://127.0.0.1:8091/api/v1/auth/login',
    '-H', 'Content-Type: application/json', '-d', '{"password":"testpass"}'
]).decode()).get('token', '')
print('token len:', len(tok))

s = socket.create_connection(('127.0.0.1', 8091), timeout=5)
key = base64.b64encode(os.urandom(16)).decode()
req = (f'GET /api/v1/ws/rt?token={tok} HTTP/1.1\r\n'
       f'Host: 127.0.0.1:8091\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n'
       f'Sec-WebSocket-Key: {key}\r\nSec-WebSocket-Version: 13\r\n\r\n')
s.sendall(req.encode())
resp = s.recv(4096)
print('handshake status:', resp.split(b'\r\n')[0])

time.sleep(1.5)
data = s.recv(65536)
print('frame bytes:', len(data))
if data:
    m = re.search(rb'\{.*\}', data)
    if m:
        j = json.loads(m.group(0))
        users = j.get('users', [])
        online = [u for u in users if u.get('online')]
        print('total users in snapshot:', len(users))
        print('online count:', len(online))
        for u in users[:10]:
            print(' ', u.get('name'), 'online=' + str(u.get('online')), 'up=' + str(u.get('uplink')), 'down=' + str(u.get('downlink')))
s.close()

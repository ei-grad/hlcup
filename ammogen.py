with open('ammo.txt', 'wb') as f:
    for i in range(10000, 20000):
        body = (
            '{"id":%d,"email":"ipanfat@inbox.ru","first_name":"Владислав",'
            '"last_name":"Феташекий","gender":"m","birth_date":-1840924800}'
        ) % i
        r = (
            'POST /users/new HTTP/1.1\r\n'
            'Host: localhost\r\n'
            'Content-Length: %d\r\n'
            'Connection: close\r\n'
            '\r\n'
            '%s'
        ) % (len(body.encode('utf-8')), body)
        f.write(b'%d\n' % len(r.encode('utf-8')))
        f.write(r.encode('utf-8'))
        f.write(b'\r\n')

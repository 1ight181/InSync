cd /d "%~dp0"
:: создает серты для клиента и сервера, требует наличия ca.crt и валидного server.ext
:: создает для сервера
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=server.local"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 3650 -sha256 -extfile server.ext
:: создает для клиента
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/CN=client"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 3650 -sha256
timeout /t 5
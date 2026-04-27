cd /d "%~dp0"
:: Создает корневой сертификат
openssl genrsa -out ca.key 2048
openssl req -x509 -new -key ca.key -nodes -days 3650 -out ca.crt -subj "/CN=ca"
timeout /t 5
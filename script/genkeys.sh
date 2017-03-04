mkdir -p /etc/ssl/localcerts

openssl req \
    -new \
    -newkey rsa:1024 \
    -days 2000 \
    -nodes \
    -x509 \
    -subj "/C=US/ST=NY/L=NY/O=Dis/CN=artur.dev" \
    -keyout /etc/ssl/localcerts/artur.dev.pem \
    -out /etc/ssl/localcerts/artur.dev.crt

certutil -d sql:$HOME/.pki/nssdb -A -t TC -n "artur.dev" -i /etc/ssl/localcerts/artur.dev.crt



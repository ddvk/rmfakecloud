#!/bin/sh
set -e
# thanks to  https://gist.github.com/Soarez/9688998

cat <<EOF > csr.conf
[ req ]
default_bits = 2048
default_keyfile = proxy.key
encrypt_key = no
default_md = sha256
prompt = no
utf8 = yes
distinguished_name = dn
req_extensions = ext
x509_extensions = caext

[ dn ]
C = AA
ST = QQ
L = JJ
O  = the culture
CN = *.appspot.com

[ ext ]
subjectAltName=@san
basicConstraints=CA:FALSE
subjectKeyIdentifier = hash


[ caext ]
subjectAltName=@san

[ san ]
DNS.1 = *.appspot.com
DNS.2 = my.remarkable.com
# DNS.3 = any additional hosts
EOF

# host
echo "Generating proxy keys..."
openssl genrsa -out proxy.key 2048
openssl rsa -in proxy.key -pubout -out proxy.pubkey
openssl req -new -config ./csr.conf -key proxy.key -out proxy.csr 


# ca
echo "Generating ca..."
openssl genrsa -out ca.key 2048
openssl req -new -sha256 -x509 -key ca.key -out ca.crt -days 3650 -subj /CN=rmfakecloud

# Signing
openssl x509 -req  -in proxy.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out proxy.crt -days 3650 -extfile csr.conf -extensions caext
cat proxy.crt ca.crt > proxy.bundle.crt

echo "showing result"
openssl x509 -in proxy.bundle.crt -text -noout 

echo "Generation complete"


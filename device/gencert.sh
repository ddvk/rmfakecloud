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
distinguished_name = my_req_distinguished_name
req_extensions = my_extensions

[ my_req_distinguished_name ]
C = AA
ST = QQ
L = JJ
O  = the culture
CN = *.appspot.com

[ my_extensions ]
basicConstraints=CA:FALSE
subjectAltName=@my_subject_alt_names
subjectKeyIdentifier = hash

[ my_subject_alt_names ]
DNS.1 = *.appspot.com
DNS.2 = my.remarkable.com
EOF

# host
echo "Generating proxy keys..."
openssl genrsa -out proxy.key 2048
#openssl rsa -in proxy.key -pubout -out proxy.pubkey
openssl req -new -out proxy.csr -config csr.conf -key proxy.key


# ca
echo "Generating ca..."
openssl genrsa -out ca.key 2048
openssl req -new -x509 -key ca.key -out ca.crt -days 3650

# Signing
openssl x509 -req -in proxy.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out proxy.crt -days 3650
cat proxy.crt ca.crt > proxy.bundle.crt

echo "showing result"
openssl x509 -in proxy.bundle.crt -text -noout -dates


# install
echo "Installing the generated CA"
mkdir -p /usr/local/share/ca-certificates
cp ca.crt /usr/local/share/ca-certificates
update-ca-certificates


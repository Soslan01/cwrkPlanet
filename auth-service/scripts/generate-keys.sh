#!/bin/bash

mkdir -p config/keys

echo "Generating private key..."
openssl genrsa -out config/keys/auth_private.pem 2048

echo "Generating public key..."
openssl rsa -in config/keys/auth_private.pem -pubout -out config/keys/auth_public.pem

echo "Keys generated successfully!"
echo "Private key: config/keys/auth_private.pem"
echo "Public key: config/keys/auth_public.pem"
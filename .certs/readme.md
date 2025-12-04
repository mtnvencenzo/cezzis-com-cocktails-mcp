# Regenerate ssl
```bash
# 1. Generate the certificate with Subject Alternative Names (SANs):
openssl req -x509 -nodes -days 9999 -newkey rsa:2048 -keyout cocktails-mcp.key -out cocktails-mcp.crt -config cocktails-mcp.conf -extensions v3_req

# 2. Install to system trust store (Linux):
sudo cp ./cocktails-mcp.crt /usr/local/share/ca-certificates/cocktails-mcp.crt
sudo update-ca-certificates

# 3. Install to Chrome's certificate database (NSS):
sudo apt update && sudo apt install -y libnss3-tools
certutil -d sql:$HOME/.pki/nssdb -A -t "CP,CP," -n "cocktails-mcp" -i ./cocktails-mcp.crt

# 4. Verify it was added:
certutil -d sql:$HOME/.pki/nssdb -L | grep cocktails-mcp

# 5. Optionally convert to a pfx for use with .net and kestrel
openssl pkcs12 -export -out cocktails-mcp.pfx -inkey cocktails-mcp.key -in cocktails-mcp.crt -passout pass:password

# 6. Make sure everythings readable by all users
chmod 644 ./cocktails-mcp.crt
chmod 644 ./cocktails-mcp.key
chmod 644 ./cocktails-mcp.pfx
```
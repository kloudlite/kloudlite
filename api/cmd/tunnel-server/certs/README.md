# TLS Certificates

Place your TLS certificate and key files in this directory:

- `tls.crt` - TLS certificate file
- `tls.key` - TLS private key file

## Generate Self-Signed Certificate (for testing only)

```bash
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout tls.key \
  -out tls.crt \
  -days 365 \
  -subj "/CN=localhost"
```

## Production

For production, use certificates from a trusted CA like Let's Encrypt.

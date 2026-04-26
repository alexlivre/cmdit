# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | ✅ Currently active |

## Reporting a Vulnerability

If you discover a security vulnerability within cmdit, please report it responsibly.

**Do NOT** create a public GitHub issue for security vulnerabilities.

Instead, please send details to the maintainer via:

1. **GitHub Security Advisories** — https://github.com/alexlivre/cmdit/security/advisories/new
2. **Email** — (maintainer's contact)

Please include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested fixes (optional)

### What to expect

- Acknowledgment within 48 hours
- Regular updates on the progress
- Credit in the security advisory (if desired)
- Public disclosure after a fix is released

## Security Best Practices

cmdit is a terminal text editor. Keep these practices in mind:

- **Clipboard:** cmdit currently uses an internal clipboard. OSC52 (system clipboard via terminal escape codes) is planned for v2.
- **File access:** cmdit reads and writes files as directed. It does not execute external commands.
- **Remote sessions:** When using cmdit over SSH, ensure your connection is encrypted.
- **Unsigned binaries:** We are working on binary signing for future releases.

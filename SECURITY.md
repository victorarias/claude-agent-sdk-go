# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| main    | :white_check_mark: |
| < 1.0   | :x:                |

**Note:** This project is currently in pre-release (version < 1.0). Security updates are provided on a best-effort basis for the main branch only.

## Reporting a Vulnerability

We take the security of Claude Agent SDK Go seriously. If you discover a security vulnerability, please report it responsibly.

### Reporting Process

**DO NOT** create a public GitHub issue for security vulnerabilities.

Instead, please use one of the following methods:

1. **GitHub Security Advisories** (Preferred)
   - Go to the [Security tab](https://github.com/victorarias/claude-agent-sdk-go/security/advisories)
   - Click "Report a vulnerability"
   - Provide detailed information about the vulnerability

2. **Email**
   - Send details to: [security contact needed]
   - Include "SECURITY" in the subject line
   - Encrypt sensitive details using our PGP key (if available)

### What to Include

When reporting a vulnerability, please include:

- Description of the vulnerability
- Steps to reproduce the issue
- Affected versions
- Potential impact and severity
- Any suggested fixes or mitigations (if available)

### Response Timeline

- **Initial Response:** Within 48 hours of receiving your report
- **Status Update:** Within 7 days with our assessment and planned action
- **Resolution:** Varies based on complexity; we aim to release patches as quickly as possible

### Disclosure Policy

- We follow responsible disclosure practices
- Security vulnerabilities will be disclosed publicly only after:
  - A fix has been developed and released
  - Users have had reasonable time to upgrade (typically 2-4 weeks)
  - Coordination with the reporter on disclosure timing

### Security Contact

For security-related questions or concerns:

- **GitHub:** [@victorarias](https://github.com/victorarias)
- **Email:** [security contact needed]

## Security Best Practices

When using this SDK:

1. **Keep Dependencies Updated:** Regularly update to the latest version to receive security patches
2. **API Keys:** Never commit API keys or credentials to version control
3. **Input Validation:** Validate and sanitize all user inputs before passing to the SDK
4. **Error Handling:** Implement proper error handling to avoid exposing sensitive information
5. **Subprocess Security:** The SDK spawns Claude CLI subprocesses - ensure your environment is secure
6. **MCP Servers:** When using MCP servers, only connect to trusted servers with validated configurations

## Known Security Considerations

- **Subprocess Execution:** This SDK spawns the Claude CLI as a subprocess. Ensure the Claude CLI binary is from a trusted source.
- **File System Access:** The SDK may grant file system access to Claude. Use appropriate permission modes and working directory restrictions.
- **Network Communication:** MCP servers may communicate over network. Validate server configurations and use secure connections.

## Security Updates

Security advisories and updates will be published:

- In GitHub Security Advisories
- In release notes for security patches
- On the project's main README when significant

Thank you for helping keep Claude Agent SDK Go and its users safe!

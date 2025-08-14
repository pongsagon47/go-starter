# Security Policy

## 🔒 Supported Versions

| Version | Supported |
| ------- | --------- |
| 1.x.x   | ✅ Yes    |
| < 1.0   | ❌ No     |

## 🚨 Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

### 📧 Contact

- **Email**: security@pongsagon47.dev
- **Response Time**: Within 48 hours

### 🔍 What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### 🛡️ Security Best Practices

When using Go Starter:

#### **🔐 Environment Variables**

- Never commit `.env` files
- Use strong passwords and secrets
- Rotate API keys regularly
- Use different secrets for each environment

#### **🗄️ Database Security**

- Use SSL/TLS connections in production
- Create dedicated database users with minimal permissions
- Enable database audit logging
- Regular security updates

#### **🚀 Redis Security**

- Use Redis AUTH password
- Bind Redis to localhost in development
- Use SSL/TLS for Redis connections
- Regular Redis security updates

#### **🌐 API Security**

- Enable rate limiting (included)
- Use HTTPS in production
- Validate all input data
- Implement proper authentication
- Use CORS appropriately

#### **🐳 Docker Security**

- Don't run containers as root
- Use official base images
- Scan images for vulnerabilities
- Keep base images updated

## 🔒 Security Features Included

- ✅ **Rate Limiting** - DDoS protection
- ✅ **CORS** - Cross-origin protection
- ✅ **Helmet** - Security headers
- ✅ **Input Validation** - Data sanitization
- ✅ **Error Handling** - No information leakage
- ✅ **Logging** - Security audit trail

## 📚 Security Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Guide](https://github.com/OWASP/Go-SCP)
- [Redis Security](https://redis.io/topics/security)
- [Docker Security](https://docs.docker.com/engine/security/)

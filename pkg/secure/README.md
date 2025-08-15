# 🔒 Secure Package

AES-GCM encryption/decryption utilities for securing sensitive data at rest and in transit with authenticated encryption.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Encryption/Decryption](#encryption-decryption)
- [Examples](#examples)
- [Security Considerations](#security-considerations)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in go-starter
import "go-starter/pkg/secure"
```

## ⚡ Quick Start

### Basic Encryption/Decryption

```go
package main

import (
    "fmt"
    "go-starter/pkg/secure"
    "go-starter/config"
)

func main() {
    // Configuration
    cfg := &config.SecureConfig{
        Key: "your-base64-encoded-32-byte-key", // Generate with: openssl rand -base64 32
    }

    // Create secure instance
    s, err := secure.NewSecure(cfg)
    if err != nil {
        panic(err)
    }

    // Encrypt sensitive data
    plaintext := "sensitive user data"
    encrypted, err := s.Encrypt(plaintext)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Encrypted: %s\n", encrypted)

    // Decrypt data
    decrypted, err := s.Decrypt(encrypted)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Decrypted: %s\n", decrypted)
}
```

## ⚙️ Configuration

### **Secure Configuration Structure**

```go
type SecureConfig struct {
    Key string // Base64-encoded 32-byte encryption key
}
```

### **Environment Setup**

```env
# Generate a secure key:
# openssl rand -base64 32
ENCRYPTION_KEY=your-base64-encoded-key-here
```

### **Key Generation**

```bash
# Generate a new encryption key
openssl rand -base64 32

# Example output:
# K7gNU3sdo+OL0wNhqoVWhr3g6s1xYv72ol/pe/Unols=
```

### **Configuration in Application**

```go
// In config/config.go
type SecureConfig struct {
    Key string `env:"ENCRYPTION_KEY" validate:"required"`
}

func (cfg *Config) LoadSecureConfig() *SecureConfig {
    return &SecureConfig{
        Key: os.Getenv("ENCRYPTION_KEY"),
    }
}
```

## 🔐 Encryption/Decryption

### **Core Methods**

```go
type Secure struct {
    key       []byte
    gcm       cipher.AEAD
    nonceSize int
}

// Encrypt encrypts plaintext and returns base64-encoded ciphertext
func (s *Secure) Encrypt(plaintext string) (string, error)

// Decrypt decrypts base64-encoded ciphertext and returns plaintext
func (s *Secure) Decrypt(encodedCiphertext string) (string, error)
```

### **How It Works**

1. **AES-256-GCM**: Uses Advanced Encryption Standard with Galois/Counter Mode
2. **Authenticated Encryption**: Provides both confidentiality and authenticity
3. **Random Nonce**: Each encryption uses a unique random nonce
4. **Base64 Encoding**: Output is base64-encoded for safe storage/transmission

### **Security Features**

- ✅ **AES-256 encryption** - Industry standard symmetric encryption
- ✅ **GCM mode** - Authenticated encryption (detects tampering)
- ✅ **Random nonces** - Each encryption is unique
- ✅ **Constant-time operations** - Resistant to timing attacks
- ✅ **Memory safety** - Secure key handling

## 💡 Examples

### **1. User PII Encryption**

```go
type UserService struct {
    secure *secure.Secure
}

func NewUserService(secureConfig *config.SecureConfig) (*UserService, error) {
    s, err := secure.NewSecure(secureConfig)
    if err != nil {
        return nil, err
    }

    return &UserService{
        secure: s,
    }, nil
}

func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    // Encrypt sensitive data before storing
    encryptedSSN, err := s.secure.Encrypt(req.SSN)
    if err != nil {
        return nil, fmt.Errorf("failed to encrypt SSN: %w", err)
    }

    encryptedPhone, err := s.secure.Encrypt(req.Phone)
    if err != nil {
        return nil, fmt.Errorf("failed to encrypt phone: %w", err)
    }

    user := &User{
        ID:           uuid.New().String(),
        Name:         req.Name,
        Email:        req.Email, // Email can be searchable, so not encrypted
        SSN:          encryptedSSN,
        Phone:        encryptedPhone,
        CreatedAt:    time.Now(),
    }

    return user, s.userRepo.Create(user)
}

func (s *UserService) GetUserWithDecryptedData(id string) (*UserResponse, error) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        return nil, err
    }

    // Decrypt sensitive data for authorized access
    decryptedSSN, err := s.secure.Decrypt(user.SSN)
    if err != nil {
        return nil, fmt.Errorf("failed to decrypt SSN: %w", err)
    }

    decryptedPhone, err := s.secure.Decrypt(user.Phone)
    if err != nil {
        return nil, fmt.Errorf("failed to decrypt phone: %w", err)
    }

    return &UserResponse{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        SSN:       decryptedSSN,
        Phone:     decryptedPhone,
        CreatedAt: user.CreatedAt,
    }, nil
}
```

### **2. Payment Information Security**

```go
type PaymentService struct {
    secure *secure.Secure
}

func (s *PaymentService) StorePaymentMethod(userID string, req PaymentMethodRequest) (*PaymentMethod, error) {
    // Encrypt sensitive payment data
    encryptedCardNumber, err := s.secure.Encrypt(req.CardNumber)
    if err != nil {
        return nil, fmt.Errorf("failed to encrypt card number: %w", err)
    }

    // Note: In production, you should use a payment processor's vault
    // and never store actual card numbers

    paymentMethod := &PaymentMethod{
        ID:              uuid.New().String(),
        UserID:          userID,
        CardLast4:       req.CardNumber[len(req.CardNumber)-4:], // Store last 4 digits
        EncryptedCard:   encryptedCardNumber,
        ExpiryMonth:     req.ExpiryMonth,
        ExpiryYear:      req.ExpiryYear,
        CardType:        detectCardType(req.CardNumber),
        CreatedAt:       time.Now(),
    }

    return paymentMethod, s.paymentRepo.Create(paymentMethod)
}

func (s *PaymentService) ProcessPayment(paymentMethodID string, amount float64) (*Payment, error) {
    paymentMethod, err := s.paymentRepo.GetPaymentMethodByID(paymentMethodID)
    if err != nil {
        return nil, err
    }

    // Decrypt card for payment processing
    cardNumber, err := s.secure.Decrypt(paymentMethod.EncryptedCard)
    if err != nil {
        return nil, fmt.Errorf("failed to decrypt card number: %w", err)
    }

    // Process payment with external gateway
    result, err := s.paymentGateway.ProcessPayment(PaymentRequest{
        CardNumber:  cardNumber,
        ExpiryMonth: paymentMethod.ExpiryMonth,
        ExpiryYear:  paymentMethod.ExpiryYear,
        Amount:      amount,
    })

    // Clear sensitive data from memory
    cardNumber = ""

    if err != nil {
        return nil, err
    }

    return &Payment{
        ID:              uuid.New().String(),
        PaymentMethodID: paymentMethodID,
        Amount:          amount,
        Status:          result.Status,
        TransactionID:   result.TransactionID,
        CreatedAt:       time.Now(),
    }, nil
}
```

### **3. API Token Encryption**

```go
type TokenService struct {
    secure *secure.Secure
}

func (s *TokenService) CreateAPIToken(userID string, permissions []string) (*APIToken, error) {
    // Create token payload
    payload := TokenPayload{
        UserID:      userID,
        Permissions: permissions,
        IssuedAt:    time.Now(),
        ExpiresAt:   time.Now().Add(30 * 24 * time.Hour), // 30 days
    }

    // Serialize payload
    payloadJSON, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    // Encrypt token payload
    encryptedToken, err := s.secure.Encrypt(string(payloadJSON))
    if err != nil {
        return nil, fmt.Errorf("failed to encrypt token: %w", err)
    }

    token := &APIToken{
        ID:        uuid.New().String(),
        UserID:    userID,
        Token:     encryptedToken,
        ExpiresAt: payload.ExpiresAt,
        CreatedAt: time.Now(),
    }

    return token, s.tokenRepo.Create(token)
}

func (s *TokenService) ValidateToken(tokenString string) (*TokenPayload, error) {
    // Decrypt token
    decryptedPayload, err := s.secure.Decrypt(tokenString)
    if err != nil {
        return nil, fmt.Errorf("invalid token: %w", err)
    }

    // Parse payload
    var payload TokenPayload
    if err := json.Unmarshal([]byte(decryptedPayload), &payload); err != nil {
        return nil, fmt.Errorf("invalid token format: %w", err)
    }

    // Check expiration
    if time.Now().After(payload.ExpiresAt) {
        return nil, errors.New("token expired")
    }

    return &payload, nil
}
```

### **4. Configuration Secrets**

```go
type ConfigService struct {
    secure *secure.Secure
}

func (s *ConfigService) StoreSecret(key, value string) error {
    encrypted, err := s.secure.Encrypt(value)
    if err != nil {
        return fmt.Errorf("failed to encrypt secret: %w", err)
    }

    return s.configRepo.SetEncryptedValue(key, encrypted)
}

func (s *ConfigService) GetSecret(key string) (string, error) {
    encrypted, err := s.configRepo.GetEncryptedValue(key)
    if err != nil {
        return "", err
    }

    if encrypted == "" {
        return "", errors.New("secret not found")
    }

    decrypted, err := s.secure.Decrypt(encrypted)
    if err != nil {
        return "", fmt.Errorf("failed to decrypt secret: %w", err)
    }

    return decrypted, nil
}

// Usage
func initializeExternalServices(configService *ConfigService) error {
    // Retrieve encrypted API keys
    stripeKey, err := configService.GetSecret("stripe_secret_key")
    if err != nil {
        return err
    }

    twilioKey, err := configService.GetSecret("twilio_auth_token")
    if err != nil {
        return err
    }

    // Initialize services with decrypted keys
    stripeClient := stripe.NewClient(stripeKey)
    twilioClient := twilio.NewClient(twilioKey)

    return nil
}
```

### **5. Session Data Encryption**

```go
type SessionService struct {
    secure *secure.Secure
    cache  cache.Cache
}

func (s *SessionService) CreateSession(userID string, metadata map[string]interface{}) (string, error) {
    sessionData := SessionData{
        UserID:    userID,
        Metadata:  metadata,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }

    // Serialize session data
    data, err := json.Marshal(sessionData)
    if err != nil {
        return "", err
    }

    // Encrypt session data
    encryptedData, err := s.secure.Encrypt(string(data))
    if err != nil {
        return "", fmt.Errorf("failed to encrypt session: %w", err)
    }

    // Generate session ID
    sessionID := uuid.New().String()

    // Store encrypted session in cache
    err = s.cache.Set(fmt.Sprintf("session:%s", sessionID), encryptedData, 24*time.Hour)
    if err != nil {
        return "", err
    }

    return sessionID, nil
}

func (s *SessionService) GetSession(sessionID string) (*SessionData, error) {
    // Retrieve encrypted session
    encryptedData, err := s.cache.Get(fmt.Sprintf("session:%s", sessionID))
    if err != nil {
        return nil, err
    }

    // Decrypt session data
    decryptedData, err := s.secure.Decrypt(encryptedData)
    if err != nil {
        return nil, fmt.Errorf("failed to decrypt session: %w", err)
    }

    // Parse session data
    var sessionData SessionData
    if err := json.Unmarshal([]byte(decryptedData), &sessionData); err != nil {
        return nil, err
    }

    // Check expiration
    if time.Now().After(sessionData.ExpiresAt) {
        s.cache.Delete(fmt.Sprintf("session:%s", sessionID))
        return nil, errors.New("session expired")
    }

    return &sessionData, nil
}
```

## 🛡️ Security Considerations

### **1. Key Management**

```go
// ❌ DON'T: Hardcode keys
const encryptionKey = "my-secret-key"

// ✅ DO: Use environment variables
key := os.Getenv("ENCRYPTION_KEY")

// ✅ DO: Use key management services in production
// AWS KMS, HashiCorp Vault, etc.
```

### **2. Key Rotation**

```go
type SecureWithRotation struct {
    current *secure.Secure
    old     *secure.Secure
}

func (s *SecureWithRotation) Decrypt(ciphertext string) (string, error) {
    // Try current key first
    plaintext, err := s.current.Decrypt(ciphertext)
    if err == nil {
        return plaintext, nil
    }

    // Fallback to old key for migration
    return s.old.Decrypt(ciphertext)
}

func (s *SecureWithRotation) Encrypt(plaintext string) (string, error) {
    // Always encrypt with current key
    return s.current.Encrypt(plaintext)
}
```

### **3. Error Handling**

```go
func safeDecrypt(s *secure.Secure, ciphertext string) string {
    plaintext, err := s.Decrypt(ciphertext)
    if err != nil {
        // Log error but don't expose details
        logger.Error("Decryption failed", zap.Error(err))
        return "" // Return empty string or default value
    }
    return plaintext
}
```

## 🎯 Best Practices

### **1. What to Encrypt**

```go
// ✅ ENCRYPT: Sensitive personal data
type User struct {
    ID       string
    Email    string // Searchable, don't encrypt
    Name     string // May be searchable, consider carefully
    SSN      string // ENCRYPT
    Phone    string // ENCRYPT
    Address  string // ENCRYPT
}

// ✅ ENCRYPT: Financial data
type Payment struct {
    ID           string
    Amount       float64 // Don't encrypt (needed for queries)
    CardLast4    string  // Don't encrypt (safe to display)
    FullCardNum  string  // ENCRYPT (or better: don't store)
}

// ✅ ENCRYPT: API keys and secrets
type Config struct {
    DatabaseURL    string // ENCRYPT
    APIKey         string // ENCRYPT
    WebhookSecret  string // ENCRYPT
}
```

### **2. Database Integration**

```go
// Custom GORM data type for encrypted fields
type EncryptedString struct {
    secure    *secure.Secure
    plaintext string
}

func (es *EncryptedString) Scan(value interface{}) error {
    if value == nil {
        es.plaintext = ""
        return nil
    }

    encrypted := string(value.([]byte))
    plaintext, err := es.secure.Decrypt(encrypted)
    if err != nil {
        return err
    }

    es.plaintext = plaintext
    return nil
}

func (es EncryptedString) Value() (driver.Value, error) {
    if es.plaintext == "" {
        return nil, nil
    }

    return es.secure.Encrypt(es.plaintext)
}

func (es *EncryptedString) String() string {
    return es.plaintext
}

// Usage
type User struct {
    ID    string
    Email string
    SSN   EncryptedString `gorm:"type:text"`
    Phone EncryptedString `gorm:"type:text"`
}
```

### **3. Testing Encrypted Data**

```go
func TestUserEncryption(t *testing.T) {
    // Use test encryption key
    cfg := &config.SecureConfig{
        Key: "dGVzdC1rZXktZm9yLXVuaXQtdGVzdHMtMzItYnl0ZXM=", // test key
    }

    s, err := secure.NewSecure(cfg)
    require.NoError(t, err)

    original := "sensitive data"

    // Test encryption
    encrypted, err := s.Encrypt(original)
    require.NoError(t, err)
    assert.NotEqual(t, original, encrypted)
    assert.NotEmpty(t, encrypted)

    // Test decryption
    decrypted, err := s.Decrypt(encrypted)
    require.NoError(t, err)
    assert.Equal(t, original, decrypted)

    // Test that each encryption is unique (different nonces)
    encrypted2, err := s.Encrypt(original)
    require.NoError(t, err)
    assert.NotEqual(t, encrypted, encrypted2)
}
```

### **4. Memory Security**

```go
func processPayment(cardNumber string) error {
    // Clear sensitive data from memory when done
    defer func() {
        // Zero out the string (in Go, strings are immutable,
        // so this doesn't actually work - use []byte instead)
        cardNumber = ""
    }()

    // Better: use []byte for sensitive data
    cardBytes := []byte(cardNumber)
    defer func() {
        // Zero out the byte slice
        for i := range cardBytes {
            cardBytes[i] = 0
        }
    }()

    // Process payment...
    return nil
}
```

### **5. Monitoring and Alerting**

```go
func (s *UserService) GetUserPII(userID string) (*UserPII, error) {
    // Log access to sensitive data
    logger.Info("PII access requested",
        zap.String("user_id", userID),
        zap.String("accessor", getCurrentUser()),
        zap.String("reason", "customer_support"),
    )

    // Decrypt and return PII
    user, err := s.getUserWithDecryptedData(userID)
    if err != nil {
        logger.Error("Failed to decrypt user PII",
            zap.String("user_id", userID),
            zap.Error(err),
        )
        return nil, err
    }

    return user, nil
}
```

## ⚠️ Common Pitfalls

### **1. Key Management Issues**

```go
// ❌ DON'T: Store keys with data
// ❌ DON'T: Use weak keys
// ❌ DON'T: Reuse keys across environments

// ✅ DO: Use proper key management
func initSecure() (*secure.Secure, error) {
    key := os.Getenv("ENCRYPTION_KEY")
    if key == "" {
        return nil, errors.New("ENCRYPTION_KEY environment variable required")
    }

    return secure.NewSecure(&config.SecureConfig{Key: key})
}
```

### **2. Searchability Issues**

```go
// ❌ Problem: Can't search encrypted fields
SELECT * FROM users WHERE encrypted_ssn = 'encrypted_value'; // Won't work

// ✅ Solution: Use searchable hashes for lookup
type User struct {
    ID          string
    Email       string // Searchable
    SSNHash     string // SHA-256 hash for lookup
    EncryptedSSN string // Encrypted value
}

func findUserBySSN(ssn string) (*User, error) {
    hash := sha256.Sum256([]byte(ssn))
    hashString := hex.EncodeToString(hash[:])

    return db.Where("ssn_hash = ?", hashString).First(&User{}).Error
}
```

## 🔗 Related Packages

- [`config`](../../config/) - Secure configuration
- [`pkg/cache`](../cache/) - Encrypted session storage
- [`pkg/logger`](../logger/) - Security event logging

## 📚 Additional Resources

- [Go Cryptography Best Practices](https://golang.org/pkg/crypto/)
- [OWASP Cryptographic Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html)
- [AES-GCM Encryption](https://tools.ietf.org/html/rfc5116)

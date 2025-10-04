# OAuth Authentication Implementation for Cezzis Cocktails MCP Server

## Overview

This document outlines multiple approaches to implement OAuth authentication for the Cezzis Cocktails MCP server, enabling access to user-specific features like favorites, ratings, and profile management.

## Challenge

MCP servers typically run in headless environments without direct browser access, making traditional OAuth redirect flows challenging. The Cezzis API uses Azure Entra External Ids with OAuth2 authorization code flow with PKCE.

## Implementation Approaches

### 1. Device Code Flow (Recommended)

**Best for**: MCP servers that need user authentication without browser redirects.

**How it works**:
1. MCP server requests a device code from Azure Entra External Id Tenant
2. User opens browser separately and enters the device code
3. MCP server polls for tokens until authentication completes
4. Tokens are securely stored and used for API calls

**Files Created**:
- `/internal/auth/deviceflow.go` - Core device flow implementation
- `/internal/auth/storage.go` - Encrypted token storage
- `/internal/tools/authTools.go` - Authentication MCP tools
- `/internal/api/cocktailsapi/authFactory.go` - Authenticated API factory

**Key Features**:
- Device code authentication flow
- Encrypted token storage in user's home directory
- Automatic token loading on startup
- Background token polling
- Secure token management

### 2. Alternative Approaches

#### A. Configuration-Based Authentication
Store pre-obtained tokens in environment variables or config files:

```env
CEZZIS_ACCESS_TOKEN=eyJ0eXAiOiJKV1Q...
CEZZIS_REFRESH_TOKEN=eyJ0eXAiOiJKV1Q...
```

**Pros**: Simple implementation
**Cons**: Manual token management, security concerns

#### B. Local Web Server
Start a temporary web server for OAuth callback:

```go
// Start local server on random port
server := &http.Server{Addr: ":0"}
// Handle OAuth callback
// Extract authorization code and exchange for tokens
```

**Pros**: Full OAuth flow
**Cons**: Requires open ports, complexity

#### C. External Token Provider
Delegate authentication to external service:

```go
// Call external service that handles OAuth
tokenResponse := httpClient.Post("https://auth-service.com/token", authRequest)
```

**Pros**: Separation of concerns
**Cons**: Additional infrastructure needed

## Implementation Details

### Core Components

1. **AuthManager** (`/internal/auth/deviceflow.go`)
   - Manages device code flow
   - Handles token storage and retrieval
   - Provides authentication status

2. **TokenStorage** (`/internal/auth/storage.go`)
   - Encrypts and stores tokens securely
   - Loads tokens on startup
   - Manages token lifecycle

3. **Authentication Tools** (`/internal/tools/authTools.go`)
   - `auth_login` - Initiates device flow
   - `auth_status` - Check authentication status
   - `auth_logout` - Clear stored tokens

4. **Authenticated API Factory** (`/internal/api/cocktailsapi/authFactory.go`)
   - Creates API clients with OAuth headers
   - Handles both authenticated and basic requests

### Security Considerations

1. **Token Encryption**: AES-256-GCM encryption for stored tokens
2. **Secure Storage**: Tokens stored in user's home directory with restricted permissions
3. **Token Expiration**: Automatic handling of expired tokens
4. **Key Management**: Unique encryption keys per installation

### Usage Flow

1. **Initial Authentication**:
   ```
   User runs: auth_login tool
   → System provides device code and URL
   → User completes authentication in browser
   → Tokens automatically saved and loaded
   ```

2. **Subsequent API Calls**:
   ```
   Tools automatically use stored tokens
   → No re-authentication needed until tokens expire
   ```

3. **Authentication Check**:
   ```
   User runs: auth_status tool
   → Shows current authentication status
   ```

## Azure Entra External Id Tenant Configuration

Based on your API specification, the authentication endpoints are:

- **Authorization URL**: `https://login.cezzis.com/cezzis.onmicrosoft.com/sisu-p/oauth2/v2.0/authorize`
- **Token URL**: `https://login.cezzis.com/cezzis.onmicrosoft.com/sisu-p/oauth2/v2.0/token`
- **Device Code URL**: `https://login.cezzis.com/cezzis.onmicrosoft.com/sisu-p/oauth2/v2.0/devicecode`

**Scopes Required**:
- `https://cezzis.onmicrosoft.com/cocktailsapi/Account.Read`
- `https://cezzis.onmicrosoft.com/cocktailsapi/Account.Write`

**Client ID**: `84744194-da27-410f-ae0e-74f5589d4c96`

## Example Authenticated Tools

### Cocktail Rating Tool
```go
// Rate a cocktail (requires authentication)
cocktail_rate --cocktailId="clover-club" --stars=5
```

### Favorites Management Tool
```go
// Get user's favorite cocktails
cocktails_favorites_get

// Add/remove favorites
cocktails_favorites_manage --cocktailId="clover-club" --action="add"
```

### Profile Management Tool
```go
// Get user profile
account_profile_get

// Update profile settings
account_settings_update --givenName="John" --familyName="Doe"
```

## Error Handling

- **Authentication Required**: Clear error messages directing users to `auth_login`
- **Token Expiration**: Automatic detection and user notification
- **Network Errors**: Retry logic with exponential backoff
- **API Errors**: Meaningful error messages with suggested actions

## Testing Strategy

1. **Unit Tests**: Individual component testing
2. **Integration Tests**: End-to-end authentication flow
3. **Security Tests**: Token encryption/decryption
4. **Error Scenarios**: Network failures, expired tokens, invalid responses

## Deployment Considerations

1. **Environment Variables**: Configure Azure Entra External Id Tenant settings
2. **File Permissions**: Ensure secure token storage
3. **Network Access**: Required for authentication endpoints
4. **User Experience**: Clear instructions for device flow

## Next Steps

1. **Complete Integration**: Wire up authentication tools in main server
2. **Add Authenticated Endpoints**: Implement remaining API endpoints
3. **Testing**: Comprehensive testing of authentication flows
4. **Documentation**: User guides for authentication process
5. **Error Recovery**: Implement token refresh logic

This implementation provides a robust, secure authentication system that works well with the MCP protocol while maintaining good user experience and security practices.
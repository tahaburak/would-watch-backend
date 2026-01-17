# Troubleshooting Guide

## Database Connection Issues

### Error: "Unsupported or invalid secret format"

This error occurs when the `DATABASE_URL` connection string has an incorrect format, particularly when using Supabase pooler connections.

#### Symptoms
```
FATAL: Authentication error, reason: "Unsupported or invalid secret format"
```

#### Root Cause
The connection string is using the wrong username format. For Supabase pooler connections, the username must be `postgres.[project-ref]`, not `app_user.[project-ref]` or just `postgres`.

#### Solution

1. **Get your correct connection string:**
   - Go to your [Supabase Dashboard](https://app.supabase.com/project/gtjokreqhfsydfmtbtvg)
   - Navigate to: **Settings > Database**
   - Under "Connection string", select **"Transaction pooler"**
   - Copy the connection string

2. **Verify the format:**
   - **Correct (Pooler)**: `postgresql://postgres.gtjokreqhfsydfmtbtvg:[password]@aws-1-eu-west-1.pooler.supabase.com:6543/postgres`
   - **Incorrect**: `postgresql://app_user.gtjokreqhfsydfmtbtvg:[password]@...` ❌
   - **Incorrect**: `postgresql://postgres:[password]@aws-1-eu-west-1.pooler.supabase.com:6543/postgres` ❌

3. **Update your `.env` file:**
   ```bash
   DATABASE_URL=postgresql://postgres.gtjokreqhfsydfmtbtvg:[YOUR-PASSWORD]@aws-1-eu-west-1.pooler.supabase.com:6543/postgres
   ```

4. **Alternative: Use direct connection (port 5432)**
   If you prefer the direct connection (not recommended for serverless):
   ```bash
   DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.gtjokreqhfsydfmtbtvg.supabase.co:5432/postgres
   ```
   Note: Direct connection uses `postgres` as username (no project-ref).

#### Quick Check Script
Run the helper script to see the correct format:
```bash
./scripts/get-db-url.sh
```

### Error: "Database URL is required"

The `DATABASE_URL` environment variable is not set in your `.env` file.

**Solution:**
1. Copy `.env.example` to `.env` if it doesn't exist
2. Add your `DATABASE_URL` following the format above

### Error: "password authentication failed for user"

This error occurs when the password in your connection string is incorrect or needs URL encoding.

#### Symptoms
```
FATAL: password authentication failed for user "postgres" (SQLSTATE 28P01)
```

#### Solution

1. **Get your database password:**
   - Go to your [Supabase Dashboard](https://app.supabase.com/project/gtjokreqhfsydfmtbtvg)
   - Navigate to: **Settings > Database**
   - Under "Database password", you can:
     - View your current password (if you remember setting it)
     - Reset your password if needed
   - Copy the password exactly as shown

2. **URL-encode special characters:**
   If your password contains special characters, you may need to URL-encode them:
   - `@` → `%40`
   - `#` → `%23`
   - `$` → `%24`
   - `%` → `%25`
   - `&` → `%26`
   - `+` → `%2B`
   - `=` → `%3D`
   - `?` → `%3F`
   - `/` → `%2F`
   - ` ` (space) → `%20`

3. **Verify your connection string:**
   ```bash
   # Format should be:
   DATABASE_URL=postgresql://postgres.gtjokreqhfsydfmtbtvg:[PASSWORD]@aws-1-eu-west-1.pooler.supabase.com:6543/postgres
   ```

4. **Test with direct connection:**
   If pooler still doesn't work, try the direct connection:
   ```bash
   DATABASE_URL=postgresql://postgres:[PASSWORD]@db.gtjokreqhfsydfmtbtvg.supabase.co:5432/postgres
   ```

### Error: "Failed to ping database" (General)

The database server is unreachable or credentials are incorrect.

**Solution:**
1. Verify your database password is correct (see above)
2. Check if your Supabase project is active (not paused)
3. Verify network connectivity
4. Try the direct connection format if pooler doesn't work
5. Check if your IP is whitelisted (if IP restrictions are enabled)

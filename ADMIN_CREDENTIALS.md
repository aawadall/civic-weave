# Admin Credentials

## 🔐 Admin Account

**Email**: `admin@civicweave.com`  
**Password**: `admin123secure`

## 🚀 How to Login

### Via Frontend:
1. Go to: https://civicweave-frontend-162941711179.us-central1.run.app/login
2. Enter email: `admin@civicweave.com`
3. Enter password: `admin123secure`
4. Click "Sign In"

### Via API:
```bash
curl -X POST https://civicweave-backend-162941711179.us-central1.run.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@civicweave.com","password":"admin123secure"}'
```

Returns JWT token for authenticated requests.

## 🛡️ Admin Capabilities

As admin, you can:
- ✅ View all users (`GET /api/admin/users`)
- ✅ Manage user roles (`POST/DELETE /api/admin/users/:id/roles`)
- ✅ Create/manage roles (`/api/admin/roles`)
- ✅ Adjust volunteer skill weights (`PUT /api/admin/volunteers/:id/skills/:skillId/weight`)
- ✅ View system stats (`GET /api/admin/stats`)
- ✅ Manage campaigns (`/api/campaigns`)

## ⚠️ Security Notes

1. **Change Password**: Use a strong password in production
2. **Rotate Password**: Update in Secret Manager:
   ```bash
   echo "your-new-secure-password" | gcloud secrets versions add admin-password --data-file=-
   ```

3. **Admin Setup Endpoint**: Now disabled for security
   - Can only be re-enabled by uncommenting in `backend/cmd/server/main.go`
   - Not needed anymore since admin exists

## 📋 Admin Password Storage

The password is stored in:
- ✅ **Google Secret Manager**: `admin-password` secret
- ✅ **Environment Variable**: `ADMIN_PASSWORD` in Cloud Run backend
- ❌ **NOT in code** or version control

## 🔄 To Create Additional Admins

### Option 1: Via Existing Admin
1. Login as admin
2. Register a new user normally
3. Use admin API to assign admin role:
   ```bash
   curl -X POST https://.../api/admin/users/{user_id}/roles \
     -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
     -d '{"role_id":"ADMIN_ROLE_ID"}'
   ```

### Option 2: Via Database (Emergency Only)
```sql
-- Connect to Cloud SQL
-- Update user role
UPDATE users SET role = 'admin', email_verified = true WHERE email = 'newadmin@example.com';

-- Create admin profile
INSERT INTO admins (id, user_id, name) 
VALUES (uuid_generate_v4(), 'USER_ID_HERE', 'Admin Name');
```

## ✅ Verification

Admin user ID: `b9883807-f832-473f-b6fd-4b06ebc16440`  
Admin profile ID: `7dcc1ec4-b224-41a3-aae0-4f733dde901e`  
Created: October 12, 2025 02:17 UTC

---

**⚠️ IMPORTANT**: Change this password before going to production!


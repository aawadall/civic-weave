# Admin User Management Status

## ✅ Registration Safety

### Current Status (After Latest Update):
- ✅ **Rollback Protection**: If volunteer profile creation fails, user record is automatically deleted
- ✅ **Email Optional**: Registration succeeds even if email sending fails
- ✅ **Graceful Degradation**: Skill addition failures don't block registration

### Registration Flow:
1. Validate user doesn't exist
2. Create `user` record
3. ❗ **NEW**: Set rollback flag
4. Create `volunteer` profile (on failure → deletes user)
5. Add skills (failures are logged but don't block)
6. Send verification email (optional)
7. Return 201 Created

## 🔧 Existing Admin API Endpoints

### User Management (`/api/admin/users`)
```
GET    /api/admin/users                        # List all users
GET    /api/admin/users/:id/roles              # Get user's roles
POST   /api/admin/users/:id/roles              # Assign role to user
DELETE /api/admin/users/:id/roles/:roleId      # Revoke role from user
GET    /api/admin/users/:id/role-assignments   # Get role assignment history
```

### Role Management (`/api/admin/roles`)
```
GET    /api/admin/roles                        # List all roles
POST   /api/admin/roles                        # Create new role
GET    /api/admin/roles/:id                    # Get role details
PUT    /api/admin/roles/:id                    # Update role
DELETE /api/admin/roles/:id                    # Delete role
GET    /api/admin/roles/:id/users              # List users with this role
```

### Volunteer Management (`/api/admin/volunteers`)
```
PUT    /api/admin/volunteers/:id/skills/:skillId/weight  # Adjust skill weight
GET    /api/admin/volunteers/:id/weight-history          # View adjustment history
GET    /api/admin/volunteers/:id/skills                  # View volunteer skills
```

## ❌ Missing: Frontend UI

### What's Available:
- ✅ All backend endpoints exist
- ✅ RBAC system fully functional
- ✅ Role-based access control enforced

### What's Missing:
- ❌ Admin dashboard UI
- ❌ User management screen
- ❌ Role assignment interface
- ❌ Bulk user operations

## 🚀 Recommended: Create Admin User Management UI

### Proposed Features:

#### 1. **User List Page** (`/admin/users`)
```jsx
- Table showing all users
- Columns: Email, Name, Role, Status, Created Date
- Actions: View, Edit Roles, Delete (if no activity)
- Search/filter by role, email, date
```

#### 2. **User Detail Page** (`/admin/users/:id`)
```jsx
- User information (email, name, role)
- Role management
  - Current roles
  - Add role button
  - Remove role button
- Activity history
- Volunteer profile (if applicable)
- Delete user button (with confirmation)
```

#### 3. **Role Management Page** (`/admin/roles`)
```jsx
- List of all roles
- Create new role
- Edit role permissions
- See users assigned to each role
```

### Priority:
**HIGH** - Needed for:
- Managing incomplete registrations
- Promoting users to team_lead/admin
- Debugging registration issues
- User support

## 🔍 How to Handle Incomplete Registrations

### Option 1: Admin UI (Recommended)
Create a "Problem Users" section showing:
- Users without volunteer profiles
- Users with incomplete profiles
- One-click fix or delete

### Option 2: CLI Script
```bash
# backend/scripts/cleanup_orphaned_users.go
go run scripts/cleanup_orphaned_users.go
```

### Option 3: Automatic Cleanup
Add a database trigger or cron job to:
- Find users without profiles
- Older than 24 hours
- Auto-delete them

## 📝 Immediate Actions Needed:

### 1. **Test Registration** ✅
- Try registering at: https://civicweave-frontend-162941711179.us-central1.run.app/register
- Verify it completes successfully
- Check database for orphaned records

### 2. **Deploy Rollback Fix** 🔄
```bash
cd /home/arashad/src/CivicWeave
make build-push
make deploy-app
```

### 3. **Build Admin UI** (Next Sprint)
- Create `/admin/users` route
- Build user list component
- Add role management
- Implement user deletion

## 🐛 Testing Scenarios:

### Test 1: Normal Registration
```bash
curl -X POST https://civicweave-backend-162941711179.us-central1.run.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User",
    "consent_given": true
  }'
```

**Expected**: 201 Created, user + volunteer created

### Test 2: Check for Orphaned Users
```sql
-- Find users without volunteer profiles
SELECT u.* FROM users u
LEFT JOIN volunteers v ON u.id = v.user_id
WHERE u.role = 'volunteer' AND v.id IS NULL;
```

### Test 3: Rollback Verification
Check backend logs for:
```
⚠️  Registration failed at volunteer creation, rolling back user: email@example.com
```

## 📊 Current Production Status:

- **Backend**: Revision `civicweave-backend-00033-rfl`
- **Email Feature**: DISABLED (`ENABLE_EMAIL=false`)
- **Rollback**: ⚠️ **NOT YET DEPLOYED** (needs build+push)
- **Admin UI**: ❌ Not built yet

## 🎯 Next Steps:

1. ✅ **Deploy rollback fix** (you have the code, needs deployment)
2. ⚠️ **Check for existing orphaned users** (run SQL query)
3. 🔨 **Build admin user management UI** (1-2 day task)
4. ✅ **Enable email** once Mailgun is configured properly


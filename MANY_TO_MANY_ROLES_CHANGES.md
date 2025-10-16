# Many-to-Many Roles Migration

## Summary
Successfully migrated the CivicWeave application from a single-role-per-user system to a many-to-many relationship between users and roles. Users can now have multiple roles assigned to them simultaneously.

## Changes Made

### 1. Database Migration
**File:** `backend/migrations/010_remove_users_role_column.sql`
- Migrates existing role data from `users.role` column to `user_roles` junction table
- Drops the old `role` column from the `users` table
- Drops the index on the deprecated role column
- Includes rollback logic (DOWN migration)

### 2. Models Updated

#### `backend/models/user.go`
- Removed deprecated `Role` field from `User` struct
- Updated all query methods to remove role column references:
  - `Create()` - removed role parameter
  - `GetByID()` - removed role from scan
  - `GetByEmail()` - removed role from scan
  - `Update()` - removed role parameter
  - `ListAllUsers()` - removed role from scan

#### `backend/models/user_queries.go`
- Updated all SQL queries to exclude the `role` column:
  - `userCreateQuery` - removed role from INSERT
  - `userGetByIDQuery` - removed role from SELECT
  - `userGetByEmailQuery` - removed role from SELECT
  - `userUpdateQuery` - removed role from UPDATE
  - `userListAllQuery` - removed role from SELECT

#### `backend/models/validation.go`
- Removed `ValidateRole()` call from `ValidateUser()` function
- The `ValidateRole()` function itself remains for backward compatibility but is no longer used

### 3. Middleware Updated

#### `backend/middleware/auth.go`
- **Claims struct:** Removed deprecated `Role` field, kept only `Roles []string`
- **AuthRequired:** Removed setting of deprecated `user_role` context, now only sets `user_roles`
- **RequireAnyRole:** Removed fallback to legacy role system, now only checks many-to-many roles
- **OptionalAuth:** Updated to only set `user_roles` context
- **GenerateJWT:** 
  - Now requires roles to be fetched successfully (no fallback)
  - Returns error if user has no roles assigned
  - Removed deprecated single `Role` field from JWT claims
- **GetUserFromContext:** Removed deprecated `Role` field from UserContext
- **UserContext struct:** Removed deprecated `Role` field

### 4. Handlers Updated

#### `backend/handlers/auth_full.go`
- Added `RoleService` field to `AuthHandler` struct
- Updated `NewAuthHandler()` constructor to accept `roleService` parameter
- **Register endpoint:** Now assigns "volunteer" role after user creation with proper error handling and rollback
- **Login endpoint:** 
  - Fetches user roles via `GetUserRoles()`
  - Checks for "volunteer" or "admin" role using boolean flags
  - Removed deprecated `role` field from response, now only returns `roles` array
- Updated logging to remove role-specific messages

#### `backend/handlers/google_oauth.go`
- Added `RoleService` field to `GoogleOAuthHandler` struct
- Updated `NewGoogleOAuthHandler()` constructor to accept `roleService` parameter
- **createUserFromGoogle:** Now assigns "volunteer" role after user creation with proper error handling and rollback
- Updated to check roles via many-to-many relationship instead of single role field

#### `backend/handlers/admin_setup.go`
- Added `RoleService` field to `AdminSetupHandler` struct
- Updated `NewAdminSetupHandler()` constructor to accept `roleService` parameter
- **CreateAdmin endpoint:** Now assigns "admin" role after user creation with proper error handling and rollback

### 5. Command-Line Tools Updated

#### `backend/cmd/seed/main.go`
- Removed deprecated `Role` field when creating admin user
- Now assigns "admin" role via `roleService.AssignRoleToUser()`
- Removed `migrateExistingUsers()` function (migration now handled by SQL migration 010)

#### `backend/cmd/seed-admin/main.go`
- Removed deprecated `Role` field when creating admin user
- Now assigns "admin" role via `roleService.AssignRoleToUser()` after user creation

#### `backend/cmd/server/main.go`
- Updated `NewAuthHandler()` call to include `roleService` parameter
- Updated `NewGoogleOAuthHandler()` call to include `roleService` parameter
- Updated handler initialization check to include `roleService != nil`

## Migration Path

### For Existing Installations

1. **Before running the migration:**
   - Backup your database
   - Ensure all users have a valid role in the old `users.role` column

2. **Run the migration:**
   ```bash
   make db-migrate
   ```

3. **Verify the migration:**
   - Check that all users have corresponding entries in `user_roles` table
   - Verify that the `role` column has been removed from `users` table

4. **Test authentication:**
   - Test login with existing users
   - Verify JWT tokens contain `roles` array
   - Test role-based access control

### For New Installations

No special steps needed - just run migrations normally and seed the database:
```bash
make db-migrate
make db-seed
```

## API Changes

### JWT Token Structure

**Before:**
```json
{
  "user_id": "...",
  "email": "...",
  "role": "volunteer",
  "roles": ["volunteer"]
}
```

**After:**
```json
{
  "user_id": "...",
  "email": "...",
  "roles": ["volunteer"]
}
```

### User Response Structure

**Before:**
```json
{
  "id": "...",
  "email": "...",
  "role": "volunteer",
  "roles": ["volunteer"]
}
```

**After:**
```json
{
  "id": "...",
  "email": "...",
  "roles": ["volunteer"]
}
```

## Backward Compatibility

- The migration is **NOT backward compatible** with the old single-role system
- Frontend applications must be updated to use the `roles` array instead of the `role` field
- JWT tokens issued before the migration will be invalid and users will need to re-login

## Testing Checklist

- [ ] User registration assigns "volunteer" role correctly
- [ ] User login returns roles array in response
- [ ] JWT tokens contain roles array
- [ ] Google OAuth registration assigns "volunteer" role correctly
- [ ] Admin creation assigns "admin" role correctly
- [ ] Role-based middleware (`RequireAnyRole`, `RequireAllRoles`) works correctly
- [ ] Users can have multiple roles assigned
- [ ] Database migration successfully migrates existing role data
- [ ] Rollback migration works correctly

## Notes

- Users must have at least one role assigned to successfully generate JWT tokens
- All user creation flows (registration, OAuth, admin setup) now explicitly assign roles
- The `ValidateRole()` function in `validation.go` is kept for potential future use but is no longer called
- Error handling includes rollback of user creation if role assignment fails




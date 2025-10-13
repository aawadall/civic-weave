# Testing Guide - Email Verification & Messaging Center

## Quick Test Checklist

### ✅ Email Auto-Verification (Feature Flag)

**Setup:**
```bash
# In your .env file
ENABLE_EMAIL=false
```

**Test Steps:**
1. Start the backend: `cd backend && go run cmd/server/main.go`
2. Start the frontend: `cd frontend && npm run dev`
3. Register a new user at `http://localhost:5173/register`
4. Try to login immediately (should work without email verification)
5. ✅ **Expected:** User can login without verifying email

**With Email Enabled:**
```bash
# In your .env file
ENABLE_EMAIL=true
MAILGUN_API_KEY=your_key
MAILGUN_DOMAIN=your_domain
```
6. Register another user
7. Try to login (should be blocked)
8. ✅ **Expected:** Error message "Please verify your email before logging in"

### ✅ Messaging Center Notification Badge

**Prerequisites:**
- Have 2 user accounts
- Have at least 1 project
- Both users should be team members of the project

**Test Steps:**

1. **Login as User A**
   - Navigate to header (top right)
   - ✅ **Expected:** See mail icon (envelope) with no badge (or 0)

2. **Login as User B (in another browser/incognito)**
   - Go to a project both users share
   - Navigate to Messages tab
   - Send a message: "Hello from User B!"

3. **Back to User A's browser**
   - Wait up to 30 seconds (auto-refresh interval)
   - OR refresh the page
   - ✅ **Expected:** See red badge with "1" on the mail icon

4. **Click the Mail Icon**
   - ✅ **Expected:** Navigate to `/messages` page
   - ✅ **Expected:** See the project listed with "1 unread"

5. **Click the Project Card**
   - ✅ **Expected:** Navigate to project's message tab
   - ✅ **Expected:** See User B's message

6. **Return to Header**
   - ✅ **Expected:** Badge should show 0 or disappear (messages marked as read)

## Visual Verification

### Header with Notification Badge
```
┌──────────────────────────────────────────────────────────┐
│ CW CivicWeave    Home  Dashboard  Projects               │
│                                                           │
│                              [📧¹] Welcome, John   Logout │
│                               ↑                          │
│                          Mail Icon with Badge            │
└──────────────────────────────────────────────────────────┘
```

### Message Center Page
```
┌──────────────────────────────────────────────────────────┐
│ Message Center                                            │
│ View all your project messages in one place              │
│                                                           │
│ ┌────────────────────────────────────────────────────┐  │
│ │ Community Garden Project              [2 unread]   │  │
│ │ Help organize our local garden                     │  │
│ └────────────────────────────────────────────────────┘  │
│                                                           │
│ ┌────────────────────────────────────────────────────┐  │
│ │ Food Drive Initiative                 [1 unread]   │  │
│ │ Coordinate food distribution                       │  │
│ └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

## Troubleshooting

### Badge Not Showing
- Check browser console for API errors
- Verify user is authenticated
- Verify user is a team member of at least one project
- Check that messages exist and are unread

### Auto-Verification Not Working
- Verify `ENABLE_EMAIL=false` is set in backend `.env`
- Restart backend server after changing `.env`
- Check backend logs for user creation

### API Errors
- Ensure backend is running on `http://localhost:8080`
- Check `VITE_API_BASE_URL` in frontend `.env`
- Verify JWT token is valid (check Network tab in browser DevTools)

## Component Architecture

```
Header.jsx
  └── MessagesIcon.jsx
       ├── Fetches: GET /api/messages/unread-counts
       ├── Polls every 30 seconds
       ├── Links to: /messages
       └── NotificationBadge.jsx
            └── Shows count when > 0

MessageCenter.jsx (route: /messages)
  ├── Fetches: GET /api/messages/unread-counts
  ├── Fetches: GET /api/projects/:id (for each project)
  └── Links to: /projects/:id/messages
```

## Performance Notes

- Badge updates every 30 seconds automatically
- API calls are lightweight (only counts, not full messages)
- Badge count is sum of all unread messages across all projects
- Maximum badge display: "99+"


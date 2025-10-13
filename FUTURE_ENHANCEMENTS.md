# Future Enhancements - Messaging & Notifications

## Email Notifications for Unread Messages

### Overview
Allow users to opt-in to receive email notifications when they have unread project messages.

### Implementation Plan

#### Database Changes
Add user preference column:
```sql
ALTER TABLE volunteers ADD COLUMN email_notifications_enabled BOOLEAN DEFAULT false;
```

Or create a more flexible preferences table:
```sql
CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email_on_message BOOLEAN DEFAULT false,
    email_digest_frequency VARCHAR(20) DEFAULT 'none' CHECK (frequency IN ('none', 'immediate', 'daily', 'weekly')),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Backend Changes

1. **User Preferences API**
   - `GET /api/me/preferences` - Get user notification preferences
   - `PUT /api/me/preferences` - Update notification preferences

2. **Email Service Extension** (`backend/services/email_service.go`)
   - Add `SendMessageNotificationEmail(to, projectName, messagePreview string)` method
   - Template for message notification emails

3. **Message Handler Modification** (`backend/handlers/message_handler.go`)
   - After creating a message, check recipient preferences
   - Send email notifications to opted-in team members
   - Optionally: batch/digest emails instead of immediate

4. **Background Worker (Optional)**
   - Daily/weekly digest of unread messages
   - Use cron job or task queue (like `github.com/hibiken/asynq`)

#### Frontend Changes

1. **Profile Settings Page**
   - Add notification preferences section
   - Toggle for "Email me when I receive messages"
   - Dropdown for frequency (immediate, daily digest, weekly digest)

2. **UI Component** (`frontend/src/components/NotificationSettings.jsx`)
   ```jsx
   - Checkbox: "Email notifications for new messages"
   - Radio buttons: Frequency (immediate, daily, weekly, none)
   ```

### Email Templates

**Immediate Notification:**
```
Subject: New message in [Project Name]

Hi [User Name],

You have a new message in [Project Name]:

"[Message preview...]"

View message: [Link to project messages]

---
Manage notification preferences: [Link to settings]
```

**Daily Digest:**
```
Subject: You have [N] unread messages

Hi [User Name],

You have unread messages in [N] projects:

1. [Project Name] - [N] unread messages
2. [Project Name] - [N] unread messages

View all messages: [Link to message center]

---
Manage notification preferences: [Link to settings]
```

### Configuration
Add to `.env`:
```
ENABLE_MESSAGE_EMAIL_NOTIFICATIONS=true
MESSAGE_NOTIFICATION_FROM_EMAIL=notifications@civicweave.com
```

### Considerations

**Rate Limiting:**
- Avoid email spam if many messages sent quickly
- Implement cooldown period (e.g., max 1 email per hour per project)

**Privacy:**
- Don't include full message content in email (just preview)
- Respect user's email verification status
- Allow users to unsubscribe easily

**Performance:**
- Don't block message sending on email delivery
- Use background job queue for email sending
- Handle Mailgun failures gracefully

**User Experience:**
- Default to `false` (opt-in, not opt-out)
- Clear messaging about what notifications they'll receive
- Easy toggle in profile settings

---

## Real-Time Messaging (Redis Pub/Sub)

### Overview
Replace 30-second polling with WebSocket/SSE for instant message notifications.

### Implementation Plan

#### Backend Changes

1. **WebSocket Handler** (`backend/handlers/websocket_handler.go`)
   - Create WebSocket endpoint `/api/ws`
   - Authenticate via JWT token in connection
   - Maintain user connections in memory

2. **Redis Pub/Sub Service** (`backend/services/message_pubsub.go`)
   ```go
   type MessagePubSub struct {
       redis *redis.Client
   }
   
   func (s *MessagePubSub) PublishMessage(projectID uuid.UUID, message *models.ProjectMessage)
   func (s *MessagePubSub) Subscribe(projectID uuid.UUID) <-chan *redis.Message
   ```

3. **Modify Message Handler**
   - After creating message, publish to Redis channel
   - Channel format: `project:{projectID}:messages`

4. **WebSocket Manager**
   - Subscribe users to their project channels
   - Broadcast new messages to connected clients
   - Handle reconnection logic

#### Frontend Changes

1. **WebSocket Hook** (`frontend/src/hooks/useWebSocket.js`)
   ```jsx
   const { connected, lastMessage } = useWebSocket()
   ```

2. **Update MessagesIcon Component**
   - Replace polling with WebSocket listener
   - Update badge in real-time on new messages

3. **Connection Status Indicator**
   - Show "connected" / "disconnected" status
   - Auto-reconnect on connection loss

### Configuration
```
ENABLE_WEBSOCKET=true
REDIS_PUBSUB_ENABLED=true
```

### Benefits
- ✅ Instant updates (no 30-second delay)
- ✅ Reduced server load (no constant polling)
- ✅ Better user experience

### Trade-offs
- ❌ More complex infrastructure
- ❌ Need to handle WebSocket scaling (sticky sessions or Redis adapter)
- ❌ Requires Redis to be running and healthy

---

## Direct User-to-User Messaging

### Overview
Add private direct messaging between users (separate from project messages).

### Database Schema
```sql
CREATE TABLE direct_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_text TEXT NOT NULL,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_by_sender BOOLEAN DEFAULT false,
    deleted_by_recipient BOOLEAN DEFAULT false
);

CREATE INDEX idx_direct_messages_from ON direct_messages(from_user_id);
CREATE INDEX idx_direct_messages_to ON direct_messages(to_user_id);
CREATE INDEX idx_direct_messages_read ON direct_messages(to_user_id, read_at) WHERE read_at IS NULL;
```

### API Endpoints
- `POST /api/messages/direct` - Send direct message
- `GET /api/messages/direct` - List conversations
- `GET /api/messages/direct/:userId` - Get conversation with specific user
- `GET /api/messages/direct/unread-count` - Get unread DM count

### UI Updates
- Add "Direct Messages" tab in Message Center
- Show user avatar + name in DM list
- Separate badge for DMs vs project messages
- Conversation view (like chat interface)

---

## System Notifications

### Overview
Allow system-generated and admin-created notifications to users. This includes automatic notifications when volunteers match projects.

### Use Cases

**1. Matching Notifications (High Priority)**
- When batch job calculates volunteer-project matches
- Notify volunteers when they're in top N candidates for a project
- Example: "You're a top match for 'Community Garden Project'!"

**2. Admin Notifications**
- System-wide announcements
- Targeted messages to specific user groups

### Database Schema
```sql
CREATE TABLE system_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    notification_type VARCHAR(50) DEFAULT 'info' CHECK (type IN ('info', 'warning', 'success', 'error')),
    target_audience VARCHAR(50) DEFAULT 'all' CHECK (audience IN ('all', 'volunteers', 'team_leads', 'admins', 'custom')),
    target_user_ids JSONB DEFAULT '[]',
    created_by_admin_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);

CREATE TABLE notification_reads (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_id UUID NOT NULL REFERENCES system_notifications(id) ON DELETE CASCADE,
    read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, notification_id)
);
```

### Features
- Admin panel to create system notifications
- Target specific user groups or all users
- Auto-expire notifications after date
- Show notification bell icon in header
- Toast/banner for important notifications
- Mark as read functionality

### Integration with Matching System

**Project-Level Configuration**

Add `auto_notify_matches` (or `hunt_mode`) field to projects table:
```sql
ALTER TABLE projects ADD COLUMN auto_notify_matches BOOLEAN DEFAULT false;
-- OR
ALTER TABLE projects ADD COLUMN hunt_mode BOOLEAN DEFAULT false;
```

**Project Creation UI** (`frontend/src/pages/projects/CreateProjectPage.jsx`)
```jsx
<div className="form-group">
  <label>
    <input
      type="checkbox"
      name="auto_notify_matches"
      checked={formData.auto_notify_matches}
      onChange={handleChange}
    />
    <span className="ml-2">
      Auto-recruit top matches
    </span>
  </label>
  <p className="text-sm text-secondary-600 mt-1">
    When enabled, volunteers who are top matches will be automatically notified 
    about this opportunity. Great for urgent or high-priority projects!
  </p>
</div>
```

**Batch Job Integration** (`backend/jobs/calculate_matches.py` or similar)

When calculating volunteer-project matches:
```python
# After calculating matches
projects = get_projects_with_hunt_mode_enabled()

for project in projects:
    top_candidates = get_top_n_candidates(project.id, n=10)
    
    # Only notify if project has hunt_mode enabled
    if project.auto_notify_matches:
        for volunteer_id in top_candidates:
            create_system_notification(
                user_id=volunteer_id,
                title="New Project Match!",
                message=f"You're a top candidate for '{project.title}'. Click to learn more!",
                notification_type="success",
                link=f"/projects/{project.id}",
                metadata={
                    "project_id": project.id,
                    "match_score": score,
                    "match_type": "distance_and_skills"
                }
            )
```

**Notification Service** (`backend/services/notification_service.go`)
```go
type NotificationService struct {
    db *sql.DB
}

func (s *NotificationService) CreateMatchNotification(
    volunteerID uuid.UUID,
    projectID uuid.UUID,
    matchScore float64,
) error {
    notification := &SystemNotification{
        Title: "New Project Match!",
        Message: fmt.Sprintf("You're a top candidate for this project (%.0f%% match)", matchScore*100),
        Type: "success",
        TargetUserIDs: []uuid.UUID{volunteerID},
        Metadata: map[string]interface{}{
            "project_id": projectID,
            "match_score": matchScore,
            "action_url": fmt.Sprintf("/projects/%s", projectID),
        },
    }
    return s.Create(notification)
}
```

**Configuration Options**
```env
# Matching Notifications (Global)
ENABLE_MATCH_NOTIFICATIONS=true  # Master switch for notification system
MATCH_NOTIFICATION_THRESHOLD=10  # Top N candidates to notify per project
MATCH_NOTIFICATION_MIN_SCORE=0.7  # Minimum match score (70%)
```

**Project-Level Settings** (per project)
- `auto_notify_matches` (boolean) - Enable/disable notifications for this specific project
- Set by project creator during project creation or in project settings
- Defaults to `false` (opt-in by project creator)
- Can be toggled on/off at any time

**Use Cases**
- ✅ **Hunt Mode ON**: Urgent projects, hard-to-fill positions, high-priority initiatives
- ❌ **Hunt Mode OFF**: Projects with enough applicants, low-urgency, passive recruitment

**UI Implementation**
- Bell icon in header (next to messages icon)
- Badge showing unread notification count
- Dropdown panel with recent notifications
- Click notification to navigate to relevant page
- "Mark all as read" action

**Notification Types**
- `match` - You're a top candidate for a project
- `application_accepted` - Your application was accepted
- `application_rejected` - Your application was declined
- `project_update` - Important updates to projects you're in
- `system` - System-wide announcements from admins

---

## Priority & Timeline Suggestion

**Phase 1 (Current):** ✅ Complete
- Basic message badge with polling
- Email verification feature flag

**Phase 2 (Next 1-2 sprints):**
- Email notifications for messages (opt-in)
- User profile notification settings

**Phase 3 (Future):**
- Real-time messaging (WebSocket + Redis Pub/Sub)
- Direct user-to-user messaging

**Phase 4 (Optional):**
- System notifications
- Message search & filtering
- Message attachments


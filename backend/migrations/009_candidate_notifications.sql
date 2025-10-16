-- UP
-- Candidate Notifications Tracking
-- Tracks which volunteers have been notified about which projects to prevent duplicate notifications

CREATE TABLE candidate_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    match_score DECIMAL(5,4) NOT NULL,
    notified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notification_batch_id UUID NOT NULL,
    UNIQUE(project_id, volunteer_id, notification_batch_id)
);

-- Add indexes for performance
CREATE INDEX idx_candidate_notifications_project_id ON candidate_notifications(project_id);
CREATE INDEX idx_candidate_notifications_volunteer_id ON candidate_notifications(volunteer_id);
CREATE INDEX idx_candidate_notifications_batch_id ON candidate_notifications(notification_batch_id);
CREATE INDEX idx_candidate_notifications_notified_at ON candidate_notifications(notified_at);

-- DOWN
DROP INDEX IF EXISTS idx_candidate_notifications_notified_at;
DROP INDEX IF EXISTS idx_candidate_notifications_batch_id;
DROP INDEX IF EXISTS idx_candidate_notifications_volunteer_id;
DROP INDEX IF EXISTS idx_candidate_notifications_project_id;
DROP TABLE IF EXISTS candidate_notifications;




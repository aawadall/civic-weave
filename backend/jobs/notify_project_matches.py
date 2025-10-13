#!/usr/bin/env python3
"""
Daily batch job to notify top matching candidates for projects.
Notifies both candidates and team leads about potential matches.
"""

import os
import sys
import psycopg2
import uuid
from datetime import datetime
from typing import List, Tuple
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Add the backend directory to the path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Configuration
TOP_K_CANDIDATES = int(os.getenv('TOP_K_CANDIDATES', '10'))
MIN_MATCH_SCORE = float(os.getenv('MIN_MATCH_SCORE', '0.6'))
SYSTEM_USER_ID = os.getenv('SYSTEM_USER_ID', '00000000-0000-0000-0000-000000000000')


def get_db_connection():
    """Get database connection from environment variables."""
    return psycopg2.connect(
        host=os.getenv('DB_HOST', 'localhost'),
        port=os.getenv('DB_PORT', '5432'),
        database=os.getenv('DB_NAME', 'civicweave'),
        user=os.getenv('DB_USER', 'postgres'),
        password=os.getenv('DB_PASSWORD', 'password')
    )


def get_recruiting_projects(cursor) -> List[Tuple]:
    """Get all projects that are actively recruiting."""
    query = """
        SELECT id, title, team_lead_id, created_by_admin_id
        FROM projects
        WHERE project_status IN ('recruiting', 'active')
        ORDER BY created_at DESC
    """
    cursor.execute(query)
    return cursor.fetchall()


def get_top_candidates(cursor, project_id: uuid.UUID, top_k: int, min_score: float, batch_id: uuid.UUID) -> List[Tuple]:
    """
    Get top K candidates for a project who haven't been notified in this batch.
    
    Returns list of (volunteer_id, user_id, match_score, volunteer_name)
    """
    query = """
        SELECT 
            v.id as volunteer_id,
            v.user_id,
            m.match_score,
            v.name as volunteer_name
        FROM volunteer_project_matches m
        JOIN volunteers v ON m.volunteer_id = v.id
        WHERE m.project_id = %s 
            AND m.match_score >= %s
            AND v.skills_visible = true
            AND NOT EXISTS (
                SELECT 1 FROM candidate_notifications cn
                WHERE cn.project_id = %s 
                    AND cn.volunteer_id = v.id
                    AND cn.notification_batch_id = %s
            )
        ORDER BY m.match_score DESC, m.matched_skill_count DESC
        LIMIT %s
    """
    cursor.execute(query, (project_id, min_score, project_id, batch_id, top_k))
    return cursor.fetchall()


def send_candidate_notification(cursor, project_id: uuid.UUID, project_title: str, 
                                volunteer_user_id: uuid.UUID, match_score: float):
    """Send notification message to a candidate volunteer."""
    message_text = f"""🎯 Great news! You're a top match for "{project_title}"!

Your skills align {int(match_score * 100)}% with this project's needs. 

Check out the project details and apply if you're interested in contributing!"""
    
    message_id = uuid.uuid4()
    system_user = uuid.UUID(SYSTEM_USER_ID)
    
    query = """
        INSERT INTO project_messages (id, project_id, sender_id, message_text, created_at)
        VALUES (%s, %s, %s, %s, NOW())
    """
    cursor.execute(query, (message_id, project_id, system_user, message_text))


def send_team_lead_notification(cursor, project_id: uuid.UUID, project_title: str,
                                team_lead_id: uuid.UUID, candidates: List[Tuple]):
    """Send summary notification to team lead about top candidates."""
    if not candidates:
        return
    
    # Build candidate list
    candidate_list = "\n".join([
        f"• {name} ({int(score * 100)}% match)"
        for _, _, score, name in candidates
    ])
    
    message_text = f"""📊 Top {len(candidates)} Candidate Matches for "{project_title}"

The matching system has identified these highly qualified volunteers:

{candidate_list}

These candidates have been notified about the project opportunity. Review their profiles and reach out if interested!"""
    
    message_id = uuid.uuid4()
    system_user = uuid.UUID(SYSTEM_USER_ID)
    
    query = """
        INSERT INTO project_messages (id, project_id, sender_id, message_text, created_at)
        VALUES (%s, %s, %s, %s, NOW())
    """
    cursor.execute(query, (message_id, project_id, system_user, message_text))


def record_notification(cursor, project_id: uuid.UUID, volunteer_id: uuid.UUID,
                       match_score: float, batch_id: uuid.UUID):
    """Record that a candidate has been notified about a project."""
    query = """
        INSERT INTO candidate_notifications 
        (id, project_id, volunteer_id, match_score, notification_batch_id, notified_at)
        VALUES (%s, %s, %s, %s, %s, NOW())
        ON CONFLICT (project_id, volunteer_id, notification_batch_id) DO NOTHING
    """
    notification_id = uuid.uuid4()
    cursor.execute(query, (notification_id, project_id, volunteer_id, match_score, batch_id))


def process_project_notifications(conn, batch_id: uuid.UUID):
    """Process notifications for all recruiting projects."""
    cursor = conn.cursor()
    
    print(f"Starting candidate notification batch at {datetime.now()}")
    print(f"Batch ID: {batch_id}")
    print(f"Configuration: TOP_K={TOP_K_CANDIDATES}, MIN_SCORE={MIN_MATCH_SCORE}")
    
    # Get all recruiting projects
    projects = get_recruiting_projects(cursor)
    print(f"Found {len(projects)} recruiting/active projects")
    
    total_candidates_notified = 0
    total_team_leads_notified = 0
    
    for project_id, project_title, team_lead_id, created_by_admin_id in projects:
        print(f"\n📋 Processing project: {project_title} ({project_id})")
        
        # Get top candidates for this project
        candidates = get_top_candidates(cursor, project_id, TOP_K_CANDIDATES, MIN_MATCH_SCORE, batch_id)
        
        if not candidates:
            print(f"  ℹ️  No new candidates found (threshold: {MIN_MATCH_SCORE})")
            continue
        
        print(f"  ✓ Found {len(candidates)} top candidates")
        
        # Notify each candidate
        for volunteer_id, user_id, match_score, volunteer_name in candidates:
            try:
                send_candidate_notification(cursor, project_id, project_title, user_id, match_score)
                record_notification(cursor, project_id, volunteer_id, match_score, batch_id)
                print(f"    → Notified {volunteer_name} ({int(match_score * 100)}% match)")
                total_candidates_notified += 1
            except Exception as e:
                print(f"    ❌ Failed to notify {volunteer_name}: {e}")
                continue
        
        # Notify team lead (or admin if no team lead)
        recipient_id = team_lead_id if team_lead_id else created_by_admin_id
        if recipient_id:
            try:
                send_team_lead_notification(cursor, project_id, project_title, recipient_id, candidates)
                print(f"  ✓ Notified team lead about {len(candidates)} candidates")
                total_team_leads_notified += 1
            except Exception as e:
                print(f"  ❌ Failed to notify team lead: {e}")
        
        conn.commit()
    
    print(f"\n{'='*60}")
    print(f"Batch completed at {datetime.now()}")
    print(f"Total candidates notified: {total_candidates_notified}")
    print(f"Total team leads notified: {total_team_leads_notified}")
    print(f"{'='*60}")


def get_notification_statistics(conn):
    """Get statistics about notifications sent."""
    cursor = conn.cursor()
    
    # Total notifications sent
    cursor.execute("SELECT COUNT(*) FROM candidate_notifications")
    total_notifications = cursor.fetchone()[0]
    
    # Notifications in last 24 hours
    cursor.execute("""
        SELECT COUNT(*) 
        FROM candidate_notifications 
        WHERE notified_at >= NOW() - INTERVAL '24 hours'
    """)
    recent_notifications = cursor.fetchone()[0]
    
    # Unique projects with notifications
    cursor.execute("SELECT COUNT(DISTINCT project_id) FROM candidate_notifications")
    projects_with_notifications = cursor.fetchone()[0]
    
    # Unique volunteers notified
    cursor.execute("SELECT COUNT(DISTINCT volunteer_id) FROM candidate_notifications")
    volunteers_notified = cursor.fetchone()[0]
    
    print(f"\n=== Notification Statistics ===")
    print(f"Total notifications sent (all time): {total_notifications}")
    print(f"Notifications sent (last 24h): {recent_notifications}")
    print(f"Projects with notifications: {projects_with_notifications}")
    print(f"Unique volunteers notified: {volunteers_notified}")


def main():
    """Main function to run the notification job."""
    try:
        print("=== Project Candidate Notification System ===")
        print(f"Started at {datetime.now()}")
        
        # Generate batch ID for this run
        batch_id = uuid.uuid4()
        
        # Connect to database
        conn = get_db_connection()
        print("Connected to database")
        
        # Process all projects
        process_project_notifications(conn, batch_id)
        
        # Show statistics (optional)
        try:
            get_notification_statistics(conn)
        except Exception as e:
            print(f"\nℹ️  Could not fetch statistics: {e}")
        
        print(f"\n✅ Job completed successfully at {datetime.now()}")
        
    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
    finally:
        if 'conn' in locals():
            conn.close()


if __name__ == "__main__":
    main()


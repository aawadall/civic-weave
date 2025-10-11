#!/usr/bin/env python3
"""
Hourly batch job to pre-calculate volunteer-initiative match scores.
Populates the volunteer_initiative_matches table for fast lookups.
"""

import os
import sys
import psycopg2
import numpy as np
from datetime import datetime
from typing import Dict, List, Tuple

# Add the backend directory to the path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

def get_db_connection():
    """Get database connection from environment variables."""
    return psycopg2.connect(
        host=os.getenv('DB_HOST', 'localhost'),
        port=os.getenv('DB_PORT', '5432'),
        database=os.getenv('DB_NAME', 'civicweave'),
        user=os.getenv('DB_USER', 'postgres'),
        password=os.getenv('DB_PASSWORD', 'password')
    )

def calculate_cosine_similarity(volunteer_weights: Dict[int, float], project_skill_ids: List[int]) -> Tuple[float, List[int], int]:
    """
    Calculate cosine similarity in restricted space (intersection only).
    
    Args:
        volunteer_weights: {skill_id: weight} mapping for volunteer
        project_skill_ids: List of required skill IDs for project
        
    Returns:
        Tuple of (cosine_score, matched_skill_ids, matched_count)
    """
    # Find intersection
    matched_ids = [sid for sid in project_skill_ids if sid in volunteer_weights]
    
    if not matched_ids:
        return 0.0, [], 0
    
    # Build vectors in intersection space
    v_vec = np.array([volunteer_weights[sid] for sid in matched_ids])
    p_vec = np.ones(len(matched_ids))  # project requirements = all 1s
    
    # Cosine similarity
    dot_product = np.dot(v_vec, p_vec)
    norm_v = np.linalg.norm(v_vec)
    norm_p = np.linalg.norm(p_vec)
    
    cosine = dot_product / (norm_v * norm_p) if norm_v > 0 and norm_p > 0 else 0.0
    
    return cosine, matched_ids, len(matched_ids)

def calculate_euclidean_similarity(volunteer_weights: Dict[int, float], project_skill_ids: List[int]) -> float:
    """
    Calculate similarity based on Euclidean distance.
    
    Args:
        volunteer_weights: {skill_id: weight} mapping for volunteer
        project_skill_ids: List of required skill IDs for project
        
    Returns:
        Euclidean similarity score (0-1)
    """
    if not project_skill_ids:
        return 0.0
    
    # Build volunteer vector in project space
    v_vec = np.array([volunteer_weights.get(sid, 0.0) for sid in project_skill_ids])
    p_vec = np.ones(len(project_skill_ids))  # project expects weight 1.0 for all skills
    
    # Calculate distance from ideal
    euclidean_dist = np.linalg.norm(v_vec - p_vec)
    
    # Normalize to [0,1] range
    max_possible_dist = np.sqrt(len(project_skill_ids))  # max distance when all weights are 0
    normalized_dist = min(euclidean_dist / max_possible_dist, 1.0)
    
    return 1.0 - normalized_dist

def calculate_coverage_score(volunteer_weights: Dict[int, float], project_skill_ids: List[int]) -> float:
    """
    Calculate simple weighted coverage score.
    
    Args:
        volunteer_weights: {skill_id: weight} mapping for volunteer
        project_skill_ids: List of required skill IDs for project
        
    Returns:
        Coverage score (0-1)
    """
    if not project_skill_ids:
        return 0.0
    
    total_weight = sum(volunteer_weights.get(sid, 0.0) for sid in project_skill_ids)
    return total_weight / len(project_skill_ids)

def recalculate_all_matches(conn):
    """Recalculate all volunteer-initiative matches."""
    cursor = conn.cursor()
    
    print(f"Starting match calculation at {datetime.now()}")
    
    # Get all active initiatives with their required skills
    cursor.execute("""
        SELECT i.id, array_agg(irs.skill_id) as required_skills
        FROM initiatives i
        JOIN initiative_required_skills irs ON i.id = irs.initiative_id
        WHERE i.status = 'active'
        GROUP BY i.id
    """)
    initiatives = cursor.fetchall()
    print(f"Found {len(initiatives)} active initiatives")
    
    # Get all volunteers with their skill weights
    cursor.execute("""
        SELECT v.id, vs.skill_id, vs.skill_weight
        FROM volunteers v
        JOIN volunteer_skills vs ON v.id = vs.volunteer_id
        ORDER BY v.id, vs.skill_id
    """)
    
    # Build volunteer skill dict: {volunteer_id: {skill_id: weight}}
    volunteer_skills = {}
    for volunteer_id, skill_id, weight in cursor.fetchall():
        if volunteer_id not in volunteer_skills:
            volunteer_skills[volunteer_id] = {}
        volunteer_skills[volunteer_id][skill_id] = float(weight)
    
    print(f"Found {len(volunteer_skills)} volunteers with skills")
    
    # Calculate matches for all combinations
    matches = []
    total_combinations = len(initiatives) * len(volunteer_skills)
    processed = 0
    
    for initiative_id, required_skills in initiatives:
        for volunteer_id, skills in volunteer_skills.items():
            # Calculate all similarity metrics
            cosine_score, matched_ids, matched_count = calculate_cosine_similarity(
                skills, required_skills
            )
            
            if matched_count > 0:  # Only store if at least 1 skill matches
                euclidean_score = calculate_euclidean_similarity(skills, required_skills)
                coverage_score = calculate_coverage_score(skills, required_skills)
                
                # Use cosine as primary match score
                jaccard = matched_count / len(required_skills)
                
                matches.append((
                    volunteer_id, initiative_id, cosine_score, jaccard,
                    matched_ids, matched_count, datetime.now()
                ))
            
            processed += 1
            if processed % 1000 == 0:
                print(f"Processed {processed}/{total_combinations} combinations...")
    
    print(f"Calculated {len(matches)} matches")
    
    # Clear old matches and insert new ones
    print("Clearing old matches...")
    cursor.execute("TRUNCATE volunteer_initiative_matches")
    
    print("Inserting new matches...")
    cursor.executemany("""
        INSERT INTO volunteer_initiative_matches 
        (volunteer_id, initiative_id, match_score, jaccard_index, 
         matched_skill_ids, matched_skill_count, calculated_at)
        VALUES (%s, %s, %s, %s, %s, %s, %s)
    """, matches)
    
    conn.commit()
    print(f"Successfully stored {len(matches)} matches at {datetime.now()}")

def get_match_statistics(conn):
    """Get statistics about the calculated matches."""
    cursor = conn.cursor()
    
    # Basic counts
    cursor.execute("SELECT COUNT(*) FROM volunteer_initiative_matches")
    total_matches = cursor.fetchone()[0]
    
    cursor.execute("SELECT COUNT(DISTINCT volunteer_id) FROM volunteer_initiative_matches")
    volunteers_with_matches = cursor.fetchone()[0]
    
    cursor.execute("SELECT COUNT(DISTINCT initiative_id) FROM volunteer_initiative_matches")
    initiatives_with_matches = cursor.fetchone()[0]
    
    # Score distribution
    cursor.execute("""
        SELECT 
            COUNT(*) as count,
            AVG(match_score) as avg_score,
            MIN(match_score) as min_score,
            MAX(match_score) as max_score
        FROM volunteer_initiative_matches
    """)
    stats = cursor.fetchone()
    
    print(f"\n=== Match Statistics ===")
    print(f"Total matches: {total_matches}")
    print(f"Volunteers with matches: {volunteers_with_matches}")
    print(f"Initiatives with matches: {initiatives_with_matches}")
    print(f"Average match score: {stats[1]:.3f}")
    print(f"Score range: {stats[2]:.3f} - {stats[3]:.3f}")

def main():
    """Main function to run the match calculation job."""
    try:
        print("=== Volunteer-Initiative Match Calculator ===")
        print(f"Started at {datetime.now()}")
        
        # Connect to database
        conn = get_db_connection()
        print("Connected to database")
        
        # Run the calculation
        recalculate_all_matches(conn)
        
        # Show statistics
        get_match_statistics(conn)
        
        print(f"Job completed successfully at {datetime.now()}")
        
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)
    finally:
        if 'conn' in locals():
            conn.close()

if __name__ == "__main__":
    main()

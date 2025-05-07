-- Materialized Views for MamaCare SL
-- Pre-computed views for reporting and analytics

-- Monthly visit statistics by district
CREATE MATERIALIZED VIEW mv_monthly_visit_stats AS
SELECT
  date_trunc('month', COALESCE(completed_date, scheduled_date)) AS month,
  hf.district,
  v.visit_type,
  v.status,
  COUNT(*) AS visit_count,
  SUM(CASE WHEN status = 'COMPLETED' THEN 1 ELSE 0 END) AS completed_count,
  SUM(CASE WHEN status = 'MISSED' THEN 1 ELSE 0 END) AS missed_count,
  CASE 
    WHEN COUNT(*) > 0 THEN 
      ROUND((SUM(CASE WHEN status = 'COMPLETED' THEN 1 ELSE 0 END)::numeric / COUNT(*)::numeric) * 100, 2)
    ELSE 0
  END AS completion_rate
FROM visits v
LEFT JOIN healthcare_facilities hf ON v.facility_id = hf.id
GROUP BY 
  date_trunc('month', COALESCE(completed_date, scheduled_date)),
  hf.district,
  v.visit_type,
  v.status
WITH NO DATA;

-- Create index for the materialized view
CREATE UNIQUE INDEX idx_mv_monthly_visit_stats ON mv_monthly_visit_stats (month, district, visit_type, status);

-- Child vaccination coverage by age group
CREATE MATERIALIZED VIEW mv_vaccination_coverage AS
SELECT
  date_trunc('month', CURRENT_DATE) AS report_month,
  hf.district,
  ir.vaccine_name,
  ir.vaccine_dose,
  COUNT(DISTINCT c.id) AS children_vaccinated,
  (
    SELECT COUNT(*) 
    FROM children c2
    WHERE 
      EXTRACT(YEAR FROM age(CURRENT_DATE, c2.date_of_birth)) * 12 +
      EXTRACT(MONTH FROM age(CURRENT_DATE, c2.date_of_birth)) 
      BETWEEN 
        (SELECT age_months FROM vaccine_schedules vs WHERE vs.vaccine_name = ir.vaccine_name AND vs.vaccine_dose = ir.vaccine_dose) - 1
      AND 
        (SELECT age_months FROM vaccine_schedules vs WHERE vs.vaccine_name = ir.vaccine_name AND vs.vaccine_dose = ir.vaccine_dose) + 3
  ) AS eligible_children,
  CASE 
    WHEN (
      SELECT COUNT(*) 
      FROM children c2
      WHERE 
        EXTRACT(YEAR FROM age(CURRENT_DATE, c2.date_of_birth)) * 12 +
        EXTRACT(MONTH FROM age(CURRENT_DATE, c2.date_of_birth)) 
        BETWEEN 
          (SELECT age_months FROM vaccine_schedules vs WHERE vs.vaccine_name = ir.vaccine_name AND vs.vaccine_dose = ir.vaccine_dose) - 1
        AND 
          (SELECT age_months FROM vaccine_schedules vs WHERE vs.vaccine_name = ir.vaccine_name AND vs.vaccine_dose = ir.vaccine_dose) + 3
    ) > 0 THEN
      ROUND(
        (COUNT(DISTINCT c.id)::numeric / 
        (
          SELECT COUNT(*) 
          FROM children c2
          WHERE 
            EXTRACT(YEAR FROM age(CURRENT_DATE, c2.date_of_birth)) * 12 +
            EXTRACT(MONTH FROM age(CURRENT_DATE, c2.date_of_birth)) 
            BETWEEN 
              (SELECT age_months FROM vaccine_schedules vs WHERE vs.vaccine_name = ir.vaccine_name AND vs.vaccine_dose = ir.vaccine_dose) - 1
            AND 
              (SELECT age_months FROM vaccine_schedules vs WHERE vs.vaccine_name = ir.vaccine_name AND vs.vaccine_dose = ir.vaccine_dose) + 3
        )::numeric) * 100, 2
      )
    ELSE 0
  END AS coverage_percentage
FROM 
  immunization_records ir
  JOIN children c ON ir.child_id = c.id
  LEFT JOIN healthcare_facilities hf ON ir.administered_at_facility_id = hf.id
  JOIN vaccine_schedules vs ON ir.vaccine_name = vs.vaccine_name AND ir.vaccine_dose = vs.vaccine_dose
GROUP BY
  date_trunc('month', CURRENT_DATE),
  hf.district,
  ir.vaccine_name,
  ir.vaccine_dose
WITH NO DATA;

-- Create index for the materialized view
CREATE UNIQUE INDEX idx_mv_vaccination_coverage ON mv_vaccination_coverage (report_month, district, vaccine_name, vaccine_dose);

-- Risk assessment summary view
CREATE MATERIALIZED VIEW mv_risk_assessment_summary AS
SELECT
  date_trunc('month', sr.screened_at) AS month,
  hf.district,
  sr.risk_level,
  COUNT(*) AS assessment_count,
  SUM(CASE WHEN sr.danger_signs_detected THEN 1 ELSE 0 END) AS danger_signs_count,
  COUNT(DISTINCT sr.mother_id) AS unique_mothers_assessed,
  COUNT(DISTINCT sr.child_id) AS unique_children_assessed,
  ROUND(AVG(EXTRACT(EPOCH FROM (COALESCE(se.accepted_at, CURRENT_TIMESTAMP) - sr.screened_at)) / 60), 2) AS avg_response_time_minutes
FROM 
  screener_results sr
  LEFT JOIN healthcare_facilities hf ON sr.facility_id = hf.id
  LEFT JOIN sos_events se ON se.user_id = sr.mother_id 
    AND se.created_at >= sr.screened_at 
    AND se.created_at <= sr.screened_at + interval '24 hours'
    AND sr.danger_signs_detected = true
GROUP BY
  date_trunc('month', sr.screened_at),
  hf.district,
  sr.risk_level
WITH NO DATA;

-- Create index for the materialized view
CREATE UNIQUE INDEX idx_mv_risk_assessment_summary ON mv_risk_assessment_summary (month, district, risk_level);

-- Child growth statistics view
CREATE MATERIALIZED VIEW mv_child_growth_stats AS
SELECT
  date_trunc('month', gm.measured_at) AS month,
  hf.district,
  EXTRACT(YEAR FROM age(gm.measured_at, c.date_of_birth)) * 12 +
  EXTRACT(MONTH FROM age(gm.measured_at, c.date_of_birth)) AS age_months,
  gm.growth_status,
  COUNT(*) AS measurement_count,
  ROUND(AVG(gm.weight_grams)::numeric / 1000, 2) AS avg_weight_kg,
  ROUND(AVG(gm.height_cm), 1) AS avg_height_cm,
  ROUND(AVG(gm.weight_for_age_z), 2) AS avg_weight_for_age_z,
  ROUND(AVG(gm.height_for_age_z), 2) AS avg_height_for_age_z
FROM 
  growth_measurements gm
  JOIN children c ON gm.child_id = c.id
  LEFT JOIN healthcare_facilities hf ON gm.measured_at_facility_id = hf.id
GROUP BY
  date_trunc('month', gm.measured_at),
  hf.district,
  EXTRACT(YEAR FROM age(gm.measured_at, c.date_of_birth)) * 12 +
  EXTRACT(MONTH FROM age(gm.measured_at, c.date_of_birth)),
  gm.growth_status
WITH NO DATA;

-- Create index for the materialized view
CREATE UNIQUE INDEX idx_mv_child_growth_stats ON mv_child_growth_stats (month, district, age_months, growth_status);

-- Function to refresh all materialized views
CREATE OR REPLACE FUNCTION refresh_all_materialized_views()
RETURNS void AS $$
BEGIN
  REFRESH MATERIALIZED VIEW CONCURRENTLY mv_monthly_visit_stats;
  REFRESH MATERIALIZED VIEW CONCURRENTLY mv_vaccination_coverage;
  REFRESH MATERIALIZED VIEW CONCURRENTLY mv_risk_assessment_summary;
  REFRESH MATERIALIZED VIEW CONCURRENTLY mv_child_growth_stats;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON MATERIALIZED VIEW mv_monthly_visit_stats IS 'Monthly statistics on maternal and child health visits by district';
COMMENT ON MATERIALIZED VIEW mv_vaccination_coverage IS 'Vaccination coverage statistics by district and vaccine type';
COMMENT ON MATERIALIZED VIEW mv_risk_assessment_summary IS 'Summary of risk assessments and response times for danger signs';
COMMENT ON MATERIALIZED VIEW mv_child_growth_stats IS 'Child growth statistics by age group, district, and growth status';
COMMENT ON FUNCTION refresh_all_materialized_views IS 'Function to refresh all materialized views concurrently, typically run via a scheduled job';

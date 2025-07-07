-- Screener questions table for MamaCare SL
-- Defines health screening questions used in assessments

CREATE TABLE IF NOT EXISTS screener_questions (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Question content
  question_text non_empty_text NOT NULL,
  question_code TEXT NOT NULL, -- Machine-readable code for this question
  
  -- Question type and format
  answer_type TEXT NOT NULL CHECK (
    answer_type IN ('BOOLEAN', 'MULTIPLE_CHOICE', 'NUMERIC', 'TEXT')
  ),
  
  -- For multiple choice questions
  answer_options JSONB, -- Array of possible answers
  
  -- Risk assessment
  is_danger_sign BOOLEAN NOT NULL DEFAULT FALSE, -- Whether this indicates a danger sign
  risk_category risk_level, -- What risk level a positive answer indicates
  
  -- Question categorization
  category TEXT NOT NULL, -- e.g., "Maternal", "Newborn", "General"
  subcategory TEXT, -- e.g., "Bleeding", "Fever", "Nutrition"
  
  -- Ordering
  display_order INTEGER NOT NULL,
  
  -- Conditional logic
  depends_on_question_id UUID REFERENCES screener_questions(id),
  depends_on_answer TEXT, -- Answer to the dependent question that triggers this question
  
  -- Translation keys for multilingual support
  translation_key TEXT NOT NULL,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  
  -- Make sure MULTIPLE_CHOICE questions have answer options
  CONSTRAINT valid_multiple_choice_options CHECK (
    answer_type != 'MULTIPLE_CHOICE' OR answer_options IS NOT NULL
  ),
  
  -- Ensure unique question codes
  UNIQUE (question_code)
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_screener_questions_updated_at
BEFORE UPDATE ON screener_questions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- No row-level security since this is reference data
-- This should be managed by admins only via Hasura permissions

-- Create indexes for common queries
CREATE INDEX idx_screener_questions_category ON screener_questions (category);
CREATE INDEX idx_screener_questions_is_danger_sign ON screener_questions (is_danger_sign);
CREATE INDEX idx_screener_questions_risk_category ON screener_questions (risk_category);
CREATE INDEX idx_screener_questions_active ON screener_questions (is_active);

-- Add comments for documentation
COMMENT ON TABLE screener_questions IS 'Defines health screening questions used in maternal and child health assessments';
COMMENT ON COLUMN screener_questions.question_code IS 'Unique code for programmatic reference';
COMMENT ON COLUMN screener_questions.is_danger_sign IS 'Whether a positive response indicates a potential emergency';
COMMENT ON COLUMN screener_questions.answer_options IS 'JSON array of options for multiple choice questions';

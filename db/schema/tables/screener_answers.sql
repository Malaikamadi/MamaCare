-- Screener answers table for MamaCare SL
-- Stores individual responses to screening questions

CREATE TABLE IF NOT EXISTS screener_answers (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Relationship to screening session
  screener_result_id UUID NOT NULL REFERENCES screener_results(id) ON DELETE CASCADE,
  
  -- Question reference
  question_id UUID NOT NULL REFERENCES screener_questions(id),
  question_text TEXT NOT NULL, -- Denormalized for historical record
  
  -- Answer data (stored in appropriate type column based on question type)
  answer_boolean BOOLEAN,
  answer_text TEXT,
  answer_numeric DECIMAL,
  answer_option TEXT, -- Selected option for multiple choice
  
  -- Risk evaluation for this specific answer
  contributed_to_risk BOOLEAN NOT NULL DEFAULT FALSE,
  individual_risk_level risk_level,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  answer_sequence INTEGER NOT NULL, -- Order in which questions were answered
  
  -- Ensure answer is stored in appropriate column based on question type
  CONSTRAINT valid_answer_type CHECK (
    (answer_boolean IS NOT NULL AND answer_text IS NULL AND answer_numeric IS NULL AND answer_option IS NULL) OR
    (answer_boolean IS NULL AND answer_text IS NOT NULL AND answer_numeric IS NULL AND answer_option IS NULL) OR
    (answer_boolean IS NULL AND answer_text IS NULL AND answer_numeric IS NOT NULL AND answer_option IS NULL) OR
    (answer_boolean IS NULL AND answer_text IS NULL AND answer_numeric IS NULL AND answer_option IS NOT NULL)
  ),
  
  -- Ensure unique question per screening session
  UNIQUE (screener_result_id, question_id)
);

-- No update trigger needed as answers shouldn't change after recording
-- (maintaining data integrity for medical records)

-- Row-level security policies for Hasura
ALTER TABLE screener_answers ENABLE ROW LEVEL SECURITY;

-- Answers inherit access permissions from parent screener_results
CREATE POLICY inherit_screener_results_permissions ON screener_answers
  USING (
    EXISTS (
      SELECT 1 FROM screener_results sr
      WHERE sr.id = screener_answers.screener_result_id
    )
  );

-- Create indexes for common queries
CREATE INDEX idx_screener_answers_result_id ON screener_answers (screener_result_id);
CREATE INDEX idx_screener_answers_question_id ON screener_answers (question_id);
CREATE INDEX idx_screener_answers_contributed_risk ON screener_answers (contributed_to_risk);
CREATE INDEX idx_screener_answers_sequence ON screener_answers (screener_result_id, answer_sequence);

-- Add comments for documentation
COMMENT ON TABLE screener_answers IS 'Stores individual responses to screening questions';
COMMENT ON COLUMN screener_answers.question_text IS 'Denormalized question text for historical record in case question changes';
COMMENT ON COLUMN screener_answers.contributed_to_risk IS 'Whether this answer contributed to overall risk assessment';
COMMENT ON COLUMN screener_answers.answer_sequence IS 'Order in which questions were answered in the screening';

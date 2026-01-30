-- Drop existing table if it exists (to ensure clean migration)
DROP TABLE IF EXISTS contact_submissions CASCADE;

-- Create contact_submissions table for storing contact form submissions
CREATE TABLE contact_submissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  subject TEXT NOT NULL,
  message TEXT NOT NULL,
  ip_address TEXT,
  user_agent TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  read_at TIMESTAMPTZ,
  replied_at TIMESTAMPTZ,
  status TEXT DEFAULT 'new',
  CONSTRAINT valid_status CHECK (status IN ('new', 'read', 'replied', 'archived'))
);

-- Create indexes for better query performance
CREATE INDEX idx_contact_submissions_email ON contact_submissions(email);
CREATE INDEX idx_contact_submissions_created_at ON contact_submissions(created_at DESC);
CREATE INDEX idx_contact_submissions_status ON contact_submissions(status);

-- Enable Row Level Security
ALTER TABLE contact_submissions ENABLE ROW LEVEL SECURITY;

-- Policy: Allow public to insert (submit contact form)
CREATE POLICY "Allow public insert on contact_submissions"
  ON contact_submissions
  FOR INSERT
  WITH CHECK (true);

-- Policy: Only authenticated users can view submissions
CREATE POLICY "Only authenticated users can view contact_submissions"
  ON contact_submissions
  FOR SELECT
  TO authenticated
  USING (true);

-- Policy: Only authenticated users can update submissions
CREATE POLICY "Only authenticated users can update contact_submissions"
  ON contact_submissions
  FOR UPDATE
  TO authenticated
  USING (true);

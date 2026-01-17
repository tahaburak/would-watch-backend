#!/bin/bash
# Helper script to get the correct Supabase database connection string
# Usage: ./scripts/get-db-url.sh

set -e

PROJECT_REF="gtjokreqhfsydfmtbtvg"
REGION="eu-west-1"

echo "To get your database connection string:"
echo ""
echo "1. Go to your Supabase Dashboard: https://app.supabase.com/project/${PROJECT_REF}"
echo "2. Navigate to: Settings > Database"
echo "3. Under 'Connection string', select 'Transaction pooler'"
echo "4. Copy the connection string and replace [YOUR-PASSWORD] with your database password"
echo ""
echo "The correct format should be:"
echo "postgresql://postgres.${PROJECT_REF}:[YOUR-PASSWORD]@aws-1-${REGION}.pooler.supabase.com:6543/postgres"
echo ""
echo "IMPORTANT: Make sure the username is 'postgres.${PROJECT_REF}' (not 'app_user.${PROJECT_REF}')"
echo ""
echo "If you're using the direct connection (port 5432), the format is:"
echo "postgresql://postgres:[YOUR-PASSWORD]@db.${PROJECT_REF}.supabase.co:5432/postgres"
echo ""
echo "Note: Direct connection doesn't include project-ref in the username."

-- Create multiple databases for microservices
SELECT 'CREATE DATABASE userdb'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'userdb')\gexec

SELECT 'CREATE DATABASE videodb'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'videodb')\gexec

SELECT 'CREATE DATABASE eventdb'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'eventdb')\gexec

SELECT 'CREATE DATABASE interactiondb'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'interactiondb')\gexec

SELECT 'CREATE DATABASE notificationdb'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'notificationdb')\gexec

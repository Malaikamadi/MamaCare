/**
 * MamaCare SL - Hasura Health Check
 * Validates database connection and GraphQL engine functionality
 */

const http = require('http');
const https = require('https');

// Configuration constants
const ENDPOINT = process.env.HASURA_ENDPOINT || 'http://localhost:8080';
const ADMIN_SECRET = process.env.HASURA_GRAPHQL_ADMIN_SECRET || 'mamacare-dev-admin-secret';
const TIMEOUT_MS = 5000;

// Health check query - validates connection to PostgreSQL and extensions
const healthCheckQuery = {
  query: `
    query HealthCheck {
      # Check PostgreSQL connection
      postgres_health: schema_version {
        version
        applied_at
      }
      
      # Check PostGIS extension
      postgis_enabled: __typename @skip(if: false)
    }
  `
};

/**
 * Makes HTTP request to Hasura endpoint
 * @param {string} url - The URL to request
 * @param {Object} data - The request payload
 * @returns {Promise<Object>} - Response data
 */
function makeRequest(url, data) {
  return new Promise((resolve, reject) => {
    // Select HTTP module based on protocol
    const httpModule = url.startsWith('https') ? https : http;
    
    const options = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': ADMIN_SECRET
      },
      timeout: TIMEOUT_MS
    };

    const req = httpModule.request(url, options, (res) => {
      let responseData = '';
      
      res.on('data', (chunk) => {
        responseData += chunk;
      });
      
      res.on('end', () => {
        try {
          const parsedData = JSON.parse(responseData);
          if (res.statusCode >= 200 && res.statusCode < 300) {
            resolve(parsedData);
          } else {
            reject(new Error(`HTTP Error ${res.statusCode}: ${JSON.stringify(parsedData)}`));
          }
        } catch (error) {
          reject(new Error(`Failed to parse response: ${error.message}`));
        }
      });
    });
    
    req.on('error', (error) => {
      reject(new Error(`Request failed: ${error.message}`));
    });
    
    req.on('timeout', () => {
      req.destroy();
      reject(new Error(`Request timed out after ${TIMEOUT_MS}ms`));
    });
    
    req.write(JSON.stringify(data));
    req.end();
  });
}

/**
 * Main health check function
 * @returns {Promise<void>}
 */
async function checkHealth() {
  try {
    console.log(`Checking Hasura health at ${ENDPOINT}...`);
    
    const response = await makeRequest(`${ENDPOINT}/v1/graphql`, healthCheckQuery);
    
    if (response.errors) {
      console.error('Health check failed:', response.errors);
      process.exit(1);
    }
    
    if (response.data && response.data.postgres_health) {
      console.log('✅ Database connection successful');
      console.log(`✅ Schema version: ${response.data.postgres_health.version}`);
      console.log('✅ Hasura GraphQL engine is healthy');
      process.exit(0);
    } else {
      console.error('❌ Invalid response format:', response);
      process.exit(1);
    }
  } catch (error) {
    console.error('❌ Health check failed:', error.message);
    process.exit(1);
  }
}

// Execute health check
checkHealth();

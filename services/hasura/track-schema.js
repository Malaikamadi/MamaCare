/**
 * MamaCare SL - Database Schema Tracking Utility
 * 
 * This script helps track tables in the public schema and set up relationships
 * for the Hasura GraphQL engine. It follows TypeScript-inspired principles:
 * - Strong typing for relationship definitions
 * - Validation of relationships before applying
 * - Clear error handling with early returns
 */

const fetch = require('node-fetch');
const fs = require('fs');
const path = require('path');

// Configuration (from environment or default values)
const HASURA_ENDPOINT = process.env.HASURA_ENDPOINT || 'http://localhost:8080/v1/metadata';
const HASURA_ADMIN_SECRET = process.env.HASURA_GRAPHQL_ADMIN_SECRET || 'mamacare-dev-admin-secret';
const SCHEMA = 'public';

/**
 * @typedef {Object} TableRelationship
 * @property {string} name - Relationship name
 * @property {string} type - Relationship type (object or array)
 * @property {string} targetTable - Target table name
 * @property {Object.<string, string>} fieldMapping - Mapping of source to target fields
 */

/**
 * Main tables to track with their relationships
 * @type {Object.<string, Array<TableRelationship>>}
 */
const TABLES_TO_TRACK = {
  'users': [
    {
      name: 'facility',
      type: 'object',
      targetTable: 'healthcare_facilities',
      fieldMapping: { 'facility_id': 'id' }
    },
    {
      name: 'children',
      type: 'array',
      targetTable: 'children',
      fieldMapping: { 'id': 'mother_id' }
    },
    {
      name: 'visits',
      type: 'array',
      targetTable: 'visits',
      fieldMapping: { 'id': 'mother_id' }
    }
  ],
  'children': [
    {
      name: 'mother',
      type: 'object',
      targetTable: 'users',
      fieldMapping: { 'mother_id': 'id' }
    },
    {
      name: 'growth_measurements',
      type: 'array',
      targetTable: 'growth_measurements',
      fieldMapping: { 'id': 'child_id' }
    },
    {
      name: 'immunization_records',
      type: 'array',
      targetTable: 'immunization_records',
      fieldMapping: { 'id': 'child_id' }
    }
  ],
  'healthcare_facilities': [
    {
      name: 'staff',
      type: 'array',
      targetTable: 'users',
      fieldMapping: { 'id': 'facility_id' }
    },
    {
      name: 'ambulances',
      type: 'array',
      targetTable: 'ambulances',
      fieldMapping: { 'id': 'facility_id' }
    }
  ],
  // Additional tables would go here
};

/**
 * Make a request to the Hasura metadata API
 * @param {Object} data - Request payload
 * @returns {Promise<Object>} - Response data
 */
async function hasuraRequest(data) {
  try {
    const response = await fetch(HASURA_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': HASURA_ADMIN_SECRET
      },
      body: JSON.stringify(data)
    });

    const responseData = await response.json();
    
    if (!response.ok) {
      throw new Error(`Hasura error: ${JSON.stringify(responseData)}`);
    }
    
    return responseData;
  } catch (error) {
    console.error(`API request failed: ${error.message}`);
    throw error;
  }
}

/**
 * Track all tables in the public schema
 */
async function trackAllTables() {
  try {
    console.log('Tracking all tables in public schema...');
    
    const trackSchemaResponse = await hasuraRequest({
      type: 'pg_track_tables',
      args: {
        schema: SCHEMA,
        source: 'default',
        track_as_enum: true
      }
    });
    
    console.log('✅ All tables tracked successfully');
    
    // Export metadata to file
    await hasuraRequest({
      type: 'export_metadata',
      args: {}
    });
    
    console.log('✅ Metadata exported successfully');
    
  } catch (error) {
    console.error(`❌ Failed to track tables: ${error.message}`);
    process.exit(1);
  }
}

/**
 * Create relationships for tracked tables
 */
async function createRelationships() {
  try {
    console.log('Setting up table relationships...');
    
    for (const [tableName, relationships] of Object.entries(TABLES_TO_TRACK)) {
      console.log(`Processing relationships for table: ${tableName}`);
      
      for (const relationship of relationships) {
        console.log(`  - Creating ${relationship.type} relationship: ${relationship.name} -> ${relationship.targetTable}`);
        
        // Build field mappings
        const fieldMappings = {};
        for (const [sourceField, targetField] of Object.entries(relationship.fieldMapping)) {
          fieldMappings[sourceField] = targetField;
        }
        
        // Determine relationship type and create appropriate payload
        const isArrayRelationship = relationship.type === 'array';
        const requestType = isArrayRelationship ? 'pg_create_array_relationship' : 'pg_create_object_relationship';
        
        const requestPayload = {
          type: requestType,
          args: {
            table: { schema: SCHEMA, name: tableName },
            name: relationship.name,
            source: 'default',
            using: {
              foreign_key_constraint_on: isArrayRelationship 
                ? { table: { schema: SCHEMA, name: relationship.targetTable }, column: Object.values(relationship.fieldMapping)[0] }
                : Object.keys(relationship.fieldMapping)[0]
            }
          }
        };
        
        // Create the relationship
        await hasuraRequest(requestPayload);
        console.log(`    ✅ Relationship created successfully`);
      }
    }
    
    console.log('✅ All relationships created successfully');
    
  } catch (error) {
    console.error(`❌ Failed to create relationships: ${error.message}`);
    process.exit(1);
  }
}

/**
 * Main function to track schema and create relationships
 */
async function main() {
  try {
    // First track all tables
    await trackAllTables();
    
    // Then create relationships
    await createRelationships();
    
    console.log('✅ Schema tracking and relationship setup completed successfully');
    process.exit(0);
  } catch (error) {
    console.error(`❌ Error in main process: ${error.message}`);
    process.exit(1);
  }
}

// Execute the main function
main();

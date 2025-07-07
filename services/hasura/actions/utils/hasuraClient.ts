import { GraphQLClient } from 'graphql-request';

/**
 * Creates a configured GraphQL client for Hasura
 * Uses proper authentication and error handling
 */
export const hasuraGraphqlClient = new GraphQLClient(
  process.env.HASURA_GRAPHQL_ENDPOINT || 'http://localhost:8080/v1/graphql',
  {
    headers: {
      'Content-Type': 'application/json',
      'x-hasura-admin-secret': process.env.HASURA_GRAPHQL_ADMIN_SECRET || ''
    },
  }
);

/**
 * Executes a GraphQL query with proper error handling and logging
 * 
 * @param query The GraphQL query to execute
 * @param variables Variables for the query
 * @returns The query result data
 */
export async function executeQuery<T = any>(query: string, variables?: Record<string, any>): Promise<T> {
  try {
    return await hasuraGraphqlClient.request<T>(query, variables);
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error(`GraphQL query failed: ${errorMessage}`, {
      query,
      variables: JSON.stringify(variables)
    });
    throw new Error(`GraphQL query failed: ${errorMessage}`);
  }
}

/**
 * Builds a where clause for filtering based on user role and context
 * Ensures proper row-level security when needed
 * 
 * @param userRole The role of the current user
 * @param userId The ID of the current user
 * @param filterParams Additional filter parameters
 * @returns A constructed where clause object for GraphQL queries
 */
export function buildWhereClause(
  userRole: string, 
  userId?: string,
  filterParams?: Record<string, any>
): Record<string, any> {
  const baseFilter = filterParams || {};
  
  // Apply role-based row restrictions
  switch (userRole) {
    case 'mother':
      // Mothers can only see their own data
      return {
        ...baseFilter,
        user_id: { _eq: userId }
      };
    case 'chw':
      // CHWs can see their assigned mothers' data
      return {
        ...baseFilter,
        _or: [
          { chw_id: { _eq: userId } },
          { assigned_chw_id: { _eq: userId } }
        ]
      };
    case 'clinician':
      // Clinicians can see data related to their facility
      return {
        ...baseFilter,
        facility_id: { _eq: userId }
      };
    case 'admin':
      // Admins can see all data
      return baseFilter;
    default:
      // Default to only showing the user's own data
      return {
        ...baseFilter,
        user_id: { _eq: userId }
      };
  }
}

/**
 * Validate that the current user has permission to access a resource
 * Throws an error if validation fails
 * 
 * @param resourceOwnerId The ID of the resource owner
 * @param currentUserId The ID of the current user
 * @param userRole The role of the current user
 * @param additionalCheck Optional additional check function
 */
export function validateResourceAccess(
  resourceOwnerId: string,
  currentUserId: string | undefined,
  userRole: string,
  additionalCheck?: () => Promise<boolean>
): void | never {
  // Admins always have access
  if (userRole === 'admin') {
    return;
  }
  
  // Users can access their own resources
  if (currentUserId && resourceOwnerId === currentUserId) {
    return;
  }
  
  // Throw error if no user ID available (except for admins)
  if (!currentUserId) {
    throw new Error('Unauthorized: User ID not available');
  }
  
  // For other roles, the caller must provide an additional check function
  if (!additionalCheck) {
    throw new Error('Unauthorized: Cannot access this resource');
  }
}

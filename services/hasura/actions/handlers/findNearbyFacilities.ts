import { 
  ActionRequest, 
  ActionResponse, 
  FindNearbyFacilitiesInput, 
  NearbyFacilityOutput,
  Geographic
} from '../types';
import * as z from 'zod';
import { hasuraGraphqlClient } from '../utils/hasuraClient';

// Validation schema using Zod
const inputSchema = z.object({
  location: z.object({
    latitude: z.number().min(-90).max(90),
    longitude: z.number().min(-180).max(180)
  }),
  radius: z.number().positive().max(100), // kilometers, with reasonable max
  facilityType: z.string().optional(),
  limit: z.number().int().positive().max(50).optional()
});

// Constants to avoid magic numbers/strings
const DEFAULT_LIMIT = 10;

/**
 * Find healthcare facilities within a specified radius of a given location
 * Uses PostGIS for efficient geospatial queries
 * 
 * @param req The action request containing location, radius and filters
 * @returns Array of nearby facilities with distance information
 */
export default async function findNearbyFacilities(
  req: ActionRequest<FindNearbyFacilitiesInput>
): Promise<ActionResponse<NearbyFacilityOutput[]>> {
  try {
    // Validate input with Zod schema
    const validatedInput = inputSchema.parse(req.input);
    const { location, radius, facilityType, limit = DEFAULT_LIMIT } = validatedInput;
    
    // Convert lat/lng to PostGIS point format
    const point = `POINT(${location.longitude} ${location.latitude})`;
    
    // Prepare variables for GraphQL query
    const variables = {
      point,
      radius,
      limit,
      facilityType: facilityType || null
    };
    
    // Execute spatial query to find facilities within radius
    const result = await queryNearbyFacilities(variables);
    
    // Transform the results into the expected output format
    const nearbyFacilities = transformFacilitiesData(result, location);
    
    return { data: nearbyFacilities };
  } catch (error) {
    console.error('Error finding nearby facilities:', error);
    throw new Error(`Failed to find nearby facilities: ${error instanceof Error ? error.message : String(error)}`);
  }
}

/**
 * Query the database for facilities within the specified radius
 */
async function queryNearbyFacilities(variables: {
  point: string;
  radius: number;
  limit: number;
  facilityType: string | null;
}): Promise<any[]> {
  try {
    const query = `
      query FindNearbyFacilities(
        $point: geometry!,
        $radius: Float!,
        $limit: Int!,
        $facilityType: String
      ) {
        healthcare_facilities(
          where: {
            _and: [
              { is_active: { _eq: true } },
              { 
                location: { 
                  _st_d_within: { 
                    distance: $radius, 
                    from: $point 
                  } 
                } 
              },
              { 
                facility_type: { 
                  _eq: $facilityType 
                }  
              }
            ]
          },
          limit: $limit,
          order_by: {
            location: {
              _st_distance: {
                from: $point,
                use_spheroid: true
              }
            }
          }
        ) {
          id
          name
          facility_type
          location
          address_line1
          address_line2
          contact_phone
          operating_hours
          district
          # Calculate distance using PostGIS
          distance: location_st_distance(
            args: {
              from: $point,
              use_spheroid: true
            }
          )
          available_services
        }
      }
    `;
    
    // Conditionally apply facility type filter
    if (!variables.facilityType) {
      // Remove the facility_type filter if not provided
      query.replace(
        `{ facility_type: { _eq: $facilityType } }`,
        ''
      );
    }
    
    const result = await hasuraGraphqlClient.request(query, variables);
    return (result as { healthcare_facilities?: any[] })?.healthcare_facilities || [];
  } catch (error) {
    console.error('Error querying facilities:', error);
    return [];
  }
}

/**
 * Transform raw database results into the expected output format
 */
function transformFacilitiesData(
  facilities: any[], 
  _userLocation: { latitude: number; longitude: number }
): NearbyFacilityOutput[] {
  return facilities.map(facility => {
    // Distance is returned in meters from PostGIS, convert to km
    const distanceKm = facility.distance / 1000;
    
    // Estimate travel time (rough approximation: ~30 km/h average speed in Sierra Leone)
    const estimatedSpeedKmPerHour = 30;
    const travelTimeMinutes = Math.round((distanceKm / estimatedSpeedKmPerHour) * 60);
    
    // Format address
    const address = [facility.address_line1, facility.address_line2]
      .filter(Boolean)
      .join(', ');
    
    // Parse PostGIS location to our Geographic type format
    const location: Geographic = facility.location;
    
    return {
      id: facility.id,
      name: facility.name,
      facilityType: facility.facility_type,
      distance: Math.round(distanceKm * 10) / 10, // Round to 1 decimal place
      travelTimeMinutes,
      location,
      address,
      contactPhone: facility.contact_phone,
      operatingHours: facility.operating_hours,
      availableServices: facility.available_services || []
    };
  });
}

// Haversine formula removed as it's not used in this implementation
// PostGIS handles all distance calculations

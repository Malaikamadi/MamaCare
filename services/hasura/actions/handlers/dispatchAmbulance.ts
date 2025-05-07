import { 
  ActionRequest, 
  ActionResponse, 
  DispatchAmbulanceInput,
  AmbulanceDispatchOutput,
  AmbulanceStatus,
  UUID,
  Geographic
} from '../types';
import {
  HasuraSosEvent,
  HasuraAmbulance,
  GetSosEventResult,
  FindAmbulancesResult
} from '../types/hasuraResponses';
import * as z from 'zod';
import { hasuraGraphqlClient } from '../utils/hasuraClient';

// Validation schema using Zod
const inputSchema = z.object({
  sosEventId: z.string().uuid(),
  facilityId: z.string().uuid().optional()
});

/**
 * Dispatch the nearest available ambulance to respond to an SOS event
 * 
 * @param req The action request containing SOS event ID and optional facility ID
 * @returns Dispatch response with ambulance and status information
 */
export default async function dispatchAmbulance(
  req: ActionRequest<DispatchAmbulanceInput>
): Promise<ActionResponse<AmbulanceDispatchOutput>> {
  try {
    // Validate input with Zod schema
    const validatedInput = inputSchema.parse(req.input);
    const { sosEventId, facilityId } = validatedInput;
    
    // Check if the SOS event exists and isn't already assigned an ambulance
    const sosEvent = await getSosEvent(sosEventId);
    
    if (!sosEvent) {
      return {
        data: {
          success: false,
          sosEventId,
          message: 'SOS event not found',
          status: AmbulanceStatus.FAILED
        }
      };
    }
    
    if (sosEvent.ambulance_dispatched) {
      return {
        data: {
          success: false,
          sosEventId,
          message: 'Ambulance already dispatched for this SOS event',
          status: AmbulanceStatus.DISPATCHED
        }
      };
    }
    
    // Find the nearest available ambulance
    let ambulance;
    
    if (facilityId) {
      // If facility is specified, find ambulance from that facility
      ambulance = await findAvailableAmbulanceFromFacility(facilityId, sosEvent.location);
    }
    
    // If no facility specified or no ambulance available at the specified facility,
    // find the nearest ambulance from any facility
    if (!ambulance) {
      ambulance = await findNearestAvailableAmbulance(sosEvent.location);
    }
    
    // If no ambulances are available
    if (!ambulance) {
      return {
        data: {
          success: false,
          sosEventId,
          message: 'No available ambulances found',
          status: AmbulanceStatus.FAILED
        }
      };
    }
    
    // Calculate estimated arrival time based on distance
    const distanceInKm = ambulance.distance || 0;
    const averageSpeedKmPerHour = 30; // Average speed in Sierra Leone roads
    const estimatedArrivalMinutes = Math.round((distanceInKm / averageSpeedKmPerHour) * 60);
    
    // Update the ambulance status
    await updateAmbulanceStatus(
      ambulance.id, 
      AmbulanceStatus.DISPATCHED, 
      sosEvent.location
    );
    
    // Update the SOS event with ambulance information
    await updateSosEvent(
      sosEventId, 
      ambulance.id, 
      ambulance.facility_id
    );
    
    // Log the dispatch in the system
    await logAmbulanceDispatch(
      sosEventId, 
      ambulance.id, 
      estimatedArrivalMinutes
    );
    
    // Prepare the success response
    return {
      data: {
        success: true,
        ambulanceId: ambulance.id,
        estimatedArrivalMinutes,
        dispatchTime: new Date().toISOString(),
        sosEventId,
        message: `Ambulance ${ambulance.id} dispatched successfully`,
        status: AmbulanceStatus.DISPATCHED
      }
    };
  } catch (error) {
    console.error('Error dispatching ambulance:', error);
    throw new Error(`Failed to dispatch ambulance: ${error instanceof Error ? error.message : String(error)}`);
  }
}

/**
 * Get SOS event details from the database
 */
async function getSosEvent(sosEventId: UUID): Promise<HasuraSosEvent | null> {
  try {
    const query = `
      query GetSosEvent($sosEventId: uuid!) {
        sos_events_by_pk(id: $sosEventId) {
          id
          status
          location
          ambulance_dispatched
          responding_facility_id
          responding_user_id
          emergency_type
          description
        }
      }
    `;
    
    const result = await hasuraGraphqlClient.request<GetSosEventResult>(query, { sosEventId });
    return result.sos_events_by_pk || null;
  } catch (error) {
    console.error('Error fetching SOS event:', error);
    throw new Error(`Failed to fetch SOS event: ${error instanceof Error ? error.message : String(error)}`);
  }
}

/**
 * Find an available ambulance from a specific facility
 */
async function findAvailableAmbulanceFromFacility(
  facilityId: UUID, 
  emergencyLocation: Geographic
): Promise<HasuraAmbulance | null> {
  try {
    const query = `
      query FindAmbulanceFromFacility($facilityId: uuid!, $point: geometry!) {
        ambulances(
          where: {
            facility_id: { _eq: $facilityId },
            current_status: { _eq: "AVAILABLE" }
          },
          order_by: {
            current_location: {
              _st_distance: {
                from: $point,
                use_spheroid: true
              }
            }
          },
          limit: 1
        ) {
          id
          facility_id
          current_location
          driver_name
          driver_contact
          license_plate
          distance: current_location_st_distance(
            args: {
              from: $point,
              use_spheroid: true
            }
          )
        }
      }
    `;
    
    const variables = {
      facilityId,
      point: emergencyLocation
    };
    
    const result = await hasuraGraphqlClient.request<FindAmbulancesResult>(query, variables);
    return result.ambulances?.[0] || null;
  } catch (error) {
    console.error('Error finding ambulance from facility:', error);
    return null;
  }
}

/**
 * Find the nearest available ambulance from any facility
 */
async function findNearestAvailableAmbulance(emergencyLocation: Geographic): Promise<HasuraAmbulance | null> {
  try {
    const query = `
      query FindNearestAmbulance($point: geometry!) {
        ambulances(
          where: {
            current_status: { _eq: "AVAILABLE" }
          },
          order_by: {
            current_location: {
              _st_distance: {
                from: $point,
                use_spheroid: true
              }
            }
          },
          limit: 1
        ) {
          id
          facility_id
          current_location
          driver_name
          driver_contact
          license_plate
          distance: current_location_st_distance(
            args: {
              from: $point,
              use_spheroid: true
            }
          )
        }
      }
    `;
    
    const variables = {
      point: emergencyLocation
    };
    
    const result = await hasuraGraphqlClient.request<FindAmbulancesResult>(query, variables);
    const ambulance = result.ambulances?.[0];
    
    if (ambulance) {
      // Convert distance from meters to kilometers
      const distanceMeters = ambulance.distance || 0;
      ambulance.distance = distanceMeters / 1000;
    }
    
    return ambulance || null;
  } catch (error) {
    console.error('Error finding nearest ambulance:', error);
    return null;
  }
}

/**
 * Update ambulance status and location
 */
async function updateAmbulanceStatus(
  ambulanceId: UUID, 
  status: AmbulanceStatus,
  _destination: Geographic
): Promise<void> {
  try {
    const mutation = `
      mutation UpdateAmbulanceStatus(
        $ambulanceId: uuid!, 
        $status: String!,
        $now: timestamptz!
      ) {
        update_ambulances_by_pk(
          pk_columns: { id: $ambulanceId }, 
          _set: { 
            current_status: $status,
            last_location_update: $now
          }
        ) {
          id
        }
      }
    `;
    
    await hasuraGraphqlClient.request(mutation, {
      ambulanceId,
      status,
      now: new Date().toISOString()
    });
  } catch (error) {
    console.error('Error updating ambulance status:', error);
    throw new Error(`Failed to update ambulance status: ${error instanceof Error ? error.message : String(error)}`);
  }
}

/**
 * Update SOS event with ambulance and facility information
 */
async function updateSosEvent(
  sosEventId: UUID, 
  ambulanceId: UUID,
  facilityId: UUID
): Promise<void> {
  try {
    const mutation = `
      mutation UpdateSosEvent(
        $sosEventId: uuid!, 
        $status: String!,
        $ambulanceDispatched: Boolean!,
        $respondingFacilityId: uuid!,
        $assignedAmbulanceId: uuid!
      ) {
        update_sos_events_by_pk(
          pk_columns: { id: $sosEventId }, 
          _set: { 
            status: $status,
            ambulance_dispatched: $ambulanceDispatched,
            responding_facility_id: $respondingFacilityId,
            assigned_ambulance_id: $assignedAmbulanceId
          }
        ) {
          id
        }
      }
    `;
    
    await hasuraGraphqlClient.request(mutation, {
      sosEventId,
      status: 'IN_PROGRESS',
      ambulanceDispatched: true,
      respondingFacilityId: facilityId,
      assignedAmbulanceId: ambulanceId
    });
  } catch (error) {
    console.error('Error updating SOS event:', error);
    throw new Error(`Failed to update SOS event: ${error instanceof Error ? error.message : String(error)}`);
  }
}

/**
 * Log the ambulance dispatch in the system
 */
async function logAmbulanceDispatch(
  sosEventId: UUID, 
  ambulanceId: UUID,
  estimatedArrivalMinutes: number
): Promise<void> {
  try {
    const mutation = `
      mutation LogAmbulanceDispatch(
        $sosEventId: uuid!, 
        $ambulanceId: uuid!,
        $dispatchTime: timestamptz!,
        $estimatedArrivalMinutes: Int!
      ) {
        insert_ambulance_dispatch_logs_one(
          object: {
            sos_event_id: $sosEventId,
            ambulance_id: $ambulanceId,
            dispatch_time: $dispatchTime,
            estimated_arrival_minutes: $estimatedArrivalMinutes,
            status: "DISPATCHED"
          }
        ) {
          id
        }
      }
    `;
    
    await hasuraGraphqlClient.request(mutation, {
      sosEventId,
      ambulanceId,
      dispatchTime: new Date().toISOString(),
      estimatedArrivalMinutes
    });
  } catch (error) {
    console.error('Error logging ambulance dispatch:', error);
    // Non-critical operation, continue even if logging fails
  }
}

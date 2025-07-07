import { 
  ActionRequest, 
  ActionResponse, 
  AppointmentScheduleInput, 
  ScheduledAppointmentOutput,
  TimeRange,
  UUID
} from '../types';
import {
  HasuraFacility,
  HasuraVisit,
  GetFacilityResult,
  GetVisitsResult,
  CheckMotherResult,
  CheckChildResult
} from '../types/hasuraResponses';
import * as z from 'zod';
import { hasuraGraphqlClient } from '../utils/hasuraClient';

// Validation schema for appointment scheduling input
const timeRangeSchema = z.object({
  startTime: z.string().regex(/^([01]\d|2[0-3]):([0-5]\d)$/),
  endTime: z.string().regex(/^([01]\d|2[0-3]):([0-5]\d)$/)
});

// Define the type for the input schema
type AppointmentInputType = {
  motherId?: string;
  childId?: string;
  visitType: string;
  facilityId: string;
  preferredDate?: string;
  preferredTimeRange?: {
    startTime: string;
    endTime: string;
  };
  notes?: string;
  isUrgent?: boolean;
};

const inputSchema = z.object({
  motherId: z.string().uuid().optional(),
  childId: z.string().uuid().optional(),
  visitType: z.string().min(1),
  facilityId: z.string().uuid(),
  preferredDate: z.string().optional(),
  preferredTimeRange: timeRangeSchema.optional(),
  notes: z.string().optional(),
  isUrgent: z.boolean().optional()
}).refine((data: AppointmentInputType) => Boolean(data.motherId || data.childId), {
  message: "Either motherId or childId must be provided"
});

// Constants for availability slots
const FACILITY_OPENING_HOUR = 8; // 8 AM
const FACILITY_CLOSING_HOUR = 17; // 5 PM

/**
 * Schedule an appointment with conflict checking and optimal time selection
 * 
 * @param req The action request containing appointment details
 * @returns Scheduled appointment information
 */
export default async function scheduleAppointment(
  req: ActionRequest<AppointmentScheduleInput>
): Promise<ActionResponse<ScheduledAppointmentOutput>> {
  try {
    // Validate input with Zod schema
    const validatedInput = inputSchema.parse(req.input);
    const { 
      motherId, 
      childId, 
      visitType, 
      facilityId, 
      preferredDate,
      preferredTimeRange,
      notes,
      isUrgent = false
    } = validatedInput;
    
    // Get facility details to verify operating hours
    const facility = await getFacilityDetails(facilityId);
    if (!facility) {
      throw new Error(`Facility with ID ${facilityId} not found`);
    }
    
    // If mother ID is provided, verify it exists
    if (motherId) {
      const motherExists = await checkMotherExists(motherId);
      if (!motherExists) {
        throw new Error(`Mother with ID ${motherId} not found`);
      }
    }
    
    // If child ID is provided, verify it exists and is related to mother if both are provided
    if (childId) {
      const childExists = await checkChildExists(childId, motherId);
      if (!childExists) {
        throw new Error(`Child with ID ${childId} not found${motherId ? ` or not related to mother ${motherId}` : ''}`);
      }
    }
    
    // Determine appointment date - either preferred or next available
    const appointmentDate = preferredDate || await findNextAvailableDate(facilityId, isUrgent);
    
    // Check facility availability for the appointment date
    const availableSlots = await getFacilityAvailability(
      facilityId, 
      appointmentDate, 
      visitType
    );
    
    if (availableSlots.length === 0) {
      // No slots available on preferred date
      const alternativeDates = await findAlternativeDates(facilityId, appointmentDate, 5);
      
      return {
        data: {
          visitId: '', // Empty as no appointment was created
          scheduledDate: '',
          scheduledTime: '',
          visitType,
          facilityId,
          facilityName: facility.name,
          motherId,
          childId,
          reminderSent: false,
          conflictResolved: false,
          suggestedAlternatives: alternativeDates
        }
      };
    }
    
    // If preferredTimeRange is provided, find the best matching slot
    let selectedSlot;
    if (preferredTimeRange) {
      selectedSlot = findBestTimeSlot(availableSlots, preferredTimeRange);
    } else {
      // Otherwise, take the first available slot
      selectedSlot = availableSlots[0];
    }
    
    // Create the appointment
    const visit = await createVisit({
      motherId,
      childId,
      facilityId,
      visitType,
      scheduledDate: appointmentDate,
      scheduledTime: selectedSlot,
      notes
    });
    
    // Handle case where visit creation failed
    if (!visit) {
      throw new Error('Failed to create visit record');
    }
    
    // Prepare the response
    return {
      data: {
        visitId: visit.id,
        scheduledDate: appointmentDate,
        scheduledTime: selectedSlot,
        visitType,
        facilityId,
        facilityName: facility.name,
        motherId,
        childId,
        reminderSent: false,
        conflictResolved: Boolean(preferredDate !== appointmentDate || 
          (preferredTimeRange && selectedSlot !== `${preferredTimeRange.startTime}-${preferredTimeRange.endTime}`)),
        suggestedAlternatives: []
      }
    };
  } catch (error) {
    console.error('Error scheduling appointment:', error);
    throw new Error(`Failed to schedule appointment: ${error instanceof Error ? error.message : String(error)}`);
  }
}

/**
 * Get facility details including operating hours
 */
async function getFacilityDetails(facilityId: UUID): Promise<HasuraFacility | null> {
  try {
    const query = `
      query GetFacilityDetails($facilityId: uuid!) {
        healthcare_facilities_by_pk(id: $facilityId) {
          id
          name
          operating_hours
          is_active
        }
      }
    `;
    
    const result = await hasuraGraphqlClient.request<GetFacilityResult>(query, { facilityId });
    return result.healthcare_facilities_by_pk || null;
  } catch (error) {
    console.error('Error fetching facility details:', error);
    throw new Error(`Failed to fetch facility details: ${error instanceof Error ? error.message : String(error)}`);
  }
}

/**
 * Check if a mother exists in the database
 */
async function checkMotherExists(motherId: UUID): Promise<boolean> {
  try {
    const query = `
      query CheckMotherExists($motherId: uuid!) {
        users_by_pk(id: $motherId) {
          id
          role
        }
      }
    `;
    
    const result = await hasuraGraphqlClient.request<CheckMotherResult>(query, { motherId });
    const user = result.users_by_pk;
    return !!user && user.role === 'mother';
  } catch (error) {
    console.error('Error checking mother existence:', error);
    return false;
  }
}

/**
 * Check if a child exists and is related to the given mother (if provided)
 */
async function checkChildExists(childId: UUID, motherId?: UUID): Promise<boolean> {
  try {
    const query = `
      query CheckChildExists($childId: uuid!, $motherId: uuid) {
        children_by_pk(id: $childId) {
          id
          mother_id
        }
      }
    `;
    
    const result = await hasuraGraphqlClient.request<CheckChildResult>(query, { childId });
    const child = result.children_by_pk;
    
    if (!child) {
      return false;
    }
    
    // If motherId is provided, check if this child belongs to that mother
    if (motherId && child.mother_id !== motherId) {
      return false;
    }
    
    return true;
  } catch (error) {
    console.error('Error checking child existence:', error);
    return false;
  }
}

/**
 * Find the next available date for an appointment at the facility
 */
async function findNextAvailableDate(facilityId: UUID, isUrgent: boolean): Promise<string> {
  // Start with tomorrow's date for regular appointments
  // For urgent appointments, start with today
  const startDate = new Date();
  if (!isUrgent) {
    startDate.setDate(startDate.getDate() + 1);
  }
  
  // Format date as YYYY-MM-DD
  const formatDate = (date: Date): string => {
    return date.toISOString().split('T')[0];
  };
  
  // Try the next 10 days to find an available date
  for (let i = 0; i < 10; i++) {
    const checkDate = new Date(startDate);
    checkDate.setDate(checkDate.getDate() + i);
    
    // Skip weekends (0 = Sunday, 6 = Saturday)
    const dayOfWeek = checkDate.getDay();
    if (dayOfWeek === 0 || dayOfWeek === 6) {
      continue;
    }
    
    const dateStr = formatDate(checkDate);
    
    // Check if the facility has any available slots on this date
    const hasAvailability = await checkDateAvailability(facilityId, dateStr);
    
    if (hasAvailability) {
      return dateStr;
    }
  }
  
  // If no availability found in the next 10 days, return the first non-weekend day
  const firstAvailableDate = new Date(startDate);
  while (true) {
    const dayOfWeek = firstAvailableDate.getDay();
    if (dayOfWeek !== 0 && dayOfWeek !== 6) {
      break;
    }
    firstAvailableDate.setDate(firstAvailableDate.getDate() + 1);
  }
  
  return formatDate(firstAvailableDate);
}

/**
 * Check if a specific date has available slots at the facility
 */
async function checkDateAvailability(facilityId: UUID, date: string): Promise<boolean> {
  try {
    const query = `
      query CheckDateAvailability($facilityId: uuid!, $date: date!) {
        visits_aggregate(
          where: {
            facility_id: { _eq: $facilityId },
            scheduled_date: { _eq: $date },
            status: { _in: ["SCHEDULED", "CONFIRMED"] }
          }
        ) {
          aggregate {
            count
          }
        }
      }
    `;
    
    const result = await hasuraGraphqlClient.request<GetVisitsResult>(query, { facilityId, date });
    const count = result.visits_aggregate?.aggregate?.count || 0;
    
    // Using a constant to avoid magic numbers
    const MAX_DAILY_APPOINTMENTS = 20;
    
    // Assuming a facility can handle 20 appointments per day
    // This is a simplified model - in reality, we would check per timeslot
    return count < MAX_DAILY_APPOINTMENTS;
  } catch (error) {
    console.error('Error checking date availability:', error);
    return false;
  }
}

/**
 * Find alternative dates when the preferred date is not available
 */
async function findAlternativeDates(
  facilityId: UUID, 
  afterDate: string, 
  limit: number
): Promise<string[]> {
  const alternatives: string[] = [];
  const startDate = new Date(afterDate);
  startDate.setDate(startDate.getDate() + 1);
  
  const formatDate = (date: Date): string => {
    return date.toISOString().split('T')[0];
  };
  
  // Check the next 14 days for available dates
  for (let i = 0; i < 14 && alternatives.length < limit; i++) {
    const checkDate = new Date(startDate);
    checkDate.setDate(checkDate.getDate() + i);
    
    // Skip weekends
    const dayOfWeek = checkDate.getDay();
    if (dayOfWeek === 0 || dayOfWeek === 6) {
      continue;
    }
    
    const dateStr = formatDate(checkDate);
    const hasAvailability = await checkDateAvailability(facilityId, dateStr);
    
    if (hasAvailability) {
      alternatives.push(dateStr);
    }
  }
  
  return alternatives;
}

/**
 * Get available time slots for a specific date at the facility
 */
async function getFacilityAvailability(
  facilityId: UUID, 
  date: string, 
  _visitType: string
): Promise<string[]> {
  try {
    // Fetch existing appointments for the specified date
    const query = `
      query GetExistingAppointments($facilityId: uuid!, $date: date!) {
        visits(
          where: {
            facility_id: { _eq: $facilityId },
            scheduled_date: { _eq: $date },
            status: { _in: ["SCHEDULED", "CONFIRMED"] }
          }
        ) {
          id
          scheduled_time
          visit_type
          expected_duration_minutes
        }
      }
    `;
    
    const result = await hasuraGraphqlClient.request<GetVisitsResult>(query, { facilityId, date });
    const existingAppointments = result.visits || [];
    
    // Generate all possible time slots for the day
    const allTimeSlots: string[] = generateTimeSlots();
    
    // Filter out time slots that are already booked
    const bookedSlots = new Set<string>();
    existingAppointments.forEach((appt: HasuraVisit) => {
      if (appt.scheduled_time) {
        bookedSlots.add(appt.scheduled_time);
      }
    });
    
    // Return available slots
    return allTimeSlots.filter(slot => !bookedSlots.has(slot));
  } catch (error) {
    console.error('Error getting facility availability:', error);
    return [];
  }
}

/**
 * Generate all possible time slots for a day
 */
function generateTimeSlots(): string[] {
  const slots: string[] = [];
  
  // Generate 30-minute slots from opening to closing time
  for (let hour = FACILITY_OPENING_HOUR; hour < FACILITY_CLOSING_HOUR; hour++) {
    for (let minute = 0; minute < 60; minute += 30) {
      const startHourStr = hour.toString().padStart(2, '0');
      const startMinuteStr = minute.toString().padStart(2, '0');
      
      let endHour = hour;
      let endMinute = minute + 30;
      if (endMinute >= 60) {
        endHour += 1;
        endMinute -= 60;
      }
      
      const endHourStr = endHour.toString().padStart(2, '0');
      const endMinuteStr = endMinute.toString().padStart(2, '0');
      
      // Skip if end time would be after closing
      if (endHour >= FACILITY_CLOSING_HOUR) {
        continue;
      }
      
      slots.push(`${startHourStr}:${startMinuteStr}-${endHourStr}:${endMinuteStr}`);
    }
  }
  
  return slots;
}

/**
 * Find the best matching time slot based on preferred time range
 */
function findBestTimeSlot(availableSlots: string[], preferredRange: TimeRange): string {
  // Convert time strings to minutes for easier comparison
  const convertTimeToMinutes = (timeStr: string): number => {
    const [hours, minutes] = timeStr.split(':').map(Number);
    return hours * 60 + minutes;
  };
  
  const preferredStartMinutes = convertTimeToMinutes(preferredRange.startTime);
  const preferredEndMinutes = convertTimeToMinutes(preferredRange.endTime);
  
  // Score each available slot based on how close it is to the preferred range
  let bestScore = Number.MAX_SAFE_INTEGER;
  let bestSlot = availableSlots[0]; // Default to first available
  
  availableSlots.forEach(slot => {
    const [slotStart, slotEnd] = slot.split('-');
    const slotStartMinutes = convertTimeToMinutes(slotStart);
    const slotEndMinutes = convertTimeToMinutes(slotEnd);
    
    // Calculate score as distance from preferred time
    const startDiff = Math.abs(slotStartMinutes - preferredStartMinutes);
    const endDiff = Math.abs(slotEndMinutes - preferredEndMinutes);
    const score = startDiff + endDiff;
    
    // Update best slot if this one is closer to preferred time
    if (score < bestScore) {
      bestScore = score;
      bestSlot = slot;
    }
  });
  
  return bestSlot;
}

/**
 * Create a new visit record in the database
 */
async function createVisit(visitData: {
  motherId?: UUID;
  childId?: UUID;
  facilityId: UUID;
  visitType: string;
  scheduledDate: string;
  scheduledTime: string;
  notes?: string;
}): Promise<HasuraVisit | null> {
  try {
    const mutation = `
      mutation CreateVisit(
        $motherId: uuid,
        $childId: uuid,
        $facilityId: uuid!,
        $visitType: String!,
        $scheduledDate: date!,
        $scheduledTime: String!,
        $notes: String
      ) {
        insert_visits_one(
          object: {
            mother_id: $motherId,
            child_id: $childId,
            facility_id: $facilityId,
            visit_type: $visitType,
            scheduled_date: $scheduledDate,
            scheduled_time: $scheduledTime,
            status: "SCHEDULED",
            reminder_sent: false,
            notes: $notes
          }
        ) {
          id
          mother_id
          child_id
          scheduled_date
          scheduled_time
        }
      }
    `;
    
    interface CreateVisitResult {
      insert_visits_one?: HasuraVisit;
    }
    
    const result = await hasuraGraphqlClient.request<CreateVisitResult>(mutation, visitData);
    return result.insert_visits_one || null;
  } catch (error) {
    console.error('Error creating visit:', error);
    throw new Error(`Failed to create visit: ${error instanceof Error ? error.message : String(error)}`);
  }
}

// Type definitions for Hasura GraphQL responses
// Following strict TypeScript principles with zero-any policy

import { Geographic } from './index';

// Use string type for IDs to avoid circular imports
export type UUID = string;

// Common interfaces for database entities
export interface HasuraUser {
  id: string;
  first_name: string;
  last_name: string;
  role: string;
  email?: string;
  date_of_birth?: string;
  expected_delivery_date?: string;
  is_high_risk?: boolean;
  chw_id?: string;
  facility_id?: string;
  medical_history?: string;
}

export interface HasuraFacility {
  id: string;
  name: string;
  facility_type: string;
  location: Geographic;
  address_line1?: string;
  address_line2?: string;
  contact_phone?: string;
  contact_email?: string;
  operating_hours?: string;
  district?: string;
}

export interface HasuraAmbulance {
  id: string;
  facility_id: string;
  vehicle_type: string;
  current_status: string;
  current_location: Geographic;
  driver_name?: string;
  driver_contact?: string;
  license_plate?: string;
  distance?: number;
}

export interface HasuraSosEvent {
  id: string;
  user_id: string;
  emergency_type: string;
  status: string;
  location: Geographic;
  location_address?: string;
  location_description?: string;
  responding_facility_id?: string;
  ambulance_dispatched?: boolean;
  assigned_ambulance_id?: string;
  description?: string;
}

export interface HasuraVisit {
  id: string;
  mother_id?: string;
  child_id?: string;
  visit_type: string;
  status: string;
  scheduled_date: string;
  scheduled_time?: string;
  facility_id: string;
  reminder_sent: boolean;
  notes?: string;
}

// Query result interfaces
export interface GetMotherDataResult {
  users_by_pk?: HasuraUser;
}

export interface GetSosEventResult {
  sos_events_by_pk?: HasuraSosEvent;
}

export interface FindAmbulancesResult {
  ambulances?: HasuraAmbulance[];
}

export interface GetFacilityResult {
  healthcare_facilities_by_pk?: HasuraFacility;
}

export interface FindFacilitiesResult {
  healthcare_facilities?: HasuraFacility[];
}

export interface GetVisitsResult {
  visits?: HasuraVisit[];
  visits_aggregate?: {
    aggregate?: {
      count?: number;
    }
  }
}

export interface CheckChildResult {
  children_by_pk?: {
    id: string;
    mother_id?: string;
  } | null;
}

export interface CheckMotherResult {
  users_by_pk?: {
    id: string;
    role: string;
  } | null;
}
// Type definitions for Hasura Action handlers
// Following TypeScript principles with strict null checks, no implicit any

// Common types
export type UUID = string;

// Geographic location type
export interface Geographic {
  type: 'Point';
  coordinates: [number, number]; // [longitude, latitude]
}

// Base Request interface for all actions
export interface ActionRequest<T> {
  action: {
    name: string;
  };
  input: T;
  session_variables: {
    'x-hasura-role': string;
    'x-hasura-user-id'?: string;
    'x-hasura-facility-id'?: string;
  };
}

// Base Response interface for all actions
export interface ActionResponse<T> {
  data: T;
}

// ===============================
// Calculate Pregnancy Risk Action
// ===============================

export interface PregnancyHealthFactors {
  age: number;
  gestationalAgeWeeks: number;
  bloodPressureSystolic?: number;
  bloodPressureDiastolic?: number;
  bloodGlucose?: number;
  previousPregnancyComplications?: string[];
  existingConditions?: string[];
  hivStatus?: string;
  height?: number;
  weight?: number;
}

export enum RiskLevel {
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  CRITICAL = 'CRITICAL',
}

export interface RiskFactorDetail {
  factor: string;
  severity: number;
  description: string;
  mitigationStrategy?: string;
}

export interface RiskAssessmentOutput {
  riskScore: number;
  riskLevel: RiskLevel;
  riskFactors: RiskFactorDetail[];
  recommendations: string[];
  suggestedFollowUpDays?: number;
  suggestedSpecialist?: string;
}

export interface CalculatePregnancyRiskInput {
  motherId: UUID;
  healthFactors: PregnancyHealthFactors;
}

// ===============================
// Find Nearby Facilities Action
// ===============================

export interface GeographicPoint {
  latitude: number;
  longitude: number;
}

export interface NearbyFacilityOutput {
  id: UUID;
  name: string;
  facilityType: string;
  distance: number;
  travelTimeMinutes?: number;
  location: Geographic;
  address?: string;
  contactPhone?: string;
  operatingHours?: string;
  availableServices?: string[];
}

export interface FindNearbyFacilitiesInput {
  location: GeographicPoint;
  radius: number;
  facilityType?: string;
  limit?: number;
}

// ===============================
// Dispatch Ambulance Action
// ===============================

export enum AmbulanceStatus {
  DISPATCHED = 'DISPATCHED',
  EN_ROUTE = 'EN_ROUTE',
  ARRIVED = 'ARRIVED',
  RETURNING = 'RETURNING',
  COMPLETED = 'COMPLETED',
  FAILED = 'FAILED',
}

export interface AmbulanceDispatchOutput {
  success: boolean;
  ambulanceId?: UUID;
  estimatedArrivalMinutes?: number;
  dispatchTime?: string;
  sosEventId: UUID;
  message?: string;
  status?: AmbulanceStatus;
}

export interface DispatchAmbulanceInput {
  sosEventId: UUID;
  facilityId?: UUID;
}

// ===============================
// Generate Health Report Action
// ===============================

export enum ReportType {
  MATERNAL_HEALTH_SUMMARY = 'MATERNAL_HEALTH_SUMMARY',
  CHILD_HEALTH_SUMMARY = 'CHILD_HEALTH_SUMMARY',
  FACILITY_PERFORMANCE = 'FACILITY_PERFORMANCE',
  CHW_PERFORMANCE = 'CHW_PERFORMANCE',
  REGIONAL_STATISTICS = 'REGIONAL_STATISTICS',
  IMMUNIZATION_COVERAGE = 'IMMUNIZATION_COVERAGE',
}

export interface ReportMetric {
  name: string;
  value: number;
  unit?: string;
  change?: number;
  changeDirection?: string;
}

export interface ChartDataset {
  label: string;
  data: number[];
  color?: string;
}

export interface ReportChart {
  title: string;
  chartType: string;
  labels: string[];
  datasets: ChartDataset[];
}

export interface HealthReportOutput {
  reportId: string;
  title: string;
  generatedAt: string;
  reportType: ReportType;
  reportUrl?: string;
  metrics: ReportMetric[];
  charts?: ReportChart[];
  recommendations: string[];
}

export interface GenerateHealthReportInput {
  regionId?: string;
  motherId?: UUID;
  startDate: string;
  endDate: string;
  reportType: ReportType;
}

// ===============================
// Schedule Appointment Action
// ===============================

export interface TimeRange {
  startTime: string;
  endTime: string;
}

export interface AppointmentScheduleInput {
  motherId?: UUID;
  childId?: UUID;
  visitType: string;
  facilityId: UUID;
  preferredDate?: string;
  preferredTimeRange?: TimeRange;
  notes?: string;
  isUrgent?: boolean;
}

export interface ScheduledAppointmentOutput {
  visitId: UUID;
  scheduledDate: string;
  scheduledTime: string;
  visitType: string;
  facilityId: UUID;
  facilityName: string;
  motherId?: UUID;
  childId?: UUID;
  reminderSent: boolean;
  conflictResolved: boolean;
  suggestedAlternatives?: string[];
}

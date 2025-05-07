import { 
  ActionRequest, 
  ActionResponse, 
  CalculatePregnancyRiskInput, 
  RiskAssessmentOutput, 
  RiskFactorDetail, 
  RiskLevel 
} from '../types';
import * as z from 'zod';
import { hasuraGraphqlClient } from '../utils/hasuraClient';

// Constants to avoid magic numbers
const RISK_THRESHOLDS = {
  LOW: 0.25,
  MEDIUM: 0.50,
  HIGH: 0.75
};

// Validation schema for inputs
const inputSchema = z.object({
  motherId: z.string().uuid(),
  healthFactors: z.object({
    age: z.number().int().positive(),
    gestationalAgeWeeks: z.number().int().min(1).max(45),
    bloodPressureSystolic: z.number().int().positive().optional(),
    bloodPressureDiastolic: z.number().int().positive().optional(),
    bloodGlucose: z.number().positive().optional(),
    previousPregnancyComplications: z.array(z.string()).optional(),
    existingConditions: z.array(z.string()).optional(),
    hivStatus: z.string().optional(),
    height: z.number().positive().optional(),
    weight: z.number().positive().optional(),
  }),
});

/**
 * Calculate pregnancy risk score based on maternal health factors
 * 
 * @param req The action request
 * @returns Risk assessment including score, level, factors, and recommendations
 */
export default async function calculatePregnancyRisk(
  req: ActionRequest<CalculatePregnancyRiskInput>
): Promise<ActionResponse<RiskAssessmentOutput>> {
  try {
    // Validate input with Zod schema
    const validatedInput = inputSchema.parse(req.input);
    const { motherId, healthFactors } = validatedInput;

    // Fetch additional mother data from Hasura if needed
    const motherData = await getMotherData(motherId);
    
    // Calculate risk score based on various factors
    const riskFactors = calculateRiskFactors(healthFactors, motherData);
    
    // Calculate overall risk score by averaging factor severities
    const totalSeverity = riskFactors.reduce((sum, factor) => sum + factor.severity, 0);
    const riskScore = riskFactors.length > 0 
      ? totalSeverity / riskFactors.length 
      : 0;
    
    // Determine risk level based on score
    const riskLevel = determineRiskLevel(riskScore);
    
    // Generate recommendations based on risk factors
    const recommendations = generateRecommendations(riskFactors, riskLevel);
    
    // Determine follow-up timeline based on risk level
    const suggestedFollowUpDays = getSuggestedFollowUpDays(riskLevel);
    
    // Suggest specialist if needed
    const suggestedSpecialist = getSuggestedSpecialist(riskFactors);
    
    // Prepare the response
    const response: RiskAssessmentOutput = {
      riskScore,
      riskLevel,
      riskFactors,
      recommendations,
      suggestedFollowUpDays,
      suggestedSpecialist
    };
    
    // Update the mother's risk status in the database
    await updateMotherRiskStatus(motherId, riskLevel);
    
    return { data: response };
  } catch (error) {
    console.error('Error calculating pregnancy risk:', error);
    throw new Error(`Failed to calculate pregnancy risk: ${error instanceof Error ? error.message : String(error)}`);
  }
}

/**
 * Retrieve mother data from the database
 */
async function getMotherData(motherId: string): Promise<Record<string, unknown>> {
  try {
    const query = `
      query GetMotherData($motherId: uuid!) {
        users_by_pk(id: $motherId) {
          id
          date_of_birth
          expected_delivery_date
          previous_pregnancies
          medical_history
        }
      }
    `;
    
    const result = await hasuraGraphqlClient.request(query, { motherId });
    return (result as { users_by_pk?: Record<string, unknown> })?.users_by_pk || {};
  } catch (error) {
    console.error('Error fetching mother data:', error);
    return {};
  }
}

/**
 * Calculate individual risk factors and their severities
 */
function calculateRiskFactors(
  healthFactors: CalculatePregnancyRiskInput['healthFactors'],
  _motherData: Record<string, unknown>
): RiskFactorDetail[] {
  const riskFactors: RiskFactorDetail[] = [];
  
  // Age-related risks (high risk if under 18 or over 35)
  if (healthFactors.age < 18) {
    riskFactors.push({
      factor: 'Adolescent Pregnancy',
      severity: 0.7,
      description: 'Mother is under 18 years old',
      mitigationStrategy: 'More frequent prenatal visits and specialized adolescent care'
    });
  } else if (healthFactors.age > 35) {
    riskFactors.push({
      factor: 'Advanced Maternal Age',
      severity: 0.6,
      description: 'Mother is over 35 years old',
      mitigationStrategy: 'Genetic counseling and more frequent monitoring'
    });
  }
  
  // Blood pressure related risks
  if (healthFactors.bloodPressureSystolic && healthFactors.bloodPressureDiastolic) {
    if (healthFactors.bloodPressureSystolic >= 140 || healthFactors.bloodPressureDiastolic >= 90) {
      riskFactors.push({
        factor: 'Hypertension',
        severity: 0.8,
        description: 'Elevated blood pressure',
        mitigationStrategy: 'Blood pressure medication and dietary changes'
      });
    }
  }
  
  // Blood glucose related risks
  if (healthFactors.bloodGlucose && healthFactors.bloodGlucose > 130) {
    riskFactors.push({
      factor: 'Gestational Diabetes Risk',
      severity: 0.75,
      description: 'Elevated blood glucose levels',
      mitigationStrategy: 'Dietary management and glucose monitoring'
    });
  }
  
  // Previous complications
  if (healthFactors.previousPregnancyComplications && healthFactors.previousPregnancyComplications.length > 0) {
    riskFactors.push({
      factor: 'History of Complications',
      severity: 0.85,
      description: `Previous complications: ${healthFactors.previousPregnancyComplications.join(', ')}`,
      mitigationStrategy: 'Specialized care based on previous complication types'
    });
  }
  
  // HIV status
  if (healthFactors.hivStatus === 'POSITIVE') {
    riskFactors.push({
      factor: 'HIV Positive Status',
      severity: 0.9,
      description: 'Mother is HIV positive',
      mitigationStrategy: 'Antiretroviral therapy and prevention of mother-to-child transmission'
    });
  }
  
  // BMI calculation if height and weight are provided
  if (healthFactors.height && healthFactors.weight) {
    const heightInMeters = healthFactors.height / 100;
    const bmi = healthFactors.weight / (heightInMeters * heightInMeters);
    
    if (bmi < 18.5) {
      riskFactors.push({
        factor: 'Underweight',
        severity: 0.65,
        description: 'BMI below 18.5',
        mitigationStrategy: 'Nutritional counseling and weight gain monitoring'
      });
    } else if (bmi >= 30) {
      riskFactors.push({
        factor: 'Obesity',
        severity: 0.7,
        description: 'BMI 30 or higher',
        mitigationStrategy: 'Dietary guidance and weight management'
      });
    }
  }
  
  // Late gestational age without proper care
  if (healthFactors.gestationalAgeWeeks > 36 && riskFactors.length > 0) {
    riskFactors.push({
      factor: 'Late Pregnancy with Risk Factors',
      severity: 0.8,
      description: 'Third trimester with existing risk factors',
      mitigationStrategy: 'Immediate comprehensive assessment and birth plan'
    });
  }
  
  return riskFactors;
}

/**
 * Determine the overall risk level based on the calculated score
 */
function determineRiskLevel(riskScore: number): RiskLevel {
  if (riskScore < RISK_THRESHOLDS.LOW) {
    return RiskLevel.LOW;
  } else if (riskScore < RISK_THRESHOLDS.MEDIUM) {
    return RiskLevel.MEDIUM;
  } else if (riskScore < RISK_THRESHOLDS.HIGH) {
    return RiskLevel.HIGH;
  } else {
    return RiskLevel.CRITICAL;
  }
}

/**
 * Generate recommendations based on identified risk factors
 */
function generateRecommendations(riskFactors: RiskFactorDetail[], riskLevel: RiskLevel): string[] {
  const recommendations: string[] = [];
  
  // Basic recommendations for all
  recommendations.push('Continue regular prenatal check-ups');
  recommendations.push('Maintain a balanced diet rich in folic acid and iron');
  
  // Add specific recommendations based on risk factors
  riskFactors.forEach(factor => {
    if (factor.mitigationStrategy) {
      recommendations.push(factor.mitigationStrategy);
    }
  });
  
  // Add recommendations based on overall risk level
  switch (riskLevel) {
    case RiskLevel.LOW:
      recommendations.push('Continue normal prenatal care schedule');
      break;
    case RiskLevel.MEDIUM:
      recommendations.push('Increase prenatal visit frequency');
      recommendations.push('Additional screening tests recommended');
      break;
    case RiskLevel.HIGH:
      recommendations.push('Consultation with specialized obstetrician recommended');
      recommendations.push('Create detailed birth plan with medical team');
      recommendations.push('Consider proximity to emergency services');
      break;
    case RiskLevel.CRITICAL:
      recommendations.push('Immediate medical consultation required');
      recommendations.push('Hospital monitoring may be necessary');
      recommendations.push('Prepare emergency transport plan');
      break;
  }
  
  return recommendations;
}

/**
 * Determine suggested follow-up timing based on risk level
 */
function getSuggestedFollowUpDays(riskLevel: RiskLevel): number | undefined {
  switch (riskLevel) {
    case RiskLevel.LOW:
      return 30;
    case RiskLevel.MEDIUM:
      return 14;
    case RiskLevel.HIGH:
      return 7;
    case RiskLevel.CRITICAL:
      return 1;
    default:
      return undefined;
  }
}

/**
 * Suggest specialist type based on risk factors
 */
function getSuggestedSpecialist(riskFactors: RiskFactorDetail[]): string | undefined {
  // Look for specific conditions that require specialists
  for (const factor of riskFactors) {
    if (factor.factor === 'Hypertension' || factor.factor === 'Preeclampsia Risk') {
      return 'Maternal-Fetal Medicine Specialist';
    }
    if (factor.factor === 'Gestational Diabetes Risk') {
      return 'Endocrinologist';
    }
    if (factor.factor === 'HIV Positive Status') {
      return 'Infectious Disease Specialist';
    }
  }
  
  // If multiple high severity factors, suggest maternal-fetal medicine
  const highSeverityFactors = riskFactors.filter(f => f.severity >= 0.8);
  if (highSeverityFactors.length >= 2) {
    return 'Maternal-Fetal Medicine Specialist';
  }
  
  return undefined;
}

/**
 * Update the mother's risk status in the database
 */
async function updateMotherRiskStatus(motherId: string, riskLevel: RiskLevel): Promise<void> {
  try {
    const mutation = `
      mutation UpdateMotherRiskStatus($motherId: uuid!, $isHighRisk: Boolean!) {
        update_users_by_pk(
          pk_columns: { id: $motherId }, 
          _set: { is_high_risk: $isHighRisk }
        ) {
          id
        }
      }
    `;
    
    const isHighRisk = riskLevel === RiskLevel.HIGH || riskLevel === RiskLevel.CRITICAL;
    
    await hasuraGraphqlClient.request(mutation, {
      motherId,
      isHighRisk
    });
  } catch (error) {
    console.error('Error updating mother risk status:', error);
    // Non-critical operation - we continue even if this fails
  }
}
